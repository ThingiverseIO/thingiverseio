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

type Export struct {
	cfg       *config.Config
	m         *manager.Manager
	listeners map[string]map[config.UUID]interface{}
	logger    *log.Logger
	r         *eventual2go.Reactor
	requests  *messages.RequestStream
}

func NewExportFromConfig(cfg *config.Config) (e *Export, err error) {
	m, err := manager.New(cfg)
	e = &Export{
		cfg:       cfg,
		m:         m,
		requests:  &messages.RequestStream{m.MessagesOfType(messages.RESULT).Transform(connection.ToMessage)},
		listeners: map[string]map[config.UUID]interface{}{},
		logger:    log.New(cfg.Logger(), fmt.Sprintf("%s EXPORT", cfg.UUID()),0),
		r:         eventual2go.NewReactor(),
	}

	e.r.React(listenEvent, e.onListen)
	e.r.AddStream(listenEvent, m.MessagesOfType(messages.LISTEN).Stream)

	e.r.React(stopListenEvent, e.onStopListen)
	e.r.AddStream(stopListenEvent, m.MessagesOfType(messages.STOPLISTEN).Stream)

	e.r.React(peerGoneEvent, e.onPeerGone)

	e.r.React(replyEvent, e.deliverResult)
	return
}

func (e *Export) Reply(r *messages.Request, params interface{}) {
	res := messages.NewResult(e.cfg.UUID(), r, params)
	e.r.Fire(replyEvent, res)
}

func (e *Export) ReplyEncoded(r *messages.Request, params []byte) {
	res := messages.NewEncodedResult(e.cfg.UUID(), r, params)
	e.r.Fire(replyEvent, res)
}

func (e *Export) Emit(function string, inparams interface{}, outparams interface{}) {
	req := messages.NewRequest("", function, messages.MANY2ONE, inparams)
	e.Reply(req, outparams)
}

func (e *Export) EmitEncoded(function string, inparams []byte, outparams []byte) {
	req := messages.NewEncodedRequest("", function, messages.MANY2ONE, inparams)
	e.ReplyEncoded(req, outparams)
}

func (e *Export) Requests() *messages.RequestStream {
	return e.requests
}

func (e *Export) onListen(d eventual2go.Data) {
	m := d.(connection.Message)
	l := messages.Unflatten(m.Payload)
	f := l.(*messages.Listen).Function
	e.logger.Println("New Listener", m.Sender, f)
	_, ok := e.listeners[f]
	if ok {
		e.listeners[f][m.Sender] = nil
	} else {
		e.listeners[f] = map[config.UUID]interface{}{m.Sender: nil}
	}
	e.r.AddFuture(peerGoneEvent, e.m.PeerLeave(m.Sender).Future)
}

func (e *Export) onStopListen(d eventual2go.Data) {
	m := d.(connection.Message)
	l := messages.Unflatten(m.Payload)
	f := l.(*messages.StopListen).Function
	e.removePeerListen(m.Sender, f)
}

func (e *Export) onPeerGone(d eventual2go.Data) {
	uuid := d.(config.UUID)
	for f, _ := range e.listeners {
		e.removePeerListen(uuid, f)
	}
}

func (e *Export) removePeerListen(uuid config.UUID, f string) {
	_, ok := e.listeners[f]
	if !ok {
		return
	}
	_, ok = e.listeners[f][uuid]
	if ok {
		delete(e.listeners[f], uuid)
	}
}

func (e *Export) deliverResult(d eventual2go.Data) {
	result := d.(*messages.Result)
	e.logger.Println("Delivering result", result.Request.Function, result.Request.CallType)
	switch result.Request.CallType {
	case messages.ONE2MANY, messages.ONE2ONE:
		e.m.SendTo(result.Request.UUID, result)

	case messages.MANY2MANY, messages.MANY2ONE:
		if ls, ok := e.listeners[result.Request.Function]; ok {
			for uuid := range ls {
				e.m.SendTo(uuid, result)
			}
		}
	}
}
