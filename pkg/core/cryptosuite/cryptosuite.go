/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cryptosuite

import (
	"sync/atomic"

	"errors"

	"sync"

	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/internal/gitee.com/zhaochuninhefei/fabric-gm/bccsp"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/logging"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/core"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/core/cryptosuite/bccsp/sw"
)

var logger = logging.NewLogger("fabsdk/core")

var initOnce sync.Once
var defaultCryptoSuite core.CryptoSuite
var initialized int32

func initSuite(defaultSuite core.CryptoSuite) error {
	if defaultSuite == nil {
		return errors.New("attempting to set invalid default suite")
	}
	initOnce.Do(func() {
		defaultCryptoSuite = defaultSuite
		atomic.StoreInt32(&initialized, 1)
	})
	return nil
}

//GetDefault returns default core
func GetDefault() core.CryptoSuite {
	if atomic.LoadInt32(&initialized) > 0 {
		return defaultCryptoSuite
	}
	//Set default suite
	logger.Info("No default cryptosuite found, using default SW implementation")

	// Use SW as the default cryptosuite when not initialized properly - should be for testing only
	s, err := sw.GetSuiteWithDefaultEphemeral()
	if err != nil {
		logger.Panicf("Could not initialize default cryptosuite: %s", err)
	}
	err = initSuite(s)
	if err != nil {
		logger.Panicf("Could not set default cryptosuite: %s", err)
	}

	return defaultCryptoSuite
}

//SetDefault sets default suite if one is not already set or created
//Make sure you set default suite before very first call to GetDefault(),
//otherwise this function will return an error
func SetDefault(newDefaultSuite core.CryptoSuite) error {
	if atomic.LoadInt32(&initialized) > 0 {
		return errors.New("default crypto suite is already set")
	}
	return initSuite(newDefaultSuite)
}

// DefaultInitialized returns true if a default suite has already been
// set.
func DefaultInitialized() bool {
	return atomic.LoadInt32(&initialized) > 0
}

//GetSM3Opts returns options for computing SM3.
func GetSM3Opts() core.HashOpts {
	return &bccsp.SM3Opts{}
}

//GetSM2KeyGenOpts returns options for SM2 key generation with curve SM2-P-256.
func GetSM2KeyGenOpts(ephemeral bool) core.KeyGenOpts {
	return &bccsp.SM2KeyGenOpts{Temporary: ephemeral}
}

//GetSM2PrivateKeyImportOpts options for SM2 secret key importation in DER format
// or PKCS#8 format.
func GetSM2PrivateKeyImportOpts(ephemeral bool) core.KeyImportOpts {
	return &bccsp.SM2PrivateKeyImportOpts{Temporary: ephemeral}
}

// //GetSHA256Opts returns options relating to SHA-256.
// func GetSHA256Opts() core.HashOpts {
// 	return &bccsp.SHA256Opts{}
// }

// //GetSHAOpts returns options for computing SHA.
// func GetSHAOpts() core.HashOpts {
// 	return &bccsp.SHAOpts{}
// }

// //GetECDSAP256KeyGenOpts returns options for ECDSA key generation with curve P-256.
// func GetECDSAP256KeyGenOpts(ephemeral bool) core.KeyGenOpts {
// 	return &bccsp.ECDSAP256KeyGenOpts{Temporary: ephemeral}
// }

// //GetECDSAPrivateKeyImportOpts options for ECDSA secret key importation in DER format
// // or PKCS#8 format.
// func GetECDSAPrivateKeyImportOpts(ephemeral bool) core.KeyImportOpts {
// 	return &bccsp.ECDSAPrivateKeyImportOpts{Temporary: ephemeral}
// }
