package core

import (
	"time"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/descriptor"
	"github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/thingiverseio/network"
)

type InputCore struct {
	Core
}

func NewInputCore(desc descriptor.Descriptor, usrCfg *config.UserConfig,
	tracker network.Tracker, provider ...network.Provider) (i InputCore, err error) {

	tags := desc.AsTagset()
	tags.Merge(usrCfg.Tags)

	intCfg := config.NewInternalConfig(false, tags)

	c, err := Initialize(&config.Config{
		Internal: intCfg,
		User:     usrCfg,
	}, tracker, provider...)

	i = InputCore{
		Core: c,
	}

	i.tracker.Arrivals().Where(i.isInterestingArrival).Listen(i.onArrival)

	return

}

func (i InputCore) isInterestingArrival(a network.Arrival) (is bool) {
	i.log.Debugf("Checking if arrived peer is interesting: %s", a.UUID)

	if is = a.IsOutput; !is {
		i.log.Debug("Peer is not an output.")
		return
	}
	if is = a.Supported(i.provider.Details); !is {
		i.log.Debug("Peer network is not supported")
	}

	return
}

func (i InputCore) onArrival(a network.Arrival) {
	i.log.Debug("Peer arrived")
	conn, err := i.provider.Connect(a.Details, a.UUID)
	if err != nil {
		return
	}

	in := i.provider.Messages().Where(network.FromSender(a.UUID))
	defer in.Close()

	next := in.FirstWhere(network.OfType(message.HELLOOK))

	// Send Hello

	conn.Send(&message.Hello{
		NetworkDetails: i.provider.EncodedDetails,
		Tag:            i.config.Internal.Tags.GetFirst(),
	})

	if next.WaitUntilTimeout(1 * time.Second) {
		i.log.Debug("Received HELLOOK")
		if next.GetResult().Decode().(*message.HelloOk).Have {
			i.log.Debug("First Tag is supported, checking all tags")
			for _, tag := range i.config.Internal.Tags.AsArray() {

				i.log.Debug("Checking for tag", tag)
				next = in.FirstWhere(network.OfType(message.HAVE))

				// Send DoHave
				conn.Send(&message.DoHave{
					Tag: tag,
				})

				if next.WaitUntilTimeout(1 * time.Second) {
					i.log.Debug("Got message HAVE")
					if !next.GetResult().Decode().(*message.Have).Have {
						// Send End

						conn.Close()
						return
					}
				} else {
					conn.Close()
					return
				}
			}

			// Send Connect

			conn.Send(&message.Connect{})
			i.Fire(connectEvent{}, conn)
			return
		}

	}
	i.log.Debug("Timeout")

	conn.Close()

}
