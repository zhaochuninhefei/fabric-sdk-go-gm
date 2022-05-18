/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package cryptosuite

import (
	"sync/atomic"
	"testing"

	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/core/cryptosuite/bccsp/sw"
	"github.com/stretchr/testify/assert"
)

const (
	sm3HashOptsAlgorithm       = "SM3"
	sm2KeyGenOpts              = "SM2"
	sm2ImportOpts              = "SM2"
	setDefAlreadySetErrorMsg   = "default crypto suite is already set"
	InvalidDefSuiteSetErrorMsg = "attempting to set invalid default suite"
)

func TestGetDefault(t *testing.T) {

	//At the beginning default suite is nil if no attempts have been made to set or get one
	assert.Empty(t, defaultCryptoSuite, "default suite should be nil if no attempts have been made to set or get one")

	//Now try to get default, it will create one and return
	defSuite := GetDefault()
	assert.NotEmpty(t, defSuite, "Not supposed to be nil defaultCryptSuite")
	assert.NotEmpty(t, defaultCryptoSuite, "default suite should have been initialized")
	assert.True(t, atomic.LoadInt32(&initialized) > 0, "'initialized' flag supposed to be set to 1")

	hashbytes, err := defSuite.Hash([]byte("Sample message"), GetSM3Opts())
	assert.Empty(t, err, "Not supposed to get error on defaultCryptSuite.Hash() call : %s", err)
	assert.NotEmpty(t, hashbytes, "Supposed to get valid hash from defaultCryptSuite.Hash()")

	//Now try to get default, which is already created
	defSuite = GetDefault()
	assert.NotEmpty(t, defSuite, "Not supposed to be nil defaultCryptSuite")
	assert.NotEmpty(t, defaultCryptoSuite, "default suite should have been initialized")
	assert.True(t, atomic.LoadInt32(&initialized) > 0, "'initialized' flag supposed to be set to 1")

	hashbytes, err = defSuite.Hash([]byte("Sample message"), GetSM3Opts())
	assert.Empty(t, err, "Not supposed to get error on defaultCryptSuite.Hash() call : %s", err)
	assert.NotEmpty(t, hashbytes, "Supposed to get valid hash from defaultCryptSuite.Hash()")

	//Now attempt to set default suite
	err = SetDefault(nil)
	assert.NotEmpty(t, err, "supposed to get error when SetDefault() gets called after GetDefault()")
	assert.True(t, err.Error() == setDefAlreadySetErrorMsg, "unexpected error : expected [%s], got [%s]", setDefAlreadySetErrorMsg, err.Error())

	//Reset
	defaultCryptoSuite = nil
	atomic.StoreInt32(&initialized, 0)

	//Now attempt to set invalid default suite
	err = SetDefault(nil)
	assert.NotEmpty(t, err, "supposed to get error when invalid default suite is set")
	assert.True(t, err.Error() == InvalidDefSuiteSetErrorMsg, "unexpected error : expected [%s], got [%s]", InvalidDefSuiteSetErrorMsg, err.Error())

	s, err := sw.GetSuiteWithDefaultEphemeral()
	if err != nil {
		t.Fatal("Unable to get default cryptosuite")
	}

	err = SetDefault(s)
	assert.Empty(t, err, "Not supposed to get error when valid default suite is set")

}

func TestHashOpts(t *testing.T) {

	//Get CryptoSuite SM3 Opts
	hashOpts := GetSM3Opts()
	assert.NotZero(t, hashOpts, "Not supposed to be empty sm3HashOpts")
	assert.True(t, hashOpts.Algorithm() == sm3HashOptsAlgorithm, "Unexpected SHA hash opts, expected [%s], got [%s]", sm3HashOptsAlgorithm, hashOpts.Algorithm())

}

func TestKeyGenOpts(t *testing.T) {

	keygenOpts := GetSM2KeyGenOpts(true)
	assert.NotEmpty(t, keygenOpts, "Not supposed to be empty SM2KeyGenOpts")
	assert.True(t, keygenOpts.Ephemeral(), "Expected keygenOpts.Ephemeral() ==> true")
	assert.True(t, keygenOpts.Algorithm() == sm2KeyGenOpts, "Unexpected ECDSA KeyGen opts, expected [%v], got [%v]", sm2KeyGenOpts, keygenOpts.Algorithm())

	// keygenOpts = GetECDSAP256KeyGenOpts(false)
	// assert.NotZero(t, keygenOpts, "Not supposed to be empty ECDSAP256KeyGenOpts")
	// assert.False(t, keygenOpts.Ephemeral(), "Expected keygenOpts.Ephemeral() ==> false")
	// assert.True(t, keygenOpts.Algorithm() == ecdsap256KeyGenOpts, "Unexpected ECDSA KeyGen opts, expected [%v], got [%v]", ecdsap256KeyGenOpts, keygenOpts.Algorithm())

}

func TestKeyImportOpts(t *testing.T) {
	importOpts := GetSM2PrivateKeyImportOpts(true)
	assert.NotEmpty(t, importOpts, "Not supposed to be empty SM2PrivateKeyImportOpts")
	assert.True(t, importOpts.Ephemeral(), "Expected keygenOpts.Ephemeral() ==> true")
	assert.True(t, importOpts.Algorithm() == sm2ImportOpts, "UUnexpected ECDSA import opts, expected [%v], got [%v]", sm2ImportOpts, importOpts.Algorithm())

	// importOpts = GetECDSAPrivateKeyImportOpts(false)
	// assert.NotZero(t, importOpts, "Not supposed to be empty ECDSAP256KeyGenOpts")
	// assert.False(t, importOpts.Ephemeral(), "Expected keygenOpts.Ephemeral() ==> false")
	// assert.True(t, importOpts.Algorithm() == ecdsaImportOpts, "Unexpected ECDSA import opts, expected [%v], got [%v]", ecdsaImportOpts, importOpts.Algorithm())
}
