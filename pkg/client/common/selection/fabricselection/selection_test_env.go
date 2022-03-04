//go:build testing
// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricselection

import (
	contextAPI "gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/context"
)

// SetClientProvider overrides the discovery client provider for unit tests
func SetClientProvider(provider func(ctx contextAPI.Client) (DiscoveryClient, error)) {
	clientProvider = provider
}
