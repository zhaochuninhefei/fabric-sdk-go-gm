/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package gateway enables Go developers to build client applications using the Hyperledger
// Fabric programming model as described in the 'Developing Applications' chapter of the Fabric
// documentation: https://hyperledger-fabric.readthedocs.io/en/latest/developapps/developing_applications.html
//
// A Gateway object is created using the Connect() function to connect to a 'gateway' peer
// as specified in a network configuration file, using an identity stored in a wallet.
// Interactions with smart contracts are then invoked within the context of this gateway connection.
//
// See https://github.com/hyperledger/fabric-samples/blob/master/fabcar/go/fabcar.go
// for a working sample.
package gateway

/*
pkg/gateway/gateway.go 区块链网络的入口
*/

import (
	"os"
	"strings"
	"time"

	fabricCaUtil "gitee.com/zhaochuninhefei/fabric-sdk-go-gm/internal/gitee.com/zhaochuninhefei/fabric-ca-gm/sdkinternal/pkg/util"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/client/resmgmt"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/errors/retry"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/context"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/core"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/fab"
	mspProvider "gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/common/providers/msp"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/core/config/lookup"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/core/cryptosuite"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/fabsdk"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/fabsdk/api"
	"gitee.com/zhaochuninhefei/zcgolog/zclog"
	"github.com/pkg/errors"
)

const (
	defaultTimeout      = 5 * time.Minute
	localhostEnvVarName = "DISCOVERY_AS_LOCALHOST"
)

// Gateway is the entry point to a Fabric network
// 区块链网络入口结构体
type Gateway struct {
	sdk        *fabsdk.FabricSDK
	options    *gatewayOptions
	cfg        core.ConfigBackend
	org        string
	mspid      string
	peers      []fab.PeerConfig
	mspfactory api.MSPProviderFactory
}

// 网络入口参数
type gatewayOptions struct {
	Identity mspProvider.SigningIdentity
	User     string
	Timeout  time.Duration
	// FromBlock specify the initial block to be considerer by event client
	FromBlock    uint64
	FromBlockSet bool
}

// Option functional arguments can be supplied when connecting to the gateway.
type Option = func(*Gateway) error

// ConfigOption specifies the gateway configuration source.
type ConfigOption = func(*Gateway) error

// IdentityOption specifies the user identity under which all transactions are performed for this gateway instance.
//
// 声明一个接口类型:IdentityOption,该接口类型必须实现函数:func(*Gateway) error
// 实现的函数是一个apply函数，以自身变量名为函数名调用。
type IdentityOption = func(*Gateway) error

// Connect to a gateway defined by a network config file.
// Must specify a config option, an identity option and zero or more strategy options.
//  Parameters:
//  config is a ConfigOption used to specify the network connection configuration.  This must contain connection details for at least one 'gateway' peer.
//  identity is an IdentityOption which assigns a signing identity for all interactions under this Gateway connection.
//  options specifies other gateway options
//
//  Returns:
//  A Transaction object for subsequent evaluation or submission.
// 根据连接配置文件与身份信息连接区块链网络
func Connect(config ConfigOption, identity IdentityOption, options ...Option) (*Gateway, error) {

	g := &Gateway{
		options: &gatewayOptions{
			Timeout: defaultTimeout,
		},
	}

	err := identity(g)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to apply identity option")
	}

	err = config(g)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to apply config option")
	}

	for _, option := range options {
		err = option(g)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to apply gateway option")
		}
	}

	return g, nil
}

// WithConfig configures the gateway from a network config, such as a ccp file.
//
//   Parameters:
//   config is a ConfigProvider function which provides config backend
//
//   Returns:
//   A ConfigOption which can be passed as the first parameter to the Connect() function
func WithConfig(config core.ConfigProvider) ConfigOption {
	return func(gw *Gateway) error {
		config = createGatewayConfigProvider(config, gw.getOrg)

		configBackend, err := config()
		if err != nil {
			return err
		}
		if len(configBackend) != 1 {
			return errors.New("invalid config file")
		}

		gw.cfg = configBackend[0]

		value, ok := gw.cfg.Lookup("client.organization")
		if !ok {
			return errors.New("No client organization defined in the config")
		}
		gw.org = value.(string)

		value, ok = gw.cfg.Lookup("organizations." + gw.org + ".mspid")
		if !ok {
			return errors.New("No client organization defined in the config")
		}
		gw.mspid = value.(string)

		opts := []fabsdk.Option{}
		if gw.mspfactory != nil {
			opts = append(opts, fabsdk.WithMSPPkg(gw.mspfactory))
		}

		sdk, err := fabsdk.New(config, opts...)

		if err != nil {
			return err
		}

		gw.sdk = sdk

		//  find the 'gateway' peers
		ctx := sdk.Context()
		client, _ := ctx()
		gw.peers, _ = client.EndpointConfig().PeersConfig(gw.org)

		return nil
	}
}

