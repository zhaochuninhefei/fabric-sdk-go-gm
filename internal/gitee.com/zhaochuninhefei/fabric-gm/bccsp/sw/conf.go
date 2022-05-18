/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sw

import (
	"crypto/elliptic"
	"fmt"
	"hash"

	"gitee.com/zhaochuninhefei/gmgo/sm2"
	"gitee.com/zhaochuninhefei/gmgo/sm3"
)

/*
bccsp/sw/conf.go 提供bccsp配置
*/

type config struct {
	// 是否使用国密
	usingGM bool
	// 国密椭圆曲线
	gmCurve elliptic.Curve
	// 国密散列函数
	gmFunction func() hash.Hash
	// 国密密钥位数
	gmByteLength int
}

// 设置安全级别配置
func (conf *config) setSecurityLevel(usingGM bool, securityLevel int, hashFamily string) (err error) {
	// 国密对应，无视 usingGM ，固定使用国密
	conf.usingGM = true
	if securityLevel == 256 && hashFamily == "SM3" {
		_ = conf.setSecurityLevelWithSM2SM3()
	} else {
		err = fmt.Errorf("bccsp国密改造版目前只支持国密SM3 256位，不支持 [%s] [%d]位", hashFamily, securityLevel)
	}
	return
}

// 设置使用国密时的安全级别配置
func (conf *config) setSecurityLevelWithSM2SM3() (err error) {
	conf.gmCurve = sm2.P256Sm2()
	conf.gmFunction = sm3.New
	conf.gmByteLength = 16
	return nil
}
