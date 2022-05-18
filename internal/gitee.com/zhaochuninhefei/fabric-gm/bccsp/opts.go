/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bccsp

import "io"

/*
 * bccsp/opts.go 实现部分`bccsp.KeyGenOpts`、`bccsp.KeyImportOpts`与`bccsp.KeyDerivOpts`接口。
 * ecdsa相关: ECDSAKeyGenOpts, ECDSAPrivateKeyImportOpts, ECDSAPKIXPublicKeyImportOpts, ECDSAGoPublicKeyImportOpts, ECDSAReRandKeyOpts
 * sm2相关: SM2KeyGenOpts, SM2PrivateKeyImportOpts, SM2PublicKeyImportOpts
 * sm4相关: SM4KeyGenOpts, SM4ImportKeyOpts
 * aes相关: AESKeyGenOpts, AES256ImportKeyOpts
 * hmac相关: HMACTruncated256AESDeriveKeyOpts, HMACDeriveKeyOpts, HMACImportKeyOpts
 * sha相关: SHAOpts
 * x509相关: X509PublicKeyImportOpts
 */

const (
	// GMX509Certificate
	GMX509Certificate = "GMX509Certificate"
	// SM4
	SM4 = "SM4"
	// SM3
	SM3 = "SM3"
	// SM2
	SM2 = "SM2"
)

// GMX509PublicKeyImportOpts contains options for importing public keys from an gmx509 certificate
type GMX509PublicKeyImportOpts struct {
	Temporary bool
}

// Algorithm returns the key importation algorithm identifier (to be used).
func (opts *GMX509PublicKeyImportOpts) Algorithm() string {
	return GMX509Certificate
}

// Ephemeral returns true if the key to generate has to be ephemeral,
// false otherwise.
func (opts *GMX509PublicKeyImportOpts) Ephemeral() bool {
	return opts.Temporary
}

// SM2KeyGenOpts contains options for SM2 key generation.
type SM2KeyGenOpts struct {
	Temporary bool
}

// Algorithm returns the key generation algorithm identifier (to be used).
func (opts *SM2KeyGenOpts) Algorithm() string {
	return SM2
}

// Ephemeral returns true if the key to generate has to be ephemeral,
// false otherwise.
func (opts *SM2KeyGenOpts) Ephemeral() bool {
	return opts.Temporary
}

// SM4KeyGenOpts contains options for SM2 key generation.
type SM4KeyGenOpts struct {
	Temporary bool
}

// Algorithm returns the key generation algorithm identifier (to be used).
func (opts *SM4KeyGenOpts) Algorithm() string {
	return SM4
}

// Ephemeral returns true if the key to generate has to be ephemeral,
// false otherwise.
func (opts *SM4KeyGenOpts) Ephemeral() bool {
	return opts.Temporary
}

//SM4ImportKeyOpts  实现  bccsp.KeyImportOpts 接口
type SM4ImportKeyOpts struct {
	Temporary bool
}

// Algorithm returns the key importation algorithm identifier (to be used).
func (opts *SM4ImportKeyOpts) Algorithm() string {
	return SM4
}

// Ephemeral returns true if the key generated has to be ephemeral,
// false otherwise.
func (opts *SM4ImportKeyOpts) Ephemeral() bool {
	return opts.Temporary
}

type SM4EncrypterDecrypterOpts struct {
	// 初始偏移量 在 CBC, CFB, OFB 分组模式下需要
	IV []byte
	// PRNG is an instance of a PRNG to be used by the underlying cipher.
	// It is used only if different from nil.
	PRNG io.Reader
}

//SM2PrivateKeyImportOpts  实现  bccsp.KeyImportOpts 接口
type SM2PrivateKeyImportOpts struct {
	Temporary bool
}

// Algorithm returns the key importation algorithm identifier (to be used).
func (opts *SM2PrivateKeyImportOpts) Algorithm() string {
	return SM2
}

// Ephemeral returns true if the key generated has to be ephemeral,
// false otherwise.
func (opts *SM2PrivateKeyImportOpts) Ephemeral() bool {
	return opts.Temporary
}

//SM2PublicKeyImportOpts  实现  bccsp.KeyImportOpts 接口
type SM2PublicKeyImportOpts struct {
	Temporary bool
}

// Algorithm returns the key importation algorithm identifier (to be used).
func (opts *SM2PublicKeyImportOpts) Algorithm() string {
	return SM2
}

// Ephemeral returns true if the key generated has to be ephemeral,
// false otherwise.
func (opts *SM2PublicKeyImportOpts) Ephemeral() bool {
	return opts.Temporary
}

//SM2GoPublicKeyImportOpts  实现  bccsp.KeyImportOpts 接口
type SM2GoPublicKeyImportOpts struct {
	Temporary bool
}

// Algorithm returns the key importation algorithm identifier (to be used).
func (opts *SM2GoPublicKeyImportOpts) Algorithm() string {
	return SM2
}

// Ephemeral returns true if the key generated has to be ephemeral,
// false otherwise.
func (opts *SM2GoPublicKeyImportOpts) Ephemeral() bool {
	return opts.Temporary
}
