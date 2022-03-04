/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defsvc

import (
	"testing"

	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/client/common/discovery/staticdiscovery"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/fab/mocks"
)

func TestCreateLocalDiscoveryProvider(t *testing.T) {
	factory := NewProviderFactory()

	config := mocks.NewMockEndpointConfig()

	dp, err := factory.CreateLocalDiscoveryProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating local discovery provider %s", err)
	}

	_, ok := dp.(*staticdiscovery.LocalProvider)
	if !ok {
		t.Fatal("Unexpected local discovery provider created")
	}
}
