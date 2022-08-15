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
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	gwproto "github.com/hyperledger/fabric-protos-go/gateway"
	"google.golang.org/grpc/status"
)

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

	// Any error that originates from a peer or orderer node external to the gateway will have its details
	// embedded within the gRPC status error. The following code shows how to extract that.
	statusErr := status.Convert(err)
	for _, detail := range statusErr.Details() {
		errDetail, _ := detail.(*gwproto.ErrorDetail)
		s.WriteString(fmt.Sprintf("Error from endpoint: %s, mspId: %s, message: %s\n", errDetail.Address, errDetail.MspId, errDetail.Message))
	}

	return s.String()
}

// IsChannelUnknownErr checks if the given error indicates the channel is unknown.
func IsChannelUnknownErr(err error) bool {
	e := ParseClientErr(err)
	return strings.Contains(e, "chaincode response 500, unknown channel")
}

// IsUnderfundedErr checks if the given error indicates the channel is underfunded.
func IsUnderfundedErr(err error) bool {
	e := ParseClientErr(err)
	return strings.Contains(e, "channel underfunded")
}
