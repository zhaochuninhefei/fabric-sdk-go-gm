/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"
	"os"

	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/context"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/core"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/fab"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/msp"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/core/config"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/core/cryptosuite"
	fabImpl "gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/fab"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/fabsdk"
	mspImpl "gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/msp"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/util/pathvar"
	"gitee.com/zhaochuninhefei/gmgo/grpc/testdata"
)

const (
	configPath = "${FABRIC_SDK_GO_PROJECT_PATH}/pkg/core/config/testdata/config_test.yaml"
)

type testFixture struct {
	cryptoSuiteConfig core.CryptoSuiteConfig
	identityConfig    msp.IdentityConfig
	endpointConfig    fab.EndpointConfig
}

func (f *testFixture) setup() (*fabsdk.FabricSDK, context.Client) {
	var err error

	backend, err := config.FromFile(testdata.Path(pathvar.Subst(configPath)))()
	if err != nil {
		panic(err)
	}

	customBackend := backend

	configProvider := func() ([]core.ConfigBackend, error) {
		return customBackend, nil
	}

	// Instantiate the SDK
	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		panic(fmt.Sprintf("SDK init failed: %s", err))
	}

	configBackend, err := sdk.Config()
	if err != nil {
		panic(fmt.Sprintf("Failed to get config: %s", err))
	}

	// set cryptoSuiteConfig
	f.cryptoSuiteConfig = cryptosuite.ConfigFromBackend(configBackend)

	// set identityConfig
	f.identityConfig, err = mspImpl.ConfigFromBackend(configBackend)
	if err != nil {
		panic(fmt.Sprintf("Failed to get identity config: %s", err))
	}

	// set endpointConfig
	f.endpointConfig, err = fabImpl.ConfigFromBackend(configBackend)
	if err != nil {
		panic(fmt.Sprintf("Failed to get endpoint config: %s", err))
	}

	// Delete all private keys from the crypto suite store
	// and users from the user store
	cleanup(f.cryptoSuiteConfig.KeyStorePath())
	cleanup(f.identityConfig.CredentialStorePath())

	// create a context with a real user/org found in the configs
	ctxProvider := sdk.Context(fabsdk.WithUser("User1"), fabsdk.WithOrg("Org1"))
	ctx, err := ctxProvider()
	if err != nil {
		panic(fmt.Sprintf("Failed to init context: %s", err))
	}

	return sdk, ctx
}

func (f *testFixture) close() {
	cleanup(f.identityConfig.CredentialStorePath())
	cleanup(f.cryptoSuiteConfig.KeyStorePath())
}

func cleanup(storePath string) {
	err := os.RemoveAll(storePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to remove dir %s: %s\n", storePath, err))
	}
}
