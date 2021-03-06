
/*
 * generated by event_generator
 *
 * DO NOT EDIT
 */

package thingiverseio

import "github.com/joernweissenborn/eventual2go"



type RequestCompleter struct {
	*eventual2go.Completer
}

func NewRequestCompleter() *RequestCompleter {
	return &RequestCompleter{eventual2go.NewCompleter()}
}

func (c *RequestCompleter) Complete(d *Request) {
	c.Completer.Complete(d)
}

func (c *RequestCompleter) Future() *RequestFuture {
	return &RequestFuture{c.Completer.Future()}
}

type RequestFuture struct {
	*eventual2go.Future
}

func (f *RequestFuture) Result() *Request {
	return f.Future.Result().(*Request)
}

type RequestCompletionHandler func(*Request) *Request

func (ch RequestCompletionHandler) toCompletionHandler() eventual2go.CompletionHandler {
	return func(d eventual2go.Data) eventual2go.Data {
		return ch(d.(*Request))
	}
}

func (f *RequestFuture) Then(ch RequestCompletionHandler) *RequestFuture {
	return &RequestFuture{f.Future.Then(ch.toCompletionHandler())}
}

func (f *RequestFuture) AsChan() chan *Request {
	c := make(chan *Request, 1)
	cmpl := func(d chan *Request) RequestCompletionHandler {
		return func(e *Request) *Request {
			d <- e
			close(d)
			return e
		}
	}
	ecmpl := func(d chan *Request) eventual2go.ErrorHandler {
		return func(error) (eventual2go.Data, error) {
			close(d)
			return nil, nil
		}
	}
	f.Then(cmpl(c))
	f.Err(ecmpl(c))
	return c
}

type RequestStreamController struct {
	*eventual2go.StreamController
}

func NewRequestStreamController() *RequestStreamController {
	return &RequestStreamController{eventual2go.NewStreamController()}
}

func (sc *RequestStreamController) Add(d *Request) {
	sc.StreamController.Add(d)
}

func (sc *RequestStreamController) Join(s *RequestStream) {
	sc.StreamController.Join(s.Stream)
}

func (sc *RequestStreamController) JoinFuture(f *RequestFuture) {
	sc.StreamController.JoinFuture(f.Future)
}

func (sc *RequestStreamController) Stream() *RequestStream {
	return &RequestStream{sc.StreamController.Stream()}
}

type RequestStream struct {
	*eventual2go.Stream
}

type RequestSubscriber func(*Request)

func (l RequestSubscriber) toSubscriber() eventual2go.Subscriber {
	return func(d eventual2go.Data) { l(d.(*Request)) }
}

func (s *RequestStream) Listen(ss RequestSubscriber) *eventual2go.Completer {
	return s.Stream.Listen(ss.toSubscriber())
}

func (s *RequestStream) ListenNonBlocking(ss RequestSubscriber) *eventual2go.Completer {
	return s.Stream.ListenNonBlocking(ss.toSubscriber())
}

type RequestFilter func(*Request) bool

func (f RequestFilter) toFilter() eventual2go.Filter {
	return func(d eventual2go.Data) bool { return f(d.(*Request)) }
}

func toRequestFilterArray(f ...RequestFilter) (filter []eventual2go.Filter){

	filter = make([]eventual2go.Filter, len(f))
	for i, el := range f {
		filter[i] = el.toFilter()
	}
	return
}

func (s *RequestStream) Where(f ...RequestFilter) *RequestStream {
	return &RequestStream{s.Stream.Where(toRequestFilterArray(f...)...)}
}

func (s *RequestStream) WhereNot(f ...RequestFilter) *RequestStream {
	return &RequestStream{s.Stream.WhereNot(toRequestFilterArray(f...)...)}
}

func (s *RequestStream) TransformWhere(t eventual2go.Transformer, f ...RequestFilter) *eventual2go.Stream {
	return s.Stream.TransformWhere(t, toRequestFilterArray(f...)...)
}

func (s *RequestStream) Split(f RequestFilter) (*RequestStream, *RequestStream)  {
	return s.Where(f), s.WhereNot(f)
}

func (s *RequestStream) First() *RequestFuture {
	return &RequestFuture{s.Stream.First()}
}

func (s *RequestStream) FirstWhere(f... RequestFilter) *RequestFuture {
	return &RequestFuture{s.Stream.FirstWhere(toRequestFilterArray(f...)...)}
}

func (s *RequestStream) FirstWhereNot(f ...RequestFilter) *RequestFuture {
	return &RequestFuture{s.Stream.FirstWhereNot(toRequestFilterArray(f...)...)}
}

func (s *RequestStream) AsChan() (c chan *Request, stop *eventual2go.Completer) {
	c = make(chan *Request)
	stop = s.Listen(pipeToRequestChan(c))
	stop.Future().Then(closeRequestChan(c))
	return
}

func pipeToRequestChan(c chan *Request) RequestSubscriber {
	return func(d *Request) {
		c <- d
	}
}

func closeRequestChan(c chan *Request) eventual2go.CompletionHandler {
	return func(d eventual2go.Data) eventual2go.Data {
		close(c)
		return nil
	}
}

type RequestCollector struct {
	*eventual2go.Collector
}

func NewRequestCollector() *RequestCollector {
	return &RequestCollector{eventual2go.NewCollector()}
}

func (c *RequestCollector) Add(d *Request) {
	c.Collector.Add(d)
}

func (c *RequestCollector) AddFuture(f *RequestFuture) {
	c.Collector.Add(f.Future)
}

func (c *RequestCollector) AddStream(s *RequestStream) {
	c.Collector.AddStream(s.Stream)
}

func (c *RequestCollector) Get() *Request {
	return c.Collector.Get().(*Request)
}

func (c *RequestCollector) Preview() *Request {
	return c.Collector.Preview().(*Request)
}

type RequestObservable struct {
	*eventual2go.Observable
}

func NewRequestObservable (value *Request) (o *RequestObservable) {
	return &RequestObservable{eventual2go.NewObservable(value)}
}

func (o *RequestObservable) Value() *Request {
	return o.Observable.Value().(*Request)
}

func (o *RequestObservable) Change(value *Request) {
	o.Observable.Change(value)
}

func (o *RequestObservable) OnChange(s RequestSubscriber) (cancel *eventual2go.Completer) {
	return o.Observable.OnChange(s.toSubscriber())
}

func (o *RequestObservable) Stream() (*RequestStream) {
	return &RequestStream{o.Observable.Stream()}
}


func (o *RequestObservable) AsChan() (c chan *Request, cancel *eventual2go.Completer) {
	return o.Stream().AsChan()
}

func (o *RequestObservable) NextChange() (f *RequestFuture) {
	return o.Stream().First()
}
