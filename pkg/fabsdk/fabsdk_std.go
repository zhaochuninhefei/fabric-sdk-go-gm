//go:build !pprof
// +build !pprof

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabsdk enables client usage of a Hyperledger Fabric network.
package fabsdk

import (
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/fabsdk/metrics"
)

func (sdk *FabricSDK) initMetrics(config *configs) {

	sdk.clientMetrics = &metrics.ClientMetrics{} // empty channel ClientMetrics for standard build.
}
