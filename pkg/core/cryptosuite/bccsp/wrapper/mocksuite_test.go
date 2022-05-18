/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wrapper

import (
	"fmt"

	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/internal/gitee.com/zhaochuninhefei/fabric-gm/bccsp"
	bccspSw "gitee.com/zhaochuninhefei/fabric-sdk-go-gm/internal/gitee.com/zhaochuninhefei/fabric-gm/bccsp/factory"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/core"
)

//getSuiteByConfig returns cryptosuite adaptor for bccsp loaded according to given config
func getSuiteByConfig(config core.CryptoSuiteConfig) (core.CryptoSuite, error) {
	opts := getOptsByConfig(config)
	bccsp, err := getBCCSPFromOpts(opts)

	if err != nil {
		return nil, err
	}
	return &CryptoSuite{BCCSP: bccsp}, nil
}

func getBCCSPFromOpts(opts *bccspSw.SwOpts) (bccsp.BCCSP, error) {
	f := &bccspSw.SWFactory{}
	config := &bccspSw.FactoryOpts{
		SwOpts: opts,
	}
	return f.Get(config)
}

//getOptsByConfig Returns Factory opts for given SDK config
func getOptsByConfig(c core.CryptoSuiteConfig) *bccspSw.SwOpts {
	// TODO: delete this check
	if c.SecurityProvider() != "SW" {
		panic(fmt.Sprintf("Unsupported BCCSP Provider: %s", c.SecurityProvider()))
	}

	opts := &bccspSw.SwOpts{
		HashFamily: c.SecurityAlgorithm(),
		SecLevel:   c.SecurityLevel(),
		FileKeystore: &bccspSw.FileKeystoreOpts{
			KeyStorePath: c.KeyStorePath(),
		},
	}

	return opts
}
