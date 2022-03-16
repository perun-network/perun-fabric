package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"

	"github.com/perun-network/perun-fabric/chaincode"
)

func main() {
	cc, err := contractapi.NewChaincode(new(chaincode.Adjudicator))
	if err != nil {
		log.Panicf("Error creating Adjudicator chaincode: %v", err)
	}

	if err := cc.Start(); err != nil {
		log.Panicf("Error starting Adjudicator chaincode: %v", err)
	}
}
