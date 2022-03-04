/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package fab

import (
	"testing"

	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/fabsdk"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/test/integration"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/test/integration/util/runner"
)

var mainSDK *fabsdk.FabricSDK
var mainTestSetup *integration.BaseSetupImpl
var mainChaincodeID string

const (
	org1Name = "Org1"
	org1User = "User1"
)

func TestMain(m *testing.M) {
	r := runner.NewWithExampleCC()
	r.Initialize()
	mainSDK = r.SDK()
	mainTestSetup = r.TestSetup()
	mainChaincodeID = r.ExampleChaincodeID()

	r.Run(m)
}
