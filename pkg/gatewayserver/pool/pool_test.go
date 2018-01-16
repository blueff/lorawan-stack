// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package pool_test

import (
	"context"
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/band"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/pool"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

type dummyLink struct {
	NextUplink chan *ttnpb.GatewayUp

	context       context.Context
	cancelContext context.CancelFunc

	AcceptDownlink       bool
	AcceptSendingUplinks bool
}

func (d *dummyLink) Send(*ttnpb.GatewayDown) error {
	if d.AcceptDownlink {
		return nil
	}
	return errors.New("Downlink refused")
}

func (d *dummyLink) Recv() (*ttnpb.GatewayUp, error) {
	up := <-d.NextUplink
	if !d.AcceptSendingUplinks {
		return nil, errors.New("Couldn't receive uplink")
	}
	return up, nil
}

func (d *dummyLink) Context() context.Context {
	if d.context == nil {
		return context.Background()
	}
	return d.context
}

func newPoolConnection() ttnpb.GtwGs_LinkServer {
	return nil
}

func downlink() *ttnpb.DownlinkMessage {
	return &ttnpb.DownlinkMessage{}
}

func ExamplePool() {
	p := pool.NewPool(log.Noop, time.Millisecond)

	gatewayInfo := ttnpb.GatewayIdentifier{GatewayID: "my-kerlink"}
	upMessages, err := p.Subscribe(gatewayInfo, newPoolConnection(), ttnpb.FrequencyPlan{BandID: band.EU_863_870})
	if err != nil {
		panic(err)
	}

	go func() {
		for upMessage := range upMessages {
			if upMessage.GatewayStatus != nil {
				fmt.Println("Gateway status received")
			}
			if upMessage.UplinkMessages != nil && len(upMessage.UplinkMessages) > 0 {
				fmt.Println("Uplink received from gateway", gatewayInfo.GatewayID, "!")
			}
		}
	}()

	go func() {
		time.Sleep(5 * time.Second)
		p.Send(gatewayInfo, &ttnpb.GatewayDown{DownlinkMessage: downlink()})
		fmt.Println("Downlink sent to gateway!")
	}()
}