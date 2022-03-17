// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	gwproto "github.com/hyperledger/fabric-protos-go/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

const (
	fabricSamplesEnv = "FABRIC_SAMPLES_DIR"

	mspID        = "Org1MSP"
	peerEndpoint = "localhost:7051"
	gatewayPeer  = "peer0.org1.example.com"
	ChannelName  = "mychannel"
)

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

func cryptoPath() string {
	return path.Join(fabricSamplesPath, "test-network/organizations/peerOrganizations/org1.example.com")
}
func certPath() string    { return cryptoPath() + "/users/User1@org1.example.com/msp/signcerts/cert.pem" }
func keyPath() string     { return cryptoPath() + "/users/User1@org1.example.com/msp/keystore/" }
func tlsCertPath() string { return cryptoPath() + "/peers/peer0.org1.example.com/tls/ca.crt" }

// NewGrpcConnection creates a gRPC connection to the Gateway server.
func NewGrpcConnection() (*grpc.ClientConn, error) {
	certificate, err := loadCertificate(tlsCertPath())
	if err != nil {
		return nil, fmt.Errorf("loading certificate: %w", err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		return nil, fmt.Errorf("creating gRPC connection: %w", err)
	}

	return connection, nil
}

// NewIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func NewIdentity() (*identity.X509Identity, error) {
	certificate, err := loadCertificate(certPath())
	if err != nil {
		return nil, fmt.Errorf("loading certificate: %w", err)
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		return nil, fmt.Errorf("creating X509Identity: %w", err)
	}

	return id, nil
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

// NewSign creates a function that generates a digital signature from a message digest using a private key.
func NewSign() (identity.Sign, error) {
	files, err := ioutil.ReadDir(keyPath())
	if err != nil {
		return nil, fmt.Errorf("reading private key directory: %w", err)
	}

	privateKeyPEM, err := ioutil.ReadFile(path.Join(keyPath(), files[0].Name()))
	if err != nil {
		return nil, fmt.Errorf("reading private key file: %w", err)
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parsing private key PEM: %w", err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return nil, fmt.Errorf("creating signer: %w", err)
	}

	return sign, nil
}

func NewGateway(clientConn *grpc.ClientConn) (*client.Gateway, error) {
	id, err := NewIdentity()
	if err != nil {
		return nil, err
	}
	sign, err := NewSign()
	if err != nil {
		return nil, err
	}

	// Create a Gateway connection for a specific client identity
	return client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConn),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
}

// FatalErr prints msg followed by err and then exits the program immedately, if
// err != nil.
func FatalErr(msg string, err error) {
	if err != nil {
		log.Fatalf("Error %s: [%T] %+v", msg, err, err)
	}
}

// FatalErr prints msg followed by err, if err != nil. The error err is then
// parsed as a client error and its full details are also printed
// The program is then exited immedately.
func FatalClientErr(msg string, err error) {
	if err != nil {
		log.Fatalf("Error %s: [%T] %+v\n%s", msg, err, err, ParseClientErr(err))
	}
}

// ParseClientErr parses the full details of err as a fabric client error.
func ParseClientErr(err error) string {
	var s strings.Builder

	switch err := err.(type) {
	case *client.EndorseError:
		s.WriteString(fmt.Sprintf("Endorse error with gRPC status %v: %s\n", status.Code(err), err))
	case *client.SubmitError:
		s.WriteString(fmt.Sprintf("Submit error with gRPC status %v: %s\n", status.Code(err), err))
	case *client.CommitStatusError:
		if errors.Is(err, context.DeadlineExceeded) {
			s.WriteString(fmt.Sprintf("Timeout waiting for transaction %s commit status: %s\n", err.TransactionID, err))
		} else {
			s.WriteString(fmt.Sprintf("Error obtaining commit status with gRPC status %v: %s\n", status.Code(err), err))
		}
	case *client.CommitError:
		s.WriteString(fmt.Sprintf("Transaction %s failed to commit with status %d: %s\n", err.TransactionID, int32(err.Code), err))
	}

	//Any error that originates from a peer or orderer node external to the gateway will have its details
	//embedded within the gRPC status error. The following code shows how to extract that.
	statusErr := status.Convert(err)
	for _, detail := range statusErr.Details() {
		errDetail := detail.(*gwproto.ErrorDetail)
		s.WriteString(fmt.Sprintf("Error from endpoint: %s, mspId: %s, message: %s\n", errDetail.Address, errDetail.MspId, errDetail.Message))
	}

	return s.String()
}
