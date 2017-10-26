package thingiverseio

import "github.com/ThingiverseIO/thingiverseio/message"

//go:generate evt2gogen -t *Result -n Result
//go:generate evt2gogen -t *Request -n Request

type Result = message.Result
type Request = message.Request
