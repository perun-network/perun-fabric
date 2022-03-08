// SPDX-License-Identifier: Apache-2.0

package chaincode

import "fmt"

func stringWithErr(s fmt.Stringer, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return s.String(), nil
}
