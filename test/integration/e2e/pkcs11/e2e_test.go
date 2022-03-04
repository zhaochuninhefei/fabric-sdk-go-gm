/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pkcs11

import (
	"testing"

	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/fabsdk"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/test/integration"

	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/core"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/core/config"
	cryptosuite "gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/core/cryptosuite/bccsp/pkcs11"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/fabsdk/factory/defcore"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/test/integration/e2e"
)

const (
	// ConfigTestFile contains the path and filename of the config for integration tests
	ConfigTestFilename = "config_e2e_pkcs11.yaml"
)

func TestE2E(t *testing.T) {
	// Create SDK setup for the integration tests
	e2e.Run(t,
		config.FromFile(integration.GetConfigPath(ConfigTestFilename)),
		fabsdk.WithCorePkg(&CustomCryptoSuiteProviderFactory{}))
}

// CustomCryptoSuiteProviderFactory is will provide custom cryptosuite (bccsp.BCCSP)
type CustomCryptoSuiteProviderFactory struct {
	defcore.ProviderFactory
}

// CreateCryptoSuiteProvider returns a new default implementation of BCCSP
func (f *CustomCryptoSuiteProviderFactory) CreateCryptoSuiteProvider(config core.CryptoSuiteConfig) (core.CryptoSuite, error) {
	return cryptosuite.GetSuiteByConfig(config)
}
