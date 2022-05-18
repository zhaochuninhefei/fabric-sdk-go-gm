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
	// ECDSA            = bccsp.ECDSA
	// ECDSAP256        = bccsp.ECDSAP256
	// ECDSAP384        = bccsp.ECDSAP384
	// ECDSAReRand      = bccsp.ECDSAReRand
	// AES              = bccsp.AES
	// AES128           = bccsp.AES128
	// AES192           = bccsp.AES192
	// AES256           = bccsp.AES256
	// HMAC             = bccsp.HMAC
	// HMACTruncated256 = bccsp.HMACTruncated256
	// SHA              = bccsp.SHA
	// SHA2             = bccsp.SHA2
	// SHA3             = bccsp.SHA3
	// SHA256           = bccsp.SHA256
	// SHA384           = bccsp.SHA384
	// SHA3_256         = bccsp.SHA3_256
	// SHA3_384         = bccsp.SHA3_384
	// X509Certificate  = bccsp.X509Certificate
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

// //GetSHAOpts returns options for computing SHA.
// func GetSHAOpts() core.HashOpts {
// 	return &bccsp.SHAOpts{}
// }

// //GetSHA256Opts returns options relating to SHA-256.
// func GetSHA256Opts() core.HashOpts {
// 	return &bccsp.SHA256Opts{}
// }

// // GetECDSAKeyGenOpts returns options for ECDSA key generation.
// func GetECDSAKeyGenOpts(ephemeral bool) core.KeyGenOpts {
// 	return &bccsp.ECDSAKeyGenOpts{Temporary: ephemeral}
// }

// //GetECDSAP256KeyGenOpts returns options for ECDSA key generation with curve P-256.
// func GetECDSAP256KeyGenOpts(ephemeral bool) core.KeyGenOpts {
// 	return &bccsp.ECDSAP256KeyGenOpts{Temporary: ephemeral}
// }

// //GetECDSAP384KeyGenOpts options for ECDSA key generation with curve P-384.
// func GetECDSAP384KeyGenOpts(ephemeral bool) core.KeyGenOpts {
// 	return &bccsp.ECDSAP384KeyGenOpts{Temporary: ephemeral}
// }

// //GetX509PublicKeyImportOpts options for importing public keys from an x509 certificate
// func GetX509PublicKeyImportOpts(ephemeral bool) core.KeyImportOpts {
// 	return &bccsp.X509PublicKeyImportOpts{Temporary: ephemeral}
// }

// //GetECDSAPrivateKeyImportOpts options for ECDSA secret key importation in DER format
// // or PKCS#8 format.
// func GetECDSAPrivateKeyImportOpts(ephemeral bool) core.KeyImportOpts {
// 	return &bccsp.ECDSAPrivateKeyImportOpts{Temporary: ephemeral}
// }
