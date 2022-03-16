#!/bin/bash
#
# SPDX-License-Identifier: Apache-2.0

set -e

export ORIGIN="$(pwd)"
export FABRIC_SAMPLES_DIR="${FABRIC_SAMPLES_DIR:-../fabric-samples}"
export TEST_NETWORK_DIR="${FABRIC_SAMPLES_DIR}/test-network"
export NETWORK_CMD="${TEST_NETWORK_DIR}/network.sh"
export CHAINCODE="$1"
export CC_INSTANCE="${CHAINCODE}-$RANDOM"

if [ -z "${CHAINCODE}" ]; then
  echo "Usage: $0 <chaincode>|down"
  echo "<chaincode> must be 'assetholder' or 'adjudicator'"
  echo "down just shuts down the network"
fi

function ensureNetworkUp() {
  if [ -d "${TEST_NETWORK_DIR}/organizations/peerOrganizations" ]; then
    echo "Test network seems to be running."
    return
  fi

  echo "Starting test network..."
  "${NETWORK_CMD}" up createChannel -c mychannel -ca
}

function networkDown() {
  "${NETWORK_CMD}" down
}

function deployCC() {
  "${NETWORK_CMD}" deployCC -ccn $CC_INSTANCE -ccp "${ORIGIN}/chaincode/$CHAINCODE/" -ccl go
}

function runTest() {
  cd "${ORIGIN}"
  go run "./tests/${CHAINCODE}" -chaincode $CC_INSTANCE
}

function e2e() {
  ensureNetworkUp
  deployCC
  runTest
}

[[ "$CHAINCODE" == "assetholder" || "$CHAINCODE" == "adjudicator" ]] && e2e

[[ "$CHAINCODE" == "down" ]] && networkDown

