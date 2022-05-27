/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

/*
pkg/gateway/spi.go 钱包存储接口
*/

// This contains the service provider interface (SPI) which provides the mechanism
// for implementing alternative gateway strategies, wallets, etc.
// This is currently experimental and will be implemented in future user stories

// WalletStore is the interface for implementations that provide backing storage for identities in a wallet.
// To create create a new backing store, implement all the methods defined in this interface and provide
// a factory method that wraps an instance of this in a new Wallet object. E.g:
//   func NewMyWallet() *Wallet {
//	   store := &myWalletStore{ }
//	   return &Wallet{store}
//   }
// WalletStore接口与wallet接口的区别可能是，WalletStore接口是公开接口，开发者可以自己实现该接口，实现自己的钱包。
type WalletStore interface {
	// 写入目标
	Put(label string, stream []byte) error
	// 获取目标
	Get(label string) ([]byte, error)
	// 列举所有存储目标
	List() ([]string, error)
	// 目标是否存在
	Exists(label string) bool
	// 删除目标
	Remove(label string) error
}
