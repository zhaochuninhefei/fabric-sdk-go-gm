/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package cryptosuitebridge

import (
	"crypto"

	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/internal/gitee.com/zhaochuninhefei/fabric-ca-gm/sdkpatch/keyutil"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/internal/gitee.com/zhaochuninhefei/fabric-gm/bccsp"
	cspsigner "gitee.com/zhaochuninhefei/fabric-sdk-go-gm/internal/gitee.com/zhaochuninhefei/fabric-gm/bccsp/signer"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/core"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/core/cryptosuite"
	"gitee.com/zhaochuninhefei/gmgo/sm2"
)

const (
	SM2               = bccsp.SM2
	SM3               = bccsp.SM3
	SM4               = bccsp.SM4
	GMX509Certificate = bccsp.GMX509Certificate
)

// NewCspSigner is a bridge for bccsp signer.New call
func NewCspSigner(csp core.CryptoSuite, key core.Key) (crypto.Signer, error) {
	return cspsigner.New(csp, key)
}

// PEMtoPrivateKey is a bridge for bccsp utils.PEMtoPrivateKey()
//
// PEM字节数组转为私钥
func PEMtoPrivateKey(raw []byte, pwd []byte) (interface{}, error) {
	return keyutil.PEMToPrivateKey(raw, pwd)
}

// PrivateKeyToDER marshals is bridge for utils.PrivateKeyToDER
func PrivateKeyToDER(privateKey *sm2.PrivateKey) ([]byte, error) {
	return keyutil.PrivateKeyToDER(privateKey)
}

//GetDefault returns default cryptosuite from bccsp factory default
func GetDefault() core.CryptoSuite {
	return cryptosuite.GetDefault()
}

//GetSM3Opts returns options for computing SM3.
func GetSM3Opts() core.HashOpts {
	return &bccsp.SM3Opts{}
}

// GetSM2KeyGenOpts returns options for SM2 key generation.
func GetSM2KeyGenOpts(ephemeral bool) core.KeyGenOpts {
	return &bccsp.SM2KeyGenOpts{Temporary: ephemeral}
}

//GetGMX509PublicKeyImportOpts options for importing public keys from an gmx509 certificate
func GetGMX509PublicKeyImportOpts(ephemeral bool) core.KeyImportOpts {
	return &bccsp.GMX509PublicKeyImportOpts{Temporary: ephemeral}
}

//GetSM2PrivateKeyImportOpts options for ECDSA secret key importation in DER format
// or PKCS#8 format.
func GetSM2PrivateKeyImportOpts(ephemeral bool) core.KeyImportOpts {
	return &bccsp.SM2PrivateKeyImportOpts{Temporary: ephemeral}
}
