/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package balancedsorter

import (
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/client/common/selection/options"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/logging"
	coptions "gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/options"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/fab"
)

var logger = logging.NewLogger("fabsdk/client")

// New returns a peer sorter that chooses a peer according to a provided balancer.
func New(opts ...coptions.Opt) options.PeerSorter {
	params := defaultParams()
	coptions.Apply(params, opts)

	return func(peers []fab.Peer) []fab.Peer {
		return params.balancer(peers)
	}
}
