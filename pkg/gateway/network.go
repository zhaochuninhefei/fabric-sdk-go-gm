/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/client/channel"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/client/event"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/client/ledger"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/context"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/fab"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/fab/events/deliverclient/seek"
	"github.com/pkg/errors"
)

// Network对象表示fabric网络通道中的peer节点集合。
//
// A Network object represents the set of peers in a Fabric network (channel).
// Applications should get a Network instance from a Gateway using the GetNetwork method.
type Network struct {
	name    string
	gateway *Gateway
	client  *channel.Client
	event   *event.Client
}

// 根据 gateway 与 channelProvider 创建网络通道实例
func newNetwork(gateway *Gateway, channelProvider context.ChannelProvider) (*Network, error) {
	n := Network{
		gateway: gateway,
	}
	// 使用channelProvider创建一个通道客户端实例，用于执行交易
	// Channel client is used to query and execute transactions
	client, err := channel.New(channelProvider)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create new channel client")
	}
	n.client = client
	// 获取通道上下文
	ctx, err := channelProvider()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create new channel context")
	}
	// 获取ChannelID
	n.name = ctx.ChannelID()

	eventOpts := []event.ClientOption{event.WithBlockEvents()}
	if gateway.options.FromBlockSet {
		eventOpts = append(eventOpts, event.WithSeekType(seek.FromBlock), event.WithBlockNum(gateway.options.FromBlock))
	}

	n.event, err = event.New(channelProvider, eventOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create new event client")
	}

	// the following is really to kick the discovery service into getting the TLScert
	// so that subsequent SubmitTransaction can connect to the orderer
	members, err := ctx.ChannelService().Membership()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to query channel membership")
	}
	members.ContainsMSP(gateway.mspid)

	return &n, nil
}

// 通道名，即ChannelID
//
// Name is the name of the network (also known as channel name)
func (n *Network) Name() string {
	return n.name
}

// 获取智能合约实例
//
// GetContract returns instance of a smart contract on the current network.
//  Parameters:
//  chaincodeID is the name of the chaincode that contains the smart contract
//
//  Returns:
//  A Contract object representing the smart contract
func (n *Network) GetContract(chaincodeID string) *Contract {
	return newContract(n, chaincodeID, "")
}

// GetContractWithName returns instance of a smart contract on the current network.
// If the chaincode instance contains more
// than one smart contract class (available using the latest contract programming model), then an
// individual class can be selected.
//  Parameters:
//  chaincodeID is the name of the chaincode that contains the smart contract
//  name is the class name of the smart contract within the chaincode.
//
//  Returns:
//  A Contract object representing the smart contract
func (n *Network) GetContractWithName(chaincodeID string, name string) *Contract {
	return newContract(n, chaincodeID, name)
}

// RegisterBlockEvent registers for block events. Unregister must be called when the registration is no longer needed.
//  Returns:
//  the registration and a channel that is used to receive events. The channel is closed when Unregister is called.
func (n *Network) RegisterBlockEvent() (fab.Registration, <-chan *fab.BlockEvent, error) {
	return n.event.RegisterBlockEvent()
}

// RegisterFilteredBlockEvent registers for filtered block events. Unregister must be called when the registration is no longer needed.
//  Returns:
//  the registration and a channel that is used to receive events. The channel is closed when Unregister is called.
func (n *Network) RegisterFilteredBlockEvent() (fab.Registration, <-chan *fab.FilteredBlockEvent, error) {
	return n.event.RegisterFilteredBlockEvent()
}

// Unregister removes the given registration and closes the event channel.
//  Parameters:
//  registration is the registration handle that was returned from RegisterBlockEvent method
func (n *Network) Unregister(registration fab.Registration) {
	n.event.Unregister(registration)
}

// 获取LedgerClient 账本客户端实例
//  目前LedgerClient用于统计该通道的区块高度与交易总数
func (n *Network) GetLedgerClient() (*ledger.Client, error) {
	return ledger.New(n.client.ChannelProvider)
}
