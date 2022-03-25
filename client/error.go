// SPDX-License-Identifier: Apache-2.0

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

	//Any error that originates from a peer or orderer node external to the gateway will have its details
	//embedded within the gRPC status error. The following code shows how to extract that.
	statusErr := status.Convert(err)
	for _, detail := range statusErr.Details() {
		errDetail := detail.(*gwproto.ErrorDetail)
		s.WriteString(fmt.Sprintf("Error from endpoint: %s, mspId: %s, message: %s\n", errDetail.Address, errDetail.MspId, errDetail.Message))
	}

	return s.String()
}