// WithSDK configures the gateway with the configuration from an existing FabricSDK instance
//
//   Parameters:
//   sdk is an instance of fabsdk.FabricSDK from which the configuration is extracted
//
//   Returns:
//   A ConfigOption which can be passed as the first parameter to the Connect() function
func WithSDK(sdk *fabsdk.FabricSDK) ConfigOption {
	return func(gw *Gateway) error {
		gw.sdk = sdk

		cfg, err := sdk.Config()

		if err != nil {
			return errors.Wrap(err, "Unable to access SDK configuration")
		}

		value, ok := cfg.Lookup("client.organization")
		if !ok {
			return errors.New("No client organization defined in the config")
		}
		gw.org = value.(string)

		return nil
	}
}

// WithIdentity is an optional argument to the Connect method which specifies
// the identity that is to be used to connect to the network.
// All operations under this gateway connection will be performed using this identity.
//
//   Parameters:
//   wallet is a Wallet implementation that contains identities
//   label is the name of the identity in the wallet to associate with the gateway
//
//   Returns:
//   An IdentityOption which can be passed as the second parameter to the Connect() function
func WithIdentity(wallet wallet, label string) IdentityOption {
	return func(gw *Gateway) error {
		// 从钱包中获取身份信息
		creds, err := wallet.Get(label)
		if err != nil {
			return err
		}
		// 根据当前身份的私钥字节数组与默认的CSP密码套件，生成私钥
		privateKey, _ := fabricCaUtil.ImportBCCSPKeyFromPEMBytes([]byte(creds.(*X509Identity).Key()), cryptosuite.GetDefault(), true)
		// 获取用于连接fabric区块链网络的身份用户:wid
		wid := &walletIdentity{
			id:                    label,
			mspID:                 creds.mspID(),
			enrollmentCertificate: []byte(creds.(*X509Identity).Certificate()),
			privateKey:            privateKey,
		}
		zclog.Debugf("===== 尝试从wallet获取账户信息 id: %s, mspID: %s, privateKey: %s", wid.id, wid.mspID, wid.privateKey)
		// 将连接fabric区块链网络的身份用户:wid设置给gateway对象
		gw.options.Identity = wid
		gw.mspfactory = &walletmsp{}

		return nil
	}
}

// WithUser is an optional argument to the Connect method which specifies
// the identity that is to be used to connect to the network.
// The creadentials are extracted from the credential store specified in the connection profile.
// All operations under this gateway connection will be performed using this identity.
//
//   Parameters:
//   user is the name of the user in the credential store.
//
//   Returns:
//   An IdentityOption which can be passed as the second parameter to the Connect() function
func WithUser(user string) IdentityOption {
	return func(gw *Gateway) error {
		gw.options.User = user
		return nil
	}
}

// WithTimeout is an optional argument to the Connect method which
// defines the commit timeout for all transaction submissions for this gateway.
func WithTimeout(timeout time.Duration) Option {
	return func(gw *Gateway) error {
		gw.options.Timeout = timeout
		return nil
	}
}

// WithBlockNum optionaly indicates the block number from which events are to be received.
func WithBlockNum(from uint64) Option {
	return func(gw *Gateway) error {
		gw.options.FromBlock = from
		gw.options.FromBlockSet = true
		return nil
	}
}

// GetNetwork returns an object representing a network channel.
//  Parameters:
//  name is the name of the network channel
//
//  Returns:
//  A Network object representing the channel
// 从Gateway获取网络通道
func (gw *Gateway) GetNetwork(name string) (*Network, error) {
	// 获取 channelProvider 通道客户端上下文
	var channelProvider context.ChannelProvider
	if gw.options.Identity != nil {
		channelProvider = gw.sdk.ChannelContext(name, fabsdk.WithIdentity(gw.options.Identity), fabsdk.WithOrg(gw.org))
	} else {
		channelProvider = gw.sdk.ChannelContext(name, fabsdk.WithUser(gw.options.User), fabsdk.WithOrg(gw.org))
	}
	return newNetwork(gw, channelProvider)
}

// Close the gateway connection and all associated resources, including removing listeners attached to networks and
// contracts created by the gateway.
func (gw *Gateway) Close() {
	// future use
}

func (gw *Gateway) getOrg() string {
	return gw.org
}

// QueryChannels 查询当前连接peer已加入的通道
func (gw *Gateway) QueryChannels() ([]string, error) {
	// 获取当前连接配置
	configBackend, err := gw.sdk.Config()
	if err != nil {
		return nil, errors.Errorf("Failed to get config backend from SDK: %s", err)
	}
	targets, err := orgTargetPeers([]string{gw.org}, configBackend)
	if err != nil {
		return nil, errors.Errorf("Creating peers failed: %s", err)
	}
	var clientContext context.ClientProvider
	if gw.options.Identity != nil {
		clientContext = gw.sdk.Context(fabsdk.WithIdentity(gw.options.Identity), fabsdk.WithOrg(gw.org))
	} else {
		clientContext = gw.sdk.Context(fabsdk.WithUser(gw.options.User), fabsdk.WithOrg(gw.org))
	}
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return nil, errors.Errorf("Failed to query channel management client: %s", err)
	}
	channelQueryResponse, err := resMgmtClient.QueryChannels(
		resmgmt.WithTargetEndpoints(targets[0]), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return nil, errors.Errorf("QueryChannels return error: %s", err)
	}
	var channelIds []string
	for _, channel := range channelQueryResponse.Channels {
		channelIds = append(channelIds, channel.ChannelId)
	}
	return channelIds, nil
}

