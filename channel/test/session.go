// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/perun-network/perun-fabric/channel"
	"github.com/perun-network/perun-fabric/channel/binding"
	"google.golang.org/grpc"

	"github.com/perun-network/perun-fabric/wallet"
)

type AdjudicatorSession struct {
	Adjudicator channel.Adjudicator
	Binding     binding.Adjudicator
	Account     *wallet.Account
	conn        *grpc.ClientConn
	gw          *client.Gateway
}

func NewAdjudicatorSession(org Org, chaincode string) (_ *AdjudicatorSession, err error) {
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
	return &AdjudicatorSession{
		Adjudicator: *channel.NewAdjudicator(network, chaincode),
		Binding:     *binding.NewAdjudicatorBinding(network, chaincode),
		Account:     acc,
		conn:        clientConn,
		gw:          gateway,
	}, nil
}

func (s AdjudicatorSession) Close() error {
	err0 := s.gw.Close()
	err1 := s.conn.Close()
	if (err0 != nil) && (err1 != nil) {
		return fmt.Errorf("closing gateway: %v; closing connection: %w", err0, err1)
	} else if err0 != nil {
		return err0
	}
	return err1
}
