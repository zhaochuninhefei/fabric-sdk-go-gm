/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

/*
pkg/gateway/x509identity.go x509身份相关定义与操作，用于钱包中存储的身份对象。
*/

import "encoding/json"

const x509Type = "X.509"

// X509Identity represents an X509 identity
//  x509身份结构体
type X509Identity struct {
	// 版本
	Version int `json:"version"`
	// 身份所属组织的MSPID
	MspID string `json:"mspId"`
	// 身份类型，目前固定为"X.509"
	IDType string `json:"type"`
	// 身份凭证(数字证书+私钥)，一般是给应用申请的client类型账户
	Credentials credentials `json:"credentials"`
}

// 身份凭证结构体
type credentials struct {
	// 数字证书PEM字符串
	Certificate string `json:"certificate"`
	// 私钥PEM字符串
	Key string `json:"privateKey"`
}

// Type returns X509 for this identity type
// 返回当前x509身份的idType，固定为"X.509"
func (x *X509Identity) idType() string {
	return x509Type
}

// 返回当前x509身份的MspID
func (x *X509Identity) mspID() string {
	return x.MspID
}

// Certificate returns the X509 certificate PEM
// 返回当前x509身份的数字证书PEM字符串
func (x *X509Identity) Certificate() string {
	return x.Credentials.Certificate
}

// Key returns the private key PEM
// 返回当前x509身份的私钥PEM字符串
func (x *X509Identity) Key() string {
	return x.Credentials.Key
}

// NewX509Identity creates an X509 identity for storage in a wallet
// 创建一个新的x509身份。
//  mspid:目标身份所属组织的mspid
//  cert:目标身份的数字证书PEM字符串
//  key:目标身份的私钥PEM字符串
func NewX509Identity(mspid string, cert string, key string) *X509Identity {
	return &X509Identity{1, mspid, x509Type, credentials{cert, key}}
}

// 将当前x509身份序列化为json字节数组
func (x *X509Identity) toJSON() ([]byte, error) {
	return json.Marshal(x)
}

// 将json字节数组反序列化为x509身份
func (x *X509Identity) fromJSON(data []byte) (Identity, error) {
	err := json.Unmarshal(data, x)

	if err != nil {
		return nil, err
	}

	return x, nil
}