// QueryOrgTargetPeers 查看目标配置下目标组织的peers
func (gw *Gateway) QueryOrgTargetPeers() ([]string, error) {
	configBackend, err := gw.sdk.Config()
	if err != nil {
		return nil, errors.Errorf("Failed to get config backend from SDK: %s", err)
	}
	return orgTargetPeers([]string{gw.org}, configBackend)
}

// orgTargetPeers 查看目标配置下目标组织的peers
func orgTargetPeers(orgs []string, configBackend ...core.ConfigBackend) ([]string, error) {
	networkConfig := fab.NetworkConfig{}
	err := lookup.New(configBackend...).UnmarshalKey("organizations", &networkConfig.Organizations)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get organizations from config ")
	}

	var peers []string
	for _, org := range orgs {
		orgConfig, ok := networkConfig.Organizations[strings.ToLower(org)]
		if !ok {
			continue
		}
		peers = append(peers, orgConfig.Peers...)
	}
	return peers, nil
}

func createGatewayConfigProvider(config core.ConfigProvider, org func() string) func() ([]core.ConfigBackend, error) {
	return func() ([]core.ConfigBackend, error) {
		configBackend, err := config()
		if err != nil {
			return nil, err
		}
		if len(configBackend) != 1 {
			return nil, errors.New("invalid config file")
		}

		cfg := configBackend[0]

		lhConfig := make([]core.ConfigBackend, 0)
		lhConfig = append(lhConfig, createGatewayConfig(cfg, org()))

		return lhConfig, nil
	}
}

func createGatewayConfig(backend core.ConfigBackend, org string) *gatewayConfig {
	var matchers map[string][]map[string]string
	if strings.ToUpper(os.Getenv(localhostEnvVarName)) == "TRUE" {
		matchers = createLocalhostMappings()
	}

	var channelConfig map[string]map[string]map[string]map[string]bool
	_, exists := backend.Lookup("channels")
	if !exists {
		channelConfig = createDefaultChannelConfig(backend, org)
	}

	return &gatewayConfig{
		backend:    backend,
		matchers:   matchers,
		channelDef: channelConfig,
	}
}

/* dynamically add the following to CCP:

entityMatchers:
  peer:
    - pattern: ([^:]+):(\\d+)
      urlSubstitutionExp: localhost:${2}
      sslTargetOverrideUrlSubstitutionExp: ${1}
      mappedHost: ${1}
  orderer:
    - pattern: ([^:]+):(\\d+)
      urlSubstitutionExp: localhost:${2}
      sslTargetOverrideUrlSubstitutionExp: ${1}
      mappedHost: ${1}
*/
func createLocalhostMappings() map[string][]map[string]string {
	matchers := make(map[string][]map[string]string)
	mappings := make([]map[string]string, 0)

	mapping := make(map[string]string)
	mapping["pattern"] = "([^:]+):(\\d+)"
	mapping["urlSubstitutionExp"] = "localhost:${2}"
	mapping["sslTargetOverrideUrlSubstitutionExp"] = "${1}"
	mapping["mappedHost"] = "${1}"
	mappings = append(mappings, mapping)

	matchers["peer"] = mappings
	matchers["orderer"] = mappings

	return matchers
}

/* dynamically add the following to CCP:

channels:
  _default:
    peers:
      <gateway_peer_name>:
        endorsingPeer: true
        chaincodeQuery: true
        ledgerQuery: true
        eventSource: true
*/
func createDefaultChannelConfig(backend core.ConfigBackend, org string) map[string]map[string]map[string]map[string]bool {
	channels := make(map[string]map[string]map[string]map[string]bool)
	_default := make(map[string]map[string]map[string]bool)
	gateways := make(map[string]map[string]bool)
	roles := make(map[string]bool)
	roles["endorsingPeer"] = true
	roles["chaincodeQuery"] = true
	roles["ledgerQuery"] = true
	roles["eventSource"] = true

	value, ok := backend.Lookup("organizations." + org + ".peers")
	if !ok {
		return nil
	}
	arr := value.([]interface{})
	for _, gatewayPeer := range arr {
		gateways[gatewayPeer.(string)] = roles
	}

	_default["peers"] = gateways
	channels["_default"] = _default
	return channels
}

type gatewayConfig struct {
	backend    core.ConfigBackend
	matchers   map[string][]map[string]string
	channelDef map[string]map[string]map[string]map[string]bool
}

func (gc *gatewayConfig) Lookup(key string) (interface{}, bool) {
	if key == "entityMatchers" && gc.matchers != nil {
		return gc.matchers, true
	}
	conf, exists := gc.backend.Lookup(key)
	if key == "channels" && gc.channelDef != nil {
		return gc.channelDef, true
	}
	return conf, exists
}
