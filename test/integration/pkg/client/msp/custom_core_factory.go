/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/core"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/fab"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/fabsdk/factory/defcore"
)

// ========== Core Provider Factory with custom crypto provider ============= //

// CustomCoreFactory is a custom factory for tests.
type CustomCoreFactory struct {
	defaultFactory    *defcore.ProviderFactory
	customCryptoSuite core.CryptoSuite
}

// NewCustomCoreFactory creates a new instance of customCoreFactory
func NewCustomCoreFactory(customCryptoSuite core.CryptoSuite) *CustomCoreFactory {
	return &CustomCoreFactory{defaultFactory: defcore.NewProviderFactory(), customCryptoSuite: customCryptoSuite}
}

// CreateCryptoSuiteProvider creates custom crypto provider
func (f *CustomCoreFactory) CreateCryptoSuiteProvider(cryptoSuiteConfig core.CryptoSuiteConfig) (core.CryptoSuite, error) {
	return f.customCryptoSuite, nil
}

// CreateSigningManager creates SigningManager
func (f *CustomCoreFactory) CreateSigningManager(cryptoProvider core.CryptoSuite) (core.SigningManager, error) {
	return f.defaultFactory.CreateSigningManager(cryptoProvider)
}

// CreateInfraProvider creates InfraProvider
func (f *CustomCoreFactory) CreateInfraProvider(config fab.EndpointConfig) (fab.InfraProvider, error) {
	return f.defaultFactory.CreateInfraProvider(config)
}
