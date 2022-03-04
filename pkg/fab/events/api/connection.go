/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/context"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/fab"
)

// Connection defines the functions for an event server connection
type Connection interface {
	// Receive sends events to the given channel
	Receive(chan<- interface{})
	// Close closes the connection
	Close()
	// Closed return true if the connection is closed
	Closed() bool
}

// ConnectionProvider creates a Connection.
type ConnectionProvider func(context context.Client, chConfig fab.ChannelCfg, peer fab.Peer) (Connection, error)
