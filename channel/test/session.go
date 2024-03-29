// Copyright 2022 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"fmt"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	adj "github.com/perun-network/perun-fabric/adjudicator"
	"github.com/perun-network/perun-fabric/channel"
	"github.com/perun-network/perun-fabric/channel/binding"
	"github.com/perun-network/perun-fabric/wallet"
	"google.golang.org/grpc"
)

// Session contains all parts to test the fabric backend with a specific party.
type Session struct {
	ClientFabricID adj.AccountID
	Adjudicator    *channel.Adjudicator
	Binding        *binding.Adjudicator
	Funder         *channel.Funder
	Account        *wallet.Account
	conn           *grpc.ClientConn
	gw             *client.Gateway
}

// NewTestSession generates a new session for testing the fabric backend with.
// The user certificate is chosen by the given organization.
// For connecting with the chaincode the adjudicator id is given.
func NewTestSession(org Org, adjudicator string) (*Session, error) {
	clientConn, err := NewGrpcConnection(org)
	if err != nil {
		return nil, fmt.Errorf("creating client conn: %w", err)
	}

	// Create a Gateway connection for a specific client identity
	gateway, acc, clientID, err := NewGateway(org, clientConn)
	if err != nil {
		clientConn.Close()
		return nil, fmt.Errorf("connecting to gateway: %w", err)
	}

	network := gateway.GetNetwork(ChannelName)
	return &Session{
		ClientFabricID: clientID,
		Adjudicator:    channel.NewAdjudicator(network, adjudicator, clientID),
		Binding:        binding.NewAdjudicatorBinding(network, adjudicator),
		Funder:         channel.NewFunder(network, adjudicator),
		Account:        acc,
		conn:           clientConn,
		gw:             gateway,
	}, nil
}

// Close closes the connection to fabric for this session.
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
