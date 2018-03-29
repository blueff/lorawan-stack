// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package component_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/rpcserver"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestPeers(t *testing.T) {
	a := assertions.New(t)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer lis.Close()

	// Starting gRPC server
	{
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		srv := rpcserver.New(ctx)
		go srv.Serve(lis)
		defer srv.Stop()
	}

	var c *component.Component

	// Starting component
	{
		config := &component.Config{
			ServiceBase: config.ServiceBase{Cluster: config.Cluster{
				Name:          "test-cluster",
				NetworkServer: lis.Addr().String(),
				TLS:           false,
			}},
		}

		c, err = component.New(test.GetLogger(t), config)
		a.So(err, should.BeNil)
		err = c.Start()
		a.So(err, should.BeNil)
	}

	time.Sleep(5 * time.Millisecond) // Wait for peers to join cluster

	unusedRoles := []ttnpb.PeerInfo_Role{
		ttnpb.PeerInfo_APPLICATION_SERVER,
		ttnpb.PeerInfo_GATEWAY_SERVER,
		ttnpb.PeerInfo_JOIN_SERVER,
		ttnpb.PeerInfo_IDENTITY_SERVER,
	}

	// GetPeer
	{
		peer := c.GetPeer(ttnpb.PeerInfo_NETWORK_SERVER, nil, nil)
		a.So(peer, should.NotBeNil)
		conn := peer.Conn()
		a.So(conn, should.NotBeNil)

		for _, role := range unusedRoles {
			peer = c.GetPeer(role, nil, nil)
			a.So(peer, should.BeNil)
		}
	}

	// GetPeers
	{
		peers := c.GetPeers(ttnpb.PeerInfo_NETWORK_SERVER, nil)
		a.So(peers, should.HaveLength, 1)

		for _, role := range unusedRoles {
			peers = c.GetPeers(role, nil)
			a.So(peers, should.HaveLength, 0)
		}
	}
}