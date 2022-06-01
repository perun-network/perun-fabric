// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/perun-network/perun-fabric/channel"
	"github.com/perun-network/perun-fabric/channel/binding"
	"github.com/perun-network/perun-fabric/wallet"
	"google.golang.org/grpc"
)

type Session struct {
	Adjudicator channel.Adjudicator
	Binding     binding.Adjudicator
	Funder      channel.Funder
	Account     *wallet.Account
	conn        *grpc.ClientConn
	gw          *client.Gateway
}

func NewTestSession(org Org, adjudicator string, assetholder string) (_ *Session, err error) {
	clientConn, err := NewGrpcConnection(org)
	if err != nil {
		return nil, fmt.Errorf("creating client conn: %w", err)
	}

	// Create a Gateway connection for a specific client identity
	gateway, acc, err := NewGateway(org, clientConn)
	if err != nil {
		clientConn.Close()
		return nil, fmt.Errorf("connecting to gateway: %w", err)
	}

	network := gateway.GetNetwork(ChannelName)
	return &Session{
		Adjudicator: *channel.NewAdjudicator(network, adjudicator),
		Binding:     *binding.NewAdjudicatorBinding(network, adjudicator),
		Funder:      *channel.NewFunder(network, assetholder),
		Account:     acc,
		conn:        clientConn,
		gw:          gateway,
	}, nil
}

func (s Session) Close() error {
	err0 := s.gw.Close()
	err1 := s.conn.Close()
	if (err0 != nil) && (err1 != nil) {
		return fmt.Errorf("closing gateway: %v; closing connection: %w", err0, err1)
	} else if err0 != nil {
		return err0
	}
	return err1
}