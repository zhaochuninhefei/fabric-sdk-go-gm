/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"gitee.com/zhaochuninhefei/fabric-chaincode-go-gm/shim"
	pb "gitee.com/zhaochuninhefei/fabric-protos-go-gm/peer"
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
