/*
Copyright (c) 2022 zhaochun
gitee.com/zhaochuninhefei/fabric-sdk-go-gm is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
		 http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package chn_browse_util

/*
pkg/util/chn_browse_util/chn_browse_util.go 通道浏览工具库，提供用于查询通道的区块与交易信息的通用函数。
*/

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"gitee.com/zhaochuninhefei/fabric-protos-go-gm/common"
	"gitee.com/zhaochuninhefei/fabric-protos-go-gm/ledger/rwset"
	"gitee.com/zhaochuninhefei/fabric-protos-go-gm/ledger/rwset/kvrwset"
	"gitee.com/zhaochuninhefei/fabric-protos-go-gm/msp"
	"gitee.com/zhaochuninhefei/fabric-protos-go-gm/peer"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/client/ledger"
	"gitee.com/zhaochuninhefei/gmgo/x509"
	"github.com/golang/protobuf/proto"
)

// 通道情报
type ChannelInfo struct {
	BlockHeight      uint64             // 区块高度
	TransTotal       uint64             // 交易总数(仅统计本次浏览的区块)
	BlockInfoWithTxs []*BlockInfoWithTx // 区块情报(包含内部交易情报)集合
	TransactionInfos []*TransactionInfo // 交易情报集合
	BlockBasicInfos  []*BlockInfoBasic  // 区块基础信息集合
}

func (t *ChannelInfo) ToString() string {
	result := fmt.Sprintf("区块高度:%d, 交易总数: %d,\n区块集合:\n",
		t.BlockHeight, t.TransTotal)
	for _, b := range t.BlockInfoWithTxs {
		result = result + "\t" + b.ToString() + "\n"
	}
	return result
}

// 区块情报(包含内部交易情报)
type BlockInfoWithTx struct {
	BlockInfoBasic                      // 区块基础信息
	TransactionInfos []*TransactionInfo // 区块内交易情报集合
}

func (t *BlockInfoWithTx) ToString() string {
	result := fmt.Sprintf("%s,\n\t\t交易集合:\n", t.BlockInfoBasic.ToString())
	for _, t := range t.TransactionInfos {
		result = result + "\t\t" + t.ToString() + "\n"
	}
	return result
}

func (t *BlockInfoWithTx) GetBasicInfo() *BlockInfoBasic {
	return &t.BlockInfoBasic
}

// 区块基础信息(区块编号、区块哈希、前区块哈希、区块内交易数量)
type BlockInfoBasic struct {
	BlockNum           uint64 // 区块编号
	BlockHeaderHash    string // 区块头哈希(16进制字符串)
	PreBlockHeaderHash string // 前区块头哈希(16进制字符串)
	TransCnt           uint64 // 区块内交易数量
}

func (t *BlockInfoBasic) ToString() string {
	result := fmt.Sprintf("区块编号: %d, 交易数量: %d, 区块头哈希: %s, 前区块头哈希: %s",
		t.BlockNum, t.TransCnt, t.BlockHeaderHash, t.PreBlockHeaderHash)
	return result
}

// 交易情报
type TransactionInfo struct {
	TxID         string                  // 交易ID
	TxCreateTime string                  // 交易创建时间
	TxCcID       string                  // 交易调用链码ID
	TxArgs       []string                // 交易输入参数
	TxReads      []*TransactionReadInfo  // 交易读取数据情报集合
	TxWrites     []*TransactionWriteInfo // 交易写入数据情报集合
	CallerMspID  string                  // 交易发起者MSPID
	CallerName   string                  // 交易发起者名称
	CallerOU     string                  // 交易发起者OU分组
	BlockNum     uint64                  // 交易所属区块编号
	TxDesc       string                  // 交易说明
	ErrorMsg     string                  // 交易解析错误消息
}

func (t *TransactionInfo) ToString() string {
	readSeys := []string{}
	for _, r := range t.TxReads {
		readSeys = append(readSeys, r.ToString())
	}
	writeSeys := []string{}
	for _, w := range t.TxWrites {
		writeSeys = append(writeSeys, w.ToString())
	}
	return fmt.Sprintf("TxID: %s, TxCreateTime: %s, BlockNum: %d, TxCcID: %s, TxArgs: %q, TxReads: %q, TxWrites: %q, CallerMspID: %s, CallerName: %s, CallerOU: %s, TxDesc: %s, ErrorMsg: %s",
		t.TxID, t.TxCreateTime, t.BlockNum, t.TxCcID, t.TxArgs, readSeys, writeSeys, t.CallerMspID, t.CallerName, t.CallerOU, t.TxDesc, t.ErrorMsg)
}

