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

package client

import (
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/perun-network/perun-fabric/wallet"
)

// NewGrpcConnection creates a gRPC connection to the Gateway server.
func NewGrpcConnection(gatewayPeer, peerEndpoint, peerTLSCertPath string) (*grpc.ClientConn, error) {
	cert, err := ReadCertificate(peerTLSCertPath)
	if err != nil {
		return nil, fmt.Errorf("loading certificate: %w", err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(cert)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		return nil, fmt.Errorf("creating gRPC connection: %w", err)
	}

	return connection, nil
}

// NewIdentity creates a client identity for a Gateway connection using an X.509
// certificate. It also returns the corresponding Fabric backend Address.
func NewIdentity(mspID, certPath string) (*identity.X509Identity, *wallet.Address, error) {
	cert, err := ReadCertificate(certPath)
	if err != nil {
		return nil, nil, fmt.Errorf("loading certificate: %w", err)
	}

	addr, err := wallet.AddressFromX509Certificate(cert)
	if err != nil {
		return nil, nil, err
	}

	id, err := identity.NewX509Identity(mspID, cert)
	if err != nil {
		return nil, nil, fmt.Errorf("creating X509Identity: %w", err)
	}

	return id, addr, nil
}

func NewAccountWithSigner(privateKeyPEMPath string) (identity.Sign, *wallet.Account, error) {
	privateKeyPEM, err := ioutil.ReadFile(privateKeyPEMPath)
	if err != nil {
		return nil, nil, fmt.Errorf("reading private key file: %w", err)
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing private key PEM: %w", err)
	}

	ecdsaPrivateKey, ok := privateKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("private key of unexpected type %T", privateKey)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("creating signer: %w", err)
	}

	return sign, (*wallet.Account)(ecdsaPrivateKey), nil
}

func ReadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}
