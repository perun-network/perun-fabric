package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"

	"github.com/perun-network/perun-fabric/chaincode"
)

func main() {
	cc, err := contractapi.NewChaincode(&chaincode.AssetHolder{})
	if err != nil {
		log.Panicf("Error creating AssetHolder chaincode: %v", err)
	}

	if err := cc.Start(); err != nil {
		log.Panicf("Error starting AssetHolder chaincode: %v", err)
	}
}
