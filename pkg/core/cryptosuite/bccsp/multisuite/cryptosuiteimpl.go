/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package multisuite

import (
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/core"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/core/cryptosuite/bccsp/pkcs11"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/core/cryptosuite/bccsp/sw"
	"github.com/pkg/errors"
)

//GetSuiteByConfig returns cryptosuite adaptor for bccsp loaded according to given config
func GetSuiteByConfig(config core.CryptoSuiteConfig) (core.CryptoSuite, error) {
	switch config.SecurityProvider() {
	case "sw":
		return sw.GetSuiteByConfig(config)
	case "pkcs11":
		return pkcs11.GetSuiteByConfig(config)
	}

	return nil, errors.Errorf("Unsupported security provider requested: %s", config.SecurityProvider())
}
