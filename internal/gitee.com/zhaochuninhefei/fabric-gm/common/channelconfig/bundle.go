/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package channelconfig

import (
	flogging "gitee.com/zhaochuninhefei/fabric-sdk-go-gm/internal/gitee.com/zhaochuninhefei/fabric-gm/sdkpatch/logbridge"
)

var logger = flogging.MustGetLogger("common.channelconfig")

// RootGroupKey is the key for namespacing the channel config, especially for
// policy evaluation.
const RootGroupKey = "Channel"
