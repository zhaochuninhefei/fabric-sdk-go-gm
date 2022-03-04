/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"fmt"
	"path/filepath"
	"testing"

	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/fab"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/msp"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/core/config"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/core/cryptosuite/bccsp/sw"
	coreMocks "gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/core/mocks"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/test/metadata"

	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/core/cryptosuite"
	fabImpl "gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/fab"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/fab/mocks"
	peerImpl "gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/fab/peer"
	mspImpl "gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/msp"
	mspmocks "gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/msp/test/mockmsp"
)

func TestCreateInfraProvider(t *testing.T) {
	newInfraProvider(t)
}

func verifyPeer(t *testing.T, peer fab.Peer, url string) {
	_, ok := peer.(*peerImpl.Peer)
	if !ok {
		t.Fatal("Unexpected peer impl created")
	}

	// Brittle tests follow
	a := peer.URL()

	if a != url {
		t.Fatalf("Unexpected URL %s", a)
	}
}

func TestCreatePeerFromConfig(t *testing.T) {
	p := newInfraProvider(t)

	url := "grpc://localhost:9999"

	peerCfg := fab.NetworkPeer{
		PeerConfig: fab.PeerConfig{
			URL: url,
		},
	}

	peer, err := p.CreatePeerFromConfig(&peerCfg)

	if err != nil {
		t.Fatalf("Unexpected error creating peer %s", err)
	}

	verifyPeer(t, peer, url)
}

func newInfraProvider(t *testing.T) *InfraProvider {
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, "config_test.yaml")
	configBackend, err := config.FromFile(configPath)()
	if err != nil {
		t.Fatalf("config.FromFile failed: %s", err)
	}

	cryptoCfg := cryptosuite.ConfigFromBackend(configBackend...)
	if err != nil {
		t.Fatalf(err.Error())
	}

	endpointCfg, err := fabImpl.ConfigFromBackend(configBackend...)
	if err != nil {
		t.Fatalf(err.Error())
	}

	identityCfg, err := mspImpl.ConfigFromBackend(configBackend...)
	if err != nil {
		t.Fatalf(err.Error())
	}

	cryptoSuite, err := sw.GetSuiteByConfig(cryptoCfg)
	if err != nil {
		panic(fmt.Sprintf("cryptosuiteimpl.GetSuiteByConfig: %s", err))
	}
	im := make(map[string]msp.IdentityManager)
	im[""] = &mocks.MockIdentityManager{}

	ctx := mocks.NewMockProviderContextCustom(cryptoCfg, endpointCfg, identityCfg, cryptoSuite, coreMocks.NewMockSigningManager(), &mspmocks.MockUserStore{}, im)
	ip := New(endpointCfg)
	ip.Initialize(ctx)

	return ip
}
