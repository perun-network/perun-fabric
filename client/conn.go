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
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"

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
func NewIdentity(mspID, certPath string) (*identity.X509Identity, *wallet.Address, string, error) {
	cert, err := ReadCertificate(certPath)
	if err != nil {
		return nil, nil, "", fmt.Errorf("loading certificate: %w", err)
	}

	addr, err := wallet.AddressFromX509Certificate(cert)
	if err != nil {
		return nil, nil, "", err
	}

	id, err := identity.NewX509Identity(mspID, cert)
	if err != nil {
		return nil, nil, "", fmt.Errorf("creating X509Identity: %w", err)
	}

	onChainID, err := calcOnChainCertID(cert)
	if err != nil {
		return nil, nil, "", fmt.Errorf("creating On Chain Idenity: %w", err)
	}

	return id, addr, onChainID, nil
}

// NewAccountWithSigner generates a new account and singer based on the given path of the private key file.
func NewAccountWithSigner(privateKeyPEMPath string) (identity.Sign, *wallet.Account, error) {
	privateKeyPEM, err := os.ReadFile(privateKeyPEMPath)
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

// ReadCertificate takes the given path to the certificate file, reads it and returns a x509.Certificate.
func ReadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

// calcOnChainCertID returns a unique ID associated with the invoking identity.
// This code is a direct copy of GetID() in the fabric-chaincode-go sdk as it is not exposed there.
// https://github.com/hyperledger/fabric-chaincode-go/blob/9207360bbddd5952479c24154353b82c4c044677/pkg/cid/cid.go#L96
func calcOnChainCertID(cert *x509.Certificate) (string, error) {
	// When IdeMix, c.cert is nil for x509 type
	// Here will return "", as there is no x509 type cert for generate id value with logic below.
	if cert == nil {
		return "", fmt.Errorf("cannot determine identity")
	}
	// The leading "x509::" distinguishes this as an X509 certificate, and
	// the subject and issuer DNs uniquely identify the X509 certificate.
	// The resulting ID will remain the same if the certificate is renewed.
	id := fmt.Sprintf("x509::%s::%s", calcDN(&cert.Subject), calcDN(&cert.Issuer))
	return base64.StdEncoding.EncodeToString([]byte(id)), nil
}

// calcDN calculates the DN (distinguished name) associated with a pkix.Name.
// This code is a direct copy of getDN() in the fabric-chaincode-go sdk as it is not exposed there.
// https://github.com/hyperledger/fabric-chaincode-go/blob/9207360bbddd5952479c24154353b82c4c044677/pkg/cid/cid.go#L218
func calcDN(name *pkix.Name) string { //nolint:gocognit
	r := name.ToRDNSequence()
	s := ""
	for i := 0; i < len(r); i++ {
		rdn := r[len(r)-1-i]
		if i > 0 {
			s += ","
		}
		for j, tv := range rdn {
			if j > 0 {
				s += "+"
			}
			typeString := tv.Type.String()
			typeName, ok := attributeTypeNames[typeString]
			if !ok {
				derBytes, err := asn1.Marshal(tv.Value)
				if err == nil {
					s += typeString + "=#" + hex.EncodeToString(derBytes)
					continue // No value escaping necessary.
				}
				typeName = typeString
			}
			valueString := fmt.Sprint(tv.Value)
			escaped := ""
			begin := 0
			for idx, c := range valueString {
				if (idx == 0 && (c == ' ' || c == '#')) ||
					(idx == len(valueString)-1 && c == ' ') {
					escaped += valueString[begin:idx]
					escaped += "\\" + string(c)
					begin = idx + 1
					continue
				}
				switch c {
				case ',', '+', '"', '\\', '<', '>', ';':
					escaped += valueString[begin:idx]
					escaped += "\\" + string(c)
					begin = idx + 1
				}
			}
			escaped += valueString[begin:]
			s += typeName + "=" + escaped
		}
	}
	return s
}

var attributeTypeNames = map[string]string{
	"2.5.4.6":  "C",
	"2.5.4.10": "O",
	"2.5.4.11": "OU",
	"2.5.4.3":  "CN",
	"2.5.4.5":  "SERIALNUMBER",
	"2.5.4.7":  "L",
	"2.5.4.8":  "ST",
	"2.5.4.9":  "STREET",
	"2.5.4.17": "POSTALCODE",
}
