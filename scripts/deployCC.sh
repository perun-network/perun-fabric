#!/bin/bash
#
# SPDX-License-Identifier: Apache-2.0

set -e

export ORIGIN="$(pwd)"
export FABRIC_SAMPLES_DIR="${FABRIC_SAMPLES_DIR:-../fabric-samples}"
export TEST_NETWORK_DIR="${FABRIC_SAMPLES_DIR}/test-network"
export NETWORK_CMD="${TEST_NETWORK_DIR}/network.sh"
export CHAINCODE="$1"
export ADJ="adjudicator"

function ensureNetworkUp() {
  if [ -d "${TEST_NETWORK_DIR}/organizations/peerOrganizations" ]; then
    echo "Test network seems to be running."
    return
  fi

  echo "Starting test network..."
  "${NETWORK_CMD}" up createChannel -c mychannel -ca
}

ensureNetworkUp
"${NETWORK_CMD}" deployCC -ccn "assetholder" -ccp "${ORIGIN}/chaincode/assetholder/" -ccl go
"${NETWORK_CMD}" deployCC -ccn "adjudicator" -ccp "${ORIGIN}/chaincode/adjudicator/" -ccl go