// 交易读取数据情报
type TransactionReadInfo struct {
	NameSpace        string // 所属链码
	ReadKey          string // 交易读取Key
	ReadBlockNum     uint64 // 交易读取区块编号
	ReadTxNumInBlock uint64 // 交易读取交易编号(区块内部)
}

func (t *TransactionReadInfo) ToString() string {
	return fmt.Sprintf("NameSpace: %s, ReadKey: %s, ReadBlockNum: %d, ReadTxNumInBlock: %d", t.NameSpace, t.ReadKey, t.ReadBlockNum, t.ReadTxNumInBlock)
}

// 交易写入数据情报
type TransactionWriteInfo struct {
	NameSpace  string // 所属链码
	WriteKey   string // 交易写入Key
	WriteValue string // 交易写入数据
	IsDelete   bool   // 是否删除
}

func (t *TransactionWriteInfo) ToString() string {
	return fmt.Sprintf("NameSpace: %s, WriteKey: %s, WriteValue: %s, IsDelete: %v", t.NameSpace, t.WriteKey, t.WriteValue, t.IsDelete)
}

// 浏览通道数据的相关参数
type BrowseChannelConfig struct {
	// 浏览上限类型
	//  0:使用BlockCountLimit作为区块浏览上限; 1:使用LastBlockHeaderHash作为区块浏览上限; 2:使用LastBlockNum作为区块浏览上限;
	BrowseLimitType int
	// 区块数量上限
	//  BrowseLimit值为0时，BrowseChannel浏览的区块数量<=BlockCountLimit。
	//  BlockCountLimit默认值为0，此时BrowseChannel浏览的区块数量无限制。
	BlockCountLimit uint64
	// 上回区块哈希
	//  BrowseLimitType值为1时，BrowseChannel浏览的区块向前不超过且不包括LastBlockHeaderHash对应的区块。
	//  LastBlockHeaderHash默认值为空。BrowseLimitType值为1时，LastBlockHeaderHash不可为空。
	LastBlockHeaderHash string
	// 上回区块编号
	//  BrowseLimit值为2时，BrowseChannel浏览的区块向前不超过且不包括LastBlockNum对应的区块。
	//  LastBlockNum默认值为0。
	LastBlockNum uint64
}

type BrowseOption func(*BrowseChannelConfig)

// BrowseChannel 浏览通道数据，遍历所有区块。
//  入参: ledgerClient 账本客户端实例
//  返回: ChannelInfo
func BrowseChannel(ledgerClient *ledger.Client) (*ChannelInfo, error) {
	config := &BrowseChannelConfig{}
	return BrowseChannelWithConfig(ledgerClient, config)
}

// BrowseChannelWithBlockCntLimit 浏览通道数据，根据入参blockCntLimit决定遍历区块的数量。
//  入参: ledgerClient 账本客户端实例
//  入参: blockCntLimit 区块数量上限
//  返回: ChannelInfo
func BrowseChannelWithBlockCntLimit(ledgerClient *ledger.Client, blockCntLimit uint64) (*ChannelInfo, error) {
	config := &BrowseChannelConfig{
		BrowseLimitType: 0,
		BlockCountLimit: blockCntLimit,
	}
	return BrowseChannelWithConfig(ledgerClient, config)
}

// BrowseChannelWithLastBlockHeaderHash 浏览通道数据，根据入参lastBlockHeaderHash决定遍历区块向前回溯的上限。
//  入参: ledgerClient 账本客户端实例
//  入参: lastBlockHeaderHash 前回浏览的最后区块头哈希
//  返回: ChannelInfo
func BrowseChannelWithLastBlockHeaderHash(ledgerClient *ledger.Client, lastBlockHeaderHash string) (*ChannelInfo, error) {
	config := &BrowseChannelConfig{
		BrowseLimitType:     1,
		LastBlockHeaderHash: lastBlockHeaderHash,
	}
	return BrowseChannelWithConfig(ledgerClient, config)
}

