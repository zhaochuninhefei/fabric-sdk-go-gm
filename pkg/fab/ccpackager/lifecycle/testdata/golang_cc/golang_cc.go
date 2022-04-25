/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	pb "gitee.com/zhaochuninhefei/fabric-protos-go-gm/peer"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

// GolangCC is a sample chaincode written in Go
type GolangCC struct {
}

// Init initializes the chaincode
func (cc *GolangCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke invokes the chaincode
func (cc *GolangCC) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(GolangCC))
	if err != nil {
		panic(err)
	}
}
