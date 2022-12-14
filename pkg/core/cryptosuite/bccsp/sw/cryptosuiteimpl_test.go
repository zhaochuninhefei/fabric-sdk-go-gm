/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sw

import (
	"bytes"
	"testing"

	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/internal/gitee.com/zhaochuninhefei/fabric-gm/bccsp"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/core"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/test/mockcore"
	"gitee.com/zhaochuninhefei/gmgo/sm3"
	"github.com/golang/mock/gomock"
)

func TestBadConfig(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := mockcore.NewMockCryptoSuiteConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("UNKNOWN")
	mockConfig.EXPECT().SecurityProvider().Return("UNKNOWN")

	//Get cryptosuite using config
	_, err := GetSuiteByConfig(mockConfig)
	if err == nil {
		t.Fatal("Unknown security provider should return error")
	}
}

func TestCryptoSuiteByConfigSW(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := mockcore.NewMockCryptoSuiteConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("sw").AnyTimes()
	mockConfig.EXPECT().SecurityAlgorithm().Return("SM3")
	mockConfig.EXPECT().SecurityLevel().Return(256)
	mockConfig.EXPECT().KeyStorePath().Return("/tmp/msp")

	//Get cryptosuite using config
	c, err := GetSuiteByConfig(mockConfig)
	if err != nil {
		t.Fatalf("Not supposed to get error, but got: %s", err)
	}

	verifyHashFn(t, c)
}

func TestCryptoSuiteByBadConfigSW(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := mockcore.NewMockCryptoSuiteConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("sw")
	mockConfig.EXPECT().SecurityAlgorithm().Return("SM3")
	mockConfig.EXPECT().SecurityLevel().Return(256)
	mockConfig.EXPECT().KeyStorePath().Return("")

	//Get cryptosuite using config
	_, err := GetSuiteByConfig(mockConfig)
	if err == nil {
		t.Fatal("Bad configuration should return error")
	}
}

func TestCryptoSuiteDefaultEphemeral(t *testing.T) {
	c, err := GetSuiteWithDefaultEphemeral()
	if err != nil {
		t.Fatalf("Not supposed to get error, but got: %s", err)
	}
	verifyHashFn(t, c)
}

func verifyHashFn(t *testing.T, c core.CryptoSuite) {
	msg := []byte("Hello")
	e := sm3.Sm3Sum(msg)
	a, err := c.Hash(msg, &bccsp.SM3Opts{})
	if err != nil {
		t.Fatalf("Not supposed to get error, but got: %s", err)
	}

	if !bytes.Equal(a, e[:]) {
		t.Fatal("Expected SHA 256 hash function")
	}
}
