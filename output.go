package thingiverseio

import (
	"fmt"
	"log"

	"github.com/joernweissenborn/eventual2go"
	"github.com/joernweissenborn/thingiverse.io/config"
	"github.com/joernweissenborn/thingiverse.io/service/connection"
	"github.com/joernweissenborn/thingiverse.io/service/manager"
	"github.com/joernweissenborn/thingiverse.io/service/messages"
)

const (
	listenEvent     = "listen"
	stopListenEvent = "stop_listen"
	peerGoneEvent   = "peer_gone"
	replyEvent      = "reply"
)

type Output struct {
	cfg       *config.Config
	m         *manager.Manager
	listeners map[string]map[config.UUID]interface{}
	logger    *log.Logger
	r         *eventual2go.Reactor
	requests  *messages.RequestStream
}

func NewOutputFromConfig(cfg *config.Config) (o *Output, err error) {
	m, err := manager.New(cfg)
	o = &Output{
		cfg:       cfg,
		m:         m,
		requests:  &messages.RequestStream{m.MessagesOfType(messages.REQUEST).Transform(connection.ToMessage)},
		listeners: map[string]map[config.UUID]interface{}{},
		logger:    log.New(cfg.Logger(), fmt.Sprintf("%s EXPORT", cfg.UUID()), 0),
		r:         eventual2go.NewReactor(),
	}

	o.r.React(listenEvent, o.onListen)
	o.r.AddStream(listenEvent, m.MessagesOfType(messages.LISTEN).Stream)

	o.r.React(stopListenEvent, o.onStopListen)
	o.r.AddStream(stopListenEvent, m.MessagesOfType(messages.STOPLISTEN).Stream)

	o.r.React(peerGoneEvent, o.onPeerGone)

	o.r.React(replyEvent, o.deliverResult)
	return
}

func (o *Output) Reply(r *messages.Request, params interface{}) {
	res := messages.NewResult(o.cfg.UUID(), r, params)
	o.r.Fire(replyEvent, res)
}

func (o *Output) ReplyEncoded(r *messages.Request, params []byte) {
	res := messages.NewEncodedResult(o.cfg.UUID(), r, params)
	o.r.Fire(replyEvent, res)
}

func (o *Output) Emit(function string, inparams interface{}, outparams interface{}) {
	req := messages.NewRequest(o.cfg.UUID(), function, messages.MANY2ONE, inparams)
	o.Reply(req, outparams)
}

func (o *Output) EmitEncoded(function string, inparams []byte, outparams []byte) {
	req := messages.NewEncodedRequest("", function, messages.MANY2ONE, inparams)
	o.ReplyEncoded(req, outparams)
}

func (o *Output) Requests() *messages.RequestStream {
	return o.requests
}

func (o *Output) onListen(d eventual2go.Data) {
	m := d.(connection.Message)
	l := messages.Unflatten(m.Payload)
	f := l.(*messages.Listen).Function
	o.logger.Println("New Listener", m.Sender, f)
	_, ok := o.listeners[f]
	if ok {
		o.listeners[f][m.Sender] = nil
	} else {
		o.listeners[f] = map[config.UUID]interface{}{m.Sender: nil}
	}
	o.r.AddFuture(peerGoneEvent, o.m.PeerLeave(m.Sender).Future)
}

func (o *Output) onStopListen(d eventual2go.Data) {
	m := d.(connection.Message)
	l := messages.Unflatten(m.Payload)
	f := l.(*messages.StopListen).Function
	o.removePeerListen(m.Sender, f)
}

func (o *Output) onPeerGone(d eventual2go.Data) {
	uuid := d.(config.UUID)
	for f, _ := range o.listeners {
		o.removePeerListen(uuid, f)
	}
}

func (o *Output) removePeerListen(uuid config.UUID, f string) {
	_, ok := o.listeners[f]
	if !ok {
		return
	}
	_, ok = o.listeners[f][uuid]
	if ok {
		delete(o.listeners[f], uuid)
	}
}

func (o *Output) deliverResult(d eventual2go.Data) {
	result := d.(*messages.Result)
	o.logger.Println("Delivering result", result.Request.Function, result.Request.CallType)
	switch result.Request.CallType {
	case messages.ONE2MANY, messages.ONE2ONE:
		o.m.SendTo(result.Request.UUID, result)

	case messages.MANY2MANY, messages.MANY2ONE:
		if ls, ok := o.listeners[result.Request.Function]; ok {
			for uuid := range ls {
				o.m.SendTo(uuid, result)
			}
		}
	}
}
