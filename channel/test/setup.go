//  Copyright 2022 PolyCrypt GmbH
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package test

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"

	pclient "github.com/perun-network/perun-fabric/client"
	"github.com/perun-network/perun-fabric/wallet"
)

const (
	// ChannelName is the name of the fabric channel the clients communicate in.
	// It is not the state channel.
	ChannelName = "mychannel"
	// AdjudicatorName is the adjudicator chaincode name.
	AdjudicatorName = "adjudicator"
	// AssetholderName is the assetholder chaincode name.
	AssetholderName = "assetholder"

	fabricSamplesEnv = "FABRIC_SAMPLES_DIR"
	evalTimeout      = 5 * time.Second
	endorseTimeout   = 15 * time.Second
	submitTimeout    = 5 * time.Second
	commitTimeout    = 1 * time.Minute
)

// Org is a fabric organization.
type Org string

const (
	// Org1 indicates demo organization one in fabric-samples.
	Org1 Org = "org1"
	// Org2 indicates demo organization two in fabric-samples.
	Org2 Org = "org2"
)

// OrgNum returns the organisation identifier.
func OrgNum(n uint) Org {
	switch n {
	case 1, 2: //nolint:gomnd
		return Org(fmt.Sprintf("org%d", n))
	}
	panic(fmt.Sprintf("invalid org number %d", n))
}

// Port returns the organisations port.
func (org Org) Port() string {
	switch org {
	case Org1:
		return "7051"
	case Org2:
		return "9051"
	}
	panic("invalid org: " + org)
}

// fabricSamplesPath should point to the checked-out fabric-samples repository
// found at https://github.com/hyperledger/fabric-samples
//
// The default value assumes the repository to be checked out next to the
// perun-fabric directory.
var fabricSamplesPath = "../fabric-samples/"

func init() {
	if fspath, ok := os.LookupEnv(fabricSamplesEnv); ok {
		fabricSamplesPath = fspath
	}
}

func mspID(org Org) string        { return "O" + string(org[1:]) + "MSP" }
func peerEndpoint(org Org) string { return "localhost:" + org.Port() }
func gatewayPeer(org Org) string  { return "peer0." + string(org) + ".example.com" }

func cryptoPath(org Org) string {
	return path.Join(fabricSamplesPath, "test-network/organizations/peerOrganizations/"+string(org)+".example.com")
}

func certPath(org Org) string {
	return cryptoPath(org) + "/users/User1@" + string(org) + ".example.com/msp/signcerts/cert.pem"
}

func keyDir(org Org) string {
	return cryptoPath(org) + "/users/User1@" + string(org) + ".example.com/msp/keystore/"
}

func keyPath(org Org) (string, error) {
	kp := keyDir(org)
	files, err := os.ReadDir(kp)
	if err != nil {
		return "", fmt.Errorf("reading private key directory: %w", err)
	}
	// the first and only file in the keystore is the secret key
	return path.Join(kp, files[0].Name()), nil
}

func tlsCertPath(org Org) string {
	return cryptoPath(org) + "/peers/peer0." + string(org) + ".example.com/tls/ca.crt"
}

// NewGrpcConnection creates a gRPC connection to the Gateway server.
func NewGrpcConnection(org Org) (*grpc.ClientConn, error) {
	return pclient.NewGrpcConnection(gatewayPeer(org), peerEndpoint(org), tlsCertPath(org))
}

// NewIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func NewIdentity(org Org) (*identity.X509Identity, *wallet.Address, string, error) {
	return pclient.NewIdentity(mspID(org), certPath(org))
}

// NewAccountWithSigner creates a function that generates a digital signature from a message digest using a private key.
func NewAccountWithSigner(org Org) (identity.Sign, *wallet.Account, error) {
	path, err := keyPath(org)
	if err != nil {
		return nil, nil, err
	}
	return pclient.NewAccountWithSigner(path)
}

// NewGateway creates a Gateway for a specific client identity with several timeouts for gRPC calls.
func NewGateway(org Org, clientConn *grpc.ClientConn) (*client.Gateway, *wallet.Account, string, error) {
	id, addr, onChainID, err := NewIdentity(org)
	if err != nil {
		return nil, nil, "", err
	}
	sign, acc, err := NewAccountWithSigner(org)
	if err != nil {
		return nil, nil, "", err
	}

	if acca := acc.Address(); !acca.Equal(addr) {
		return nil, nil, "", fmt.Errorf("identity and signer public key mismatch, %v != %v", acca, addr)
	}

	// Create a Gateway connection for a specific client identity
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConn),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(evalTimeout),
		client.WithEndorseTimeout(endorseTimeout),
		client.WithSubmitTimeout(submitTimeout),
		client.WithCommitStatusTimeout(commitTimeout),
	)

	return gw, acc, onChainID, err
}

// FatalErr prints msg followed by err and then exits the program immediately, if
// err != nil.
func FatalErr(msg string, err error) {
	if err != nil {
		log.Fatalf("Error %s: [%T] %+v", msg, err, err)
	}
}

// FatalClientErr prints msg followed by err, if err != nil. The error err is then
// parsed as a client error and its full details are also printed
// The program is then exited immedately.
func FatalClientErr(msg string, err error) {
	if err != nil {
		log.Fatalf("Error %s: [%T] %+v\n%s", msg, err, err, pclient.ParseClientErr(err))
	}
}
