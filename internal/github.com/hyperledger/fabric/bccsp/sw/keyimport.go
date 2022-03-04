/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package sw

import (
	"crypto/ecdsa"
	// "crypto/x509"
	"errors"
	"fmt"
	"reflect"

	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/internal/github.com/hyperledger/fabric/bccsp/utils"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/internal/github.com/hyperledger/fabric/bccsp"
	"gitee.com/zhaochuninhefei/gmgo/sm2"
	"gitee.com/zhaochuninhefei/gmgo/x509"
)

type aes256ImportKeyOptsKeyImporter struct{}

func (*aes256ImportKeyOptsKeyImporter) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (bccsp.Key, error) {
	aesRaw, ok := raw.([]byte)
	if !ok {
		return nil, errors.New("Invalid raw material. Expected byte array.")
	}

	if aesRaw == nil {
		return nil, errors.New("Invalid raw material. It must not be nil.")
	}

	if len(aesRaw) != 32 {
		return nil, fmt.Errorf("Invalid Key Length [%d]. Must be 32 bytes", len(aesRaw))
	}

	return &aesPrivateKey{aesRaw, false}, nil
}

type hmacImportKeyOptsKeyImporter struct{}

func (*hmacImportKeyOptsKeyImporter) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (bccsp.Key, error) {
	aesRaw, ok := raw.([]byte)
	if !ok {
		return nil, errors.New("Invalid raw material. Expected byte array.")
	}

	if len(aesRaw) == 0 {
		return nil, errors.New("Invalid raw material. It must not be nil.")
	}

	return &aesPrivateKey{aesRaw, false}, nil
}

type ecdsaPKIXPublicKeyImportOptsKeyImporter struct{}

func (*ecdsaPKIXPublicKeyImportOptsKeyImporter) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (bccsp.Key, error) {
	der, ok := raw.([]byte)
	if !ok {
		return nil, errors.New("Invalid raw material. Expected byte array.")
	}

	if len(der) == 0 {
		return nil, errors.New("Invalid raw. It must not be nil.")
	}

	lowLevelKey, err := derToPublicKey(der)
	if err != nil {
		return nil, fmt.Errorf("Failed converting PKIX to ECDSA public key [%s]", err)
	}

	ecdsaPK, ok := lowLevelKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("Failed casting to ECDSA public key. Invalid raw material.")
	}

	return &ecdsaPublicKey{ecdsaPK}, nil
}

type ecdsaPrivateKeyImportOptsKeyImporter struct{}

func (*ecdsaPrivateKeyImportOptsKeyImporter) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (bccsp.Key, error) {
	der, ok := raw.([]byte)
	if !ok {
		return nil, errors.New("[ECDSADERPrivateKeyImportOpts] Invalid raw material. Expected byte array.")
	}

	if len(der) == 0 {
		return nil, errors.New("[ECDSADERPrivateKeyImportOpts] Invalid raw. It must not be nil.")
	}

	lowLevelKey, err := derToPrivateKey(der)
	if err != nil {
		return nil, fmt.Errorf("Failed converting PKIX to ECDSA public key [%s]", err)
	}

	ecdsaSK, ok := lowLevelKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("Failed casting to ECDSA private key. Invalid raw material.")
	}

	return &ecdsaPrivateKey{ecdsaSK, true}, nil
}

type ecdsaGoPublicKeyImportOptsKeyImporter struct{}

func (*ecdsaGoPublicKeyImportOptsKeyImporter) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (bccsp.Key, error) {
	lowLevelKey, ok := raw.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("Invalid raw material. Expected *ecdsa.PublicKey.")
	}

	return &ecdsaPublicKey{lowLevelKey}, nil
}

type x509PublicKeyImportOptsKeyImporter struct {
	bccsp *CSP
}

func (ki *x509PublicKeyImportOptsKeyImporter) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (bccsp.Key, error) {
	sm2Cert, ok := raw.(*x509.Certificate)
	if !ok {
		return nil, errors.New("Invalid raw material. Expected *x509.Certificate.")
	}

	pk := sm2Cert.PublicKey

	switch pk.(type) {
	case *sm2.PublicKey:
		fmt.Printf("")
		sm2PublicKey, ok := pk.(sm2.PublicKey)
		if !ok {
			return nil, errors.New("Parse interface [] to sm2 pk error")
		}
		der, err := x509.MarshalSm2PublicKey(&sm2PublicKey)
		if err != nil {
			return nil, errors.New("MarshalSm2PublicKey error")
		}
		return ki.bccsp.KeyImporters[reflect.TypeOf(&bccsp.GMSM2PublicKeyImportOpts{})].KeyImport(
			der,
			&bccsp.GMSM2PublicKeyImportOpts{Temporary: opts.Ephemeral()})
	case *ecdsa.PublicKey:
		return ki.bccsp.KeyImporters[reflect.TypeOf(&bccsp.ECDSAGoPublicKeyImportOpts{})].KeyImport(
			pk,
			&bccsp.ECDSAGoPublicKeyImportOpts{Temporary: opts.Ephemeral()})
	// case *rsa.PublicKey:
	// 	// This path only exists to support environments that use RSA certificate
	// 	// authorities to issue ECDSA certificates.
	// 	return &rsaPublicKey{pubKey: pk}, nil
	default:
		return nil, errors.New("Certificate's public key type not recognized. Supported keys: [SM2, ECDSA]")
	}
}

type gmsm4ImportKeyOptsKeyImporter struct{}

func (*gmsm4ImportKeyOptsKeyImporter) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (k bccsp.Key, err error) {
	sm4Raw, ok := raw.([]byte)
	if !ok {
		return nil, errors.New("Invalid raw material, Expected byte array")
	}

	if sm4Raw == nil {
		return nil, errors.New("Invalid raw material, It must botbe nil")
	}

	return &gmsm4PrivateKey{utils.Clone(sm4Raw), false}, nil
}

type gmsm2PrivateKeyOptsKeyImporter struct{}

func (*gmsm2PrivateKeyOptsKeyImporter) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (k bccsp.Key, err error) {
	der, ok := raw.([]byte)
	if !ok {
		return nil, errors.New("Invalid raw material, Expected byte array")
	}

	if len(der) == 0 {
		return nil, errors.New("Invalid raw material, It must botbe nil")
	}

	gmsm2SK, err := x509.ParsePKCS8UnecryptedPrivateKey(der)

	if err != nil {
		return nil, fmt.Errorf("Failed converting to SM2 private key [%s]", err)
	}

	return &gmsm2PrivateKey{gmsm2SK}, nil
}

type gmsm2PublicKeyOptsKeyImporter struct{}

func (*gmsm2PublicKeyOptsKeyImporter) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (k bccsp.Key, err error) {
	der, ok := raw.([]byte)
	if !ok {
		return nil, errors.New("Invalid raw material, Expected byte array")
	}

	if len(der) == 0 {
		return nil, errors.New("Invalid raw material, It must botbe nil")
	}

	gmsm2SK, err := x509.ParseSm2PublicKey(der)

	if err != nil {
		return nil, fmt.Errorf("Failed converting to SM2 private key [%s]", err)
	}

	return &gmsm2PublicKey{gmsm2SK}, nil
}
