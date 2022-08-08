#!/bin/bash
#
# SPDX-License-Identifier: Apache-2.0

set -e

export ORIGIN="$(pwd)"
export FABRIC_SAMPLES_DIR="${FABRIC_SAMPLES_DIR:-../fabric-samples}"
export TEST_NETWORK_DIR="${FABRIC_SAMPLES_DIR}/test-network"
export NETWORK_CMD="${TEST_NETWORK_DIR}/network.sh"

function ensureNetworkUp() {
  if [ -d "${TEST_NETWORK_DIR}/organizations/peerOrganizations" ]; then
    echo "Test network seems to be running."
    return
  fi

  echo "Starting test network..."
  "${NETWORK_CMD}" up createChannel -c mychannel -ca
}

ensureNetworkUp

# Deploy chaincode
"${NETWORK_CMD}" deployCC -ccn "assetholder" -ccp "${ORIGIN}/chaincode/assetholder/" -ccl go
"${NETWORK_CMD}" deployCC -ccn "adjudicator" -ccp "${ORIGIN}/chaincode/adjudicator/" -ccl go

# Environment variables for submitting transactions
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_ORG1_TLS_ROOTCERT_FILE="${TEST_NETWORK_DIR}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
export CORE_PEER_ORG2_TLS_ROOTCERT_FILE="${TEST_NETWORK_DIR}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"
export CORE_ORDERERS="../${TEST_NETWORK_DIR}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem"
export CORE_PEER_MSPCONFIGPATH="../${TEST_NETWORK_DIR}/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp"
export CORE_PEER_ADDRESS=localhost:7051

# Set up peer command
export PATH=${TEST_NETWORK_DIR}/../bin:$PATH
export FABRIC_CFG_PATH=${TEST_NETWORK_DIR}/../config/
export PEER_CMD="peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${CORE_ORDERERS} -C mychannel -n adjudicator --peerAddresses localhost:7051 --tlsRootCertFiles ${CORE_PEER_ORG1_TLS_ROOTCERT_FILE} --peerAddresses localhost:9051 --tlsRootCertFiles ${CORE_PEER_ORG2_TLS_ROOTCERT_FILE}"

# Mint tokens
${PEER_CMD} -c '{"function":"MintToken","Args":["2000000000000"]}'
sleep 3
# Transfer half to other party
${PEER_CMD} -c '{"function":"TransferToken","Args":["\"eDUwOTo6Q049dXNlcjEsT1U9Y2xpZW50LE89SHlwZXJsZWRnZXIsU1Q9Tm9ydGggQ2Fyb2xpbmEsQz1VUzo6Q049Y2Eub3JnMi5leGFtcGxlLmNvbSxPPW9yZzIuZXhhbXBsZS5jb20sTD1IdXJzbGV5LFNUPUhhbXBzaGlyZSxDPVVL\"", "1000000000000"]}'