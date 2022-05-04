#!/bin/bash
#
# SPDX-License-Identifier: Apache-2.0

set -e

export ORIGIN="$(pwd)"
export FABRIC_SAMPLES_DIR="${FABRIC_SAMPLES_DIR:-fabric-samples}"
export TEST_NETWORK_DIR="${FABRIC_SAMPLES_DIR}/test-network"
export NETWORK_CMD="${TEST_NETWORK_DIR}/network.sh"
export CRL="$1"

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

case "$CRL" in
  up)
    ensureNetworkUp ;;
  down)
    networkDown ;;
  *)
    echo "Usage: $0 up|down"
    ;;
esac