// BrowseChannelWithLastBlockNum 浏览通道数据，根据入参lastBlockNum决定遍历区块向前回溯的上限。
//  入参: ledgerClient 账本客户端实例
//  入参: lastBlockNum 前回浏览的最后区块编号
//  返回: ChannelInfo
func BrowseChannelWithLastBlockNum(ledgerClient *ledger.Client, lastBlockNum uint64) (*ChannelInfo, error) {
	config := &BrowseChannelConfig{
		BrowseLimitType: 2,
		LastBlockNum:    lastBlockNum,
	}
	return BrowseChannelWithConfig(ledgerClient, config)
}

// BrowseChannelWithConfig 浏览通道数据
//  入参: ledgerClient 账本客户端实例
//  入参: config 浏览参数
//  返回: ChannelInfo
func BrowseChannelWithConfig(ledgerClient *ledger.Client, config *BrowseChannelConfig) (*ChannelInfo, error) {
	// 检查浏览参数
	if config == nil {
		return nil, fmt.Errorf("no config(*BrowseChannelConfig)")
	}
	browseLimitType := config.BrowseLimitType
	// 检查浏览上限类型
	if browseLimitType < 0 || browseLimitType > 2 {
		return nil, fmt.Errorf("not supported browseLimitType")
	}
	var lastBlockHeaderHash []byte
	if browseLimitType == 1 {
		// BrowseLimitType值为1时，LastBlockHeaderHash不可为空。
		if config.LastBlockHeaderHash == "" {
			return nil, fmt.Errorf("lastBlockHeaderHash is empty")
		}
		var err error
		lastBlockHeaderHash, err = hex.DecodeString(config.LastBlockHeaderHash)
		// LastBlockHeaderHash解码失败
		if err != nil {
			return nil, err
		}
	}
	// 查询当前最新区块链信息
	blockChainInfo, err := ledgerClient.QueryInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get blockInfo: %s", err)
	}
	// 初始化通道情报
	channelInfo := &ChannelInfo{
		BlockHeight: blockChainInfo.BCI.Height,
	}
	// 交易总数，起始值为0
	var total uint64 = 0
	// 当前区块头哈希，起始值为最新区块头哈希
	curBlockHeaderHash := blockChainInfo.BCI.CurrentBlockHash
	// 区块情报(包含内部交易情报)集合
	blockInfoWithTxs := []*BlockInfoWithTx{}
	// 交易情报集合
	transactionInfos := []*TransactionInfo{}
	// 区块基础信息集合
	blockBasicInfos := []*BlockInfoBasic{}
	// 区块数量，起始值为0
	var blockCnt uint64 = 0
	// 区块编号，起始值为当前区块链高度
	curBlockNum := channelInfo.BlockHeight
	for {
		// 区块数量+1
		blockCnt++
		// 区块编号-1
		curBlockNum--
		// 当浏览参数为使用区块数量限制时，检查区块数量是否已超过区块数量上限。
		// BlockCountLimit为0时表示没有区块数量限制。
		if browseLimitType == 0 && config.BlockCountLimit > 0 && config.BlockCountLimit < blockCnt {
			break
		}
		// 当浏览参数为使用前回区块哈希限制时，检查本次遍历的区块哈希
		if browseLimitType == 1 && bytes.Equal(curBlockHeaderHash, lastBlockHeaderHash) {
			break
		}
		// 当浏览参数为使用前回区块编号限制时，检查本次遍历的区块编号
		if browseLimitType == 2 && config.LastBlockNum == curBlockNum {
			break
		}
		// 使用当前区块头哈希查询当前区块
		block, err := ledgerClient.QueryBlockByHash(curBlockHeaderHash)
		if err != nil {
			return nil, fmt.Errorf("failed to QueryBlockByHash: %s", err)
		}
		// 反序列化当前区块
		blockInfo, err := UnmarshalBlockData(block, curBlockHeaderHash)
		if err != nil {
			return nil, fmt.Errorf("failed to UnmarshalBlockData: %s", err)
		}
		// 累加交易数量
		total += blockInfo.TransCnt
		// 追加区块情报(包含内部交易情报)
		blockInfoWithTxs = append(blockInfoWithTxs, blockInfo)
		// 追加交易情报
		transactionInfos = append(transactionInfos, blockInfo.TransactionInfos...)
		// 追加区块基础信息
		blockBasicInfos = append(blockBasicInfos, blockInfo.GetBasicInfo())
		// 获取前区块头哈希
		curBlockHeaderHash = block.Header.PreviousHash
		// 前区块头哈希为空时，表示已经遍历到第一个区块
		if len(curBlockHeaderHash) == 0 {
			break
		}
	}
	channelInfo.TransTotal = total
	channelInfo.BlockInfoWithTxs = blockInfoWithTxs
	channelInfo.TransactionInfos = transactionInfos
	channelInfo.BlockBasicInfos = blockBasicInfos
	return channelInfo, nil
}

