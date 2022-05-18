/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sw

import (
	"reflect"

	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/internal/gitee.com/zhaochuninhefei/fabric-gm/bccsp"
	"gitee.com/zhaochuninhefei/gmgo/sm3"
	"github.com/pkg/errors"
)

/*
bccsp/sw/new.go 提供创建`sw.CSP`(bccsp/sw/impl.go)的函数:
NewDefaultSecurityLevel
NewDefaultSecurityLevelWithKeystore
NewWithParams
默认hashFamily改为SM3
*/

// NewDefaultSecurityLevel returns a new instance of the software-based BCCSP
// at security level 256, hash family SM3 and using FolderBasedKeyStore as KeyStore.
func NewDefaultSecurityLevel(keyStorePath string) (bccsp.BCCSP, error) {
	ks := &fileBasedKeyStore{}
	if err := ks.Init(nil, keyStorePath, false); err != nil {
		return nil, errors.Wrapf(err, "Failed initializing key store at [%v]", keyStorePath)
	}
	// 改为使用国密
	return NewWithParams(true, 256, "SM3", ks)
}

// NewDefaultSecurityLevel returns a new instance of the software-based BCCSP
// at security level 256, hash family SM3 and using the passed KeyStore.
func NewDefaultSecurityLevelWithKeystore(keyStore bccsp.KeyStore) (bccsp.BCCSP, error) {
	// 改为使用国密
	return NewWithParams(true, 256, "SM3", keyStore)
}

// NewWithParams returns a new instance of the software-based BCCSP
// set at the passed security level, hash family and KeyStore.
// 根据参数生成swbccsp
func NewWithParams(usingGM bool, securityLevel int, hashFamily string, keyStore bccsp.KeyStore) (bccsp.BCCSP, error) {
	// Init config
	conf := &config{}
	err := conf.setSecurityLevel(usingGM, securityLevel, hashFamily)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed initializing configuration at [%v,%v]", securityLevel, hashFamily)
	}

	swbccsp, err := New(keyStore)
	if err != nil {
		return nil, err
	}

	// Notice that errors are ignored here because some test will fail if one
	// of the following call fails.

	// 添加国密相关组件 无视 usingGM 固定使用国密组件
	// if usingGM {
	// Set the key generators
	// sm2密钥对构造器
	swbccsp.AddWrapper(reflect.TypeOf(&bccsp.SM2KeyGenOpts{}), &sm2KeyGenerator{})
	// sm4密钥构造器
	swbccsp.AddWrapper(reflect.TypeOf(&bccsp.SM4KeyGenOpts{}), &sm4KeyGenerator{length: conf.gmByteLength})

	// Set the key deriver
	// 国密没有密钥派生相关实现

	// Set the key importers
	// sm2私钥导入器
	swbccsp.AddWrapper(reflect.TypeOf(&bccsp.SM2PrivateKeyImportOpts{}), &sm2PrivateKeyOptsKeyImporter{})
	// sm2公钥导入器
	swbccsp.AddWrapper(reflect.TypeOf(&bccsp.SM2PublicKeyImportOpts{}), &sm2PublicKeyOptsKeyImporter{})
	swbccsp.AddWrapper(reflect.TypeOf(&bccsp.SM2GoPublicKeyImportOpts{}), &sm2GoPublicKeyOptsKeyImporter{})
	// sm4密钥导入器
	swbccsp.AddWrapper(reflect.TypeOf(&bccsp.SM4ImportKeyOpts{}), &sm4ImportKeyOptsKeyImporter{})
	// gmx509公钥导入，x509的签名内容的核心是证书拥有者的公钥，与签名算法无关，因此可能是sm2,ecdsa或rsa
	swbccsp.AddWrapper(reflect.TypeOf(&bccsp.GMX509PublicKeyImportOpts{}), &gmx509PublicKeyImportOptsKeyImporter{bccsp: swbccsp})
	// // 国密HMac认证码导入器
	// swbccsp.AddWrapper(reflect.TypeOf(&bccsp.GMHMACImportKeyOpts{}), &gmHmacImportKeyOptsKeyImporter{})

	// Set the Hashers
	// sm3散列
	swbccsp.AddWrapper(reflect.TypeOf(&bccsp.SM3Opts{}), &hasher{hash: sm3.New})

	// Set the Signers
	// sm2私钥签名
	swbccsp.AddWrapper(reflect.TypeOf(&SM2PrivateKey{}), &sm2Signer{})

	// Set the Verifiers
	// sm2公钥验签
	swbccsp.AddWrapper(reflect.TypeOf(&SM2PublicKey{}), &sm2PublicKeyKeyVerifier{})
	// sm2私钥验签，实际还是公钥验签
	swbccsp.AddWrapper(reflect.TypeOf(&SM2PrivateKey{}), &sm2PrivateKeyVerifier{})

	// Set the Encryptors
	// sm4加密，未分组 --> 该问题已对应，修改了`bccsp/sw/sm4.go`的sm4Encryptor与sm4Decryptor的接口实现方法
	swbccsp.AddWrapper(reflect.TypeOf(&SM4Key{}), &sm4Encryptor{})

	// Set the Decryptors
	// sm4解密，未分组 --> 该问题已对应，修改了`bccsp/sw/sm4.go`的sm4Encryptor与sm4Decryptor的接口实现方法
	swbccsp.AddWrapper(reflect.TypeOf(&SM4Key{}), &sm4Decryptor{})

	swbccsp.Algorithms = "sm2-sm3-sm4"
	return swbccsp, nil
}