// UnmarshalBlockData 反序列化Block区块数据。
//  入参: block 区块数据
//  入参: curBlockHash 当前区块头哈希
//  返回: BlockInfo
func UnmarshalBlockData(block *common.Block, curBlockHash []byte) (*BlockInfoWithTx, error) {
	// 区块内交易数据集合
	tranDatas := block.Data.Data
	// 区块内交易数量
	transCnt := len(tranDatas)
	// 准备区块情报
	blockInfo := &BlockInfoWithTx{
		BlockInfoBasic: BlockInfoBasic{
			BlockNum:           block.Header.Number,
			BlockHeaderHash:    hex.EncodeToString(curBlockHash),
			PreBlockHeaderHash: hex.EncodeToString(block.Header.PreviousHash),
			TransCnt:           uint64(transCnt),
		},
	}
	// zclog.Debugf("区块编号: %d, 交易数量: %d", blockInfo.BlockNum, blockInfo.TransCnt)
	transactionInfos := []*TransactionInfo{}
	// 遍历区块内所有交易
	for i := 0; i < transCnt; i++ {
		// 创建交易情报
		transactionInfo := &TransactionInfo{
			BlockNum: blockInfo.BlockNum,
		}
		transactionInfos = append(transactionInfos, transactionInfo)
		// zclog.Debugf("第 %d 条交易数据.", i+1)

		/* 初步反序列化区块里的本条交易数据，获取payload */
		// 交易数据反序列化为 Envelope
		envelope := &common.Envelope{}
		err := proto.Unmarshal(tranDatas[i], envelope)
		if err != nil {
			transactionInfo.ErrorMsg = err.Error()
			continue
		}
		// 反序列化 envelope.Payload
		payload := &common.Payload{}
		err = proto.Unmarshal(envelope.Payload, payload)
		if err != nil {
			transactionInfo.ErrorMsg = err.Error()
			continue
		}
		// zclog.Debugf("第 %d 条交易数据 payload: %s", i+1, payload.String())

		/* 从payload的header里获取交易ID、交易创建时间、交易发起者的MSPID/CommonName/OU信息 */
		// 反序列化 payload.Header.ChannelHeader 交易ID、交易创建时间等
		channelHeader := &common.ChannelHeader{}
		err = proto.Unmarshal(payload.Header.ChannelHeader, channelHeader)
		if err != nil {
			transactionInfo.ErrorMsg = err.Error()
			continue
		}
		transactionInfo.TxID = channelHeader.TxId
		transactionInfo.TxCreateTime = time.Unix(channelHeader.Timestamp.Seconds, 0).Format("2006-01-02 15:04:05")
		// zclog.Debugf("第 %d 条交易数据 channelHeader: %s", i+1, channelHeader.String())
		// 反序列化 payload.Header.SignatureHeader 发起交易请求的身份信息(字节数组)
		signatureHeader := &common.SignatureHeader{}
		err = proto.Unmarshal(payload.Header.SignatureHeader, signatureHeader)
		if err != nil {
			transactionInfo.ErrorMsg = err.Error()
			continue
		}
		// 反序列化 signatureHeader.Creator 发起交易请求的身份信息, 包括MSPID，以及身份字节数组IdBytes
		creator := &msp.SerializedIdentity{}
		err = proto.Unmarshal(signatureHeader.Creator, creator)
		if err != nil {
			transactionInfo.ErrorMsg = err.Error()
			continue
		}
		transactionInfo.CallerMspID = creator.GetMspid()
		// 证书的pem字节数组解析为x509证书结构
		cert, err := x509.ReadCertificateFromPem(creator.IdBytes)
		if err != nil {
			transactionInfo.ErrorMsg = err.Error()
			continue
		}
		// zclog.Debugf("cert owner: %s", cert.Subject)
		transactionInfo.CallerName = cert.Subject.CommonName
		transactionInfo.CallerOU = cert.Subject.OrganizationalUnit[0]

		/* 从payload的payload.Data里进一步获取 ChaincodeActionPayload */
		// 反序列化 payload.Data
		transaction := &peer.Transaction{}
		err = proto.Unmarshal(payload.Data, transaction)
		if err != nil {
			transactionInfo.ErrorMsg = err.Error()
			continue
		}
		// zclog.Debugf("transaction: %s", transaction.String())
		// 反序列化 transaction.Actions[0].Payload
		chaincodeActionPayload := &peer.ChaincodeActionPayload{}
		err = proto.Unmarshal(transaction.Actions[0].Payload, chaincodeActionPayload)
		if err != nil {
			transactionInfo.ErrorMsg = err.Error()
			continue
		}
		// zclog.Debugf("chaincodeActionPayload: %s", chaincodeActionPayload.String())

		/* 从ChaincodeActionPayload里获取 链码以及本次合约调用的入参 */
		// 反序列化 chaincodeActionPayload.ChaincodeProposalPayload
		chaincodeProposalPayload := &peer.ChaincodeProposalPayload{}
		err = proto.Unmarshal(chaincodeActionPayload.ChaincodeProposalPayload, chaincodeProposalPayload)
		if err != nil {
			transactionInfo.ErrorMsg = err.Error()
			continue
		}
		if chaincodeProposalPayload == nil || len(chaincodeProposalPayload.Input) == 0 {
			transactionInfo.TxDesc = fmt.Sprintf("区块编号: %d, 第 %d 条交易不是合约调用。", blockInfo.BlockNum, i+1)
			continue
		}
		// zclog.Debugf("chaincodeProposalPayload: %s", chaincodeProposalPayload.String())
		// 反序列化 chaincodeProposalPayload.Input
		chaincodeInvocationSpec := &peer.ChaincodeInvocationSpec{}
		err = proto.Unmarshal(chaincodeProposalPayload.Input, chaincodeInvocationSpec)
		if err != nil {
			transactionInfo.ErrorMsg = err.Error()
			continue
		}
		// zclog.Debugf("chaincodeInvocationSpec: %s", chaincodeInvocationSpec.String())
		if chaincodeInvocationSpec != nil && chaincodeInvocationSpec.ChaincodeSpec != nil {
			if chaincodeInvocationSpec.ChaincodeSpec.Input != nil {
				// 获取本次交易的链码调用输入参数
				var args []string
				for _, v := range chaincodeInvocationSpec.ChaincodeSpec.Input.Args {
					args = append(args, string(v))
				}
				transactionInfo.TxArgs = args
			}
			if chaincodeInvocationSpec.ChaincodeSpec.ChaincodeId != nil {
				transactionInfo.TxCcID = chaincodeInvocationSpec.ChaincodeSpec.ChaincodeId.Name
			}
		} else {
			transactionInfo.ErrorMsg = fmt.Sprintf("区块编号: %d, 第 %d 条交易, chaincodeInvocationSpec是预期外的值: %s", blockInfo.BlockNum, i+1, chaincodeInvocationSpec.String())
			continue
		}

		/* 从ChaincodeActionPayload里获取 本次合约调用的读写集 */
		proposalResponsePayloadTmp := string(chaincodeActionPayload.Action.ProposalResponsePayload)
		// proposalResponsePayloadTmp的值为"Application"或"Orderer"时，代表当前交易是通道配置交易等非业务交易。
		if proposalResponsePayloadTmp == "Application" || proposalResponsePayloadTmp == "Orderer" {
			transactionInfo.TxDesc = fmt.Sprintf("区块编号: %d, 第 %d 条交易不是合约调用。", blockInfo.BlockNum, i+1)
			continue
		}
		// 反序列化 chaincodeActionPayload.Action 数据
		proposalResponsePayload := &peer.ProposalResponsePayload{}
		err = proto.Unmarshal(chaincodeActionPayload.Action.ProposalResponsePayload, proposalResponsePayload)
		if err != nil {
			transactionInfo.ErrorMsg = err.Error()
			continue
		}
		// zclog.Debugf("proposalResponsePayload: %s", proposalResponsePayload.String())
		// 反序列化 proposalResponsePayload.Extension
		chaincodeAction := &peer.ChaincodeAction{}
		err = proto.Unmarshal(proposalResponsePayload.Extension, chaincodeAction)
		if err != nil {
			transactionInfo.ErrorMsg = err.Error()
			continue
		}
		// zclog.Debugf("chaincodeAction: %s", chaincodeAction.String())
		// 反序列化 chaincodeAction.Results
		txReadWriteSet := &rwset.TxReadWriteSet{}
		err = proto.Unmarshal(chaincodeAction.Results, txReadWriteSet)
		if err != nil {
			transactionInfo.ErrorMsg = err.Error()
			continue
		}
		// zclog.Debugf("txReadWriteSet: %s", txReadWriteSet.String())
		transactionReadInfos := []*TransactionReadInfo{}
		transactionWriteInfos := []*TransactionWriteInfo{}
		errInRwset := []string{}
		// 遍历 txReadWriteSet.NsRwset
		for _, v := range txReadWriteSet.NsRwset {
			// 不处理 _lifecycle
			if v.Namespace == "_lifecycle" {
				continue
			}
			// zclog.Debugf("Namespace: %s", v.Namespace)
			readWriteSet := &kvrwset.KVRWSet{}
			err = proto.Unmarshal(v.Rwset, readWriteSet)
			if err != nil {
				errInRwset = append(errInRwset, err.Error())
				continue
			}
			for _, r := range readWriteSet.Reads {
				transactionReadInfo := &TransactionReadInfo{
					NameSpace: v.Namespace,
					ReadKey:   TrimUnknownHeader(r.Key),
				}
				if r.Version != nil {
					transactionReadInfo.ReadBlockNum = r.Version.BlockNum
					transactionReadInfo.ReadTxNumInBlock = r.Version.TxNum
				}
				transactionReadInfos = append(transactionReadInfos, transactionReadInfo)
			}
			for _, w := range readWriteSet.Writes {
				// zclog.Debugf("写集 key: %s, value: %s, IsDelete: %v", w.GetKey(), string(w.GetValue()), w.GetIsDelete())
				transactionWriteInfo := &TransactionWriteInfo{
					NameSpace:  v.Namespace,
					WriteKey:   TrimUnknownHeader(w.GetKey()),
					WriteValue: string(w.GetValue()),
					IsDelete:   w.GetIsDelete(),
				}
				transactionWriteInfos = append(transactionWriteInfos, transactionWriteInfo)
			}
		}
		if len(errInRwset) > 0 {
			transactionInfo.ErrorMsg = strings.Join(errInRwset, ";")
		}
		transactionInfo.TxReads = transactionReadInfos
		transactionInfo.TxWrites = transactionWriteInfos

	}
	blockInfo.TransactionInfos = transactionInfos
	return blockInfo, nil
}

func TrimHiddenCharacter(originStr string) string {
	srcRunes := []rune(originStr)
	dstRunes := make([]rune, 0, len(srcRunes))
	for _, c := range srcRunes {
		if c >= 0 && c <= 31 {
			continue
		}
		if c == 127 {
			continue
		}
		dstRunes = append(dstRunes, c)
	}
	return string(dstRunes)
}

// TrimUnknownHeader 去除不能正常解析的头部字节切片[0, 244, 143, 191, 191]
func TrimUnknownHeader(origin string) string {
	arrIn := []byte(origin)
	if len(arrIn) < 5 {
		return origin
	}
	// 0, 244, 143, 191, 191
	if arrIn[0] == 0 && arrIn[1] == 244 && arrIn[2] == 143 && arrIn[3] == 191 && arrIn[4] == 191 {
		return string(arrIn[5:])
	}
	return origin
}
