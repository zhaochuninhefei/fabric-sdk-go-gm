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

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

	"gitee.com/zhaochuninhefei/fabric-protos-go-gm/common"
	"gitee.com/zhaochuninhefei/fabric-protos-go-gm/ledger/rwset"
	"gitee.com/zhaochuninhefei/fabric-protos-go-gm/ledger/rwset/kvrwset"
	"gitee.com/zhaochuninhefei/fabric-protos-go-gm/msp"
	"gitee.com/zhaochuninhefei/fabric-protos-go-gm/peer"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/pkg/client/ledger"
	"gitee.com/zhaochuninhefei/gmgo/x509"
	"gitee.com/zhaochuninhefei/zcgolog/zclog"
	"github.com/golang/protobuf/proto"
)

/*
pkg/util/chn_browse_util/chn_browse_util.go 通道浏览工具库，提供用于查询通道的区块与交易信息的通用函数。
*/

// 通道情报
type ChannelInfo struct {
	BlockHeight      uint64             // 区块高度
	TransTotal       uint64             // 交易总数
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
	result := fmt.Sprintf("区块编号: %d, 交易数量: %d, 区块哈希: %s, 前区块哈希: %s,\n\t\t交易集合:\n",
		t.BlockNum, t.TransCnt, t.BlockHash, t.PreBlockHash)
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
	BlockNum     uint64 // 区块编号
	BlockHash    string // 区块哈希(16进制字符串)
	PreBlockHash string // 前一区块哈希(16进制字符串)
	TransCnt     uint64 // 区块内交易数量
}

func (t *BlockInfoBasic) ToString() string {
	result := fmt.Sprintf("区块编号: %d, 交易数量: %d, 区块哈希: %s, 前区块哈希: %s",
		t.BlockNum, t.TransCnt, t.BlockHash, t.PreBlockHash)
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
	return fmt.Sprintf("TxID: %s, TxCreateTime: %s, TxCcID: %s, TxArgs: %q, TxReads: %q, TxWrites: %q, CallerMspID: %s, CallerName: %s, CallerOU: %s",
		t.TxID, t.TxCreateTime, t.TxCcID, t.TxArgs, readSeys, writeSeys, t.CallerMspID, t.CallerName, t.CallerOU)
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
	//  0:使用BlockCountLimit作为区块浏览上限; 1:使用LastBlockHash作为区块浏览上限; 2:使用LastBlockNum作为区块浏览上限;
	BrowseLimitType int
	// 区块数量上限
	//  BrowseLimit值为0时，BrowseChannel浏览的区块数量<=BlockCountLimit。
	//  BlockCountLimit默认值为0，此时BrowseChannel浏览的区块数量无限制。
	BlockCountLimit uint64
	// 上回区块哈希
	//  BrowseLimitType值为1时，BrowseChannel浏览的区块向前不超过且不包括LastBlockHash对应的区块。
	//  LastBlockHash默认值为空。BrowseLimitType值为1时，LastBlockHash不可为空。
	LastBlockHash string
	// 上回区块编号
	//  BrowseLimit值为2时，BrowseChannel浏览的区块向前不超过且不包括LastBlockNum对应的区块。
	//  LastBlockNum默认值为0。
	LastBlockNum uint64
}

type BrowseOption func(*BrowseChannelConfig)

// BrowseChannel 浏览通道数据
//  入参: ledgerClient 账本客户端实例
//  返回: ChannelInfo
func BrowseChannel(ledgerClient *ledger.Client) (*ChannelInfo, error) {
	blockChainInfo, err := ledgerClient.QueryInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get blockInfo: %s", err)
	}
	channelInfo := &ChannelInfo{
		BlockHeight: blockChainInfo.BCI.Height,
	}

	var total uint64 = 0
	curBlockHash := blockChainInfo.BCI.CurrentBlockHash
	blockInfoWithTxs := []*BlockInfoWithTx{}
	transactionInfos := []*TransactionInfo{}
	blockBasicInfos := []*BlockInfoBasic{}
	for {
		block, err := ledgerClient.QueryBlockByHash(curBlockHash)
		if err != nil {
			return nil, fmt.Errorf("failed to QueryBlockByHash: %s", err)
		}
		blockInfo, err := UnmarshalBlockData(block, curBlockHash)
		if err != nil {
			return nil, fmt.Errorf("failed to UnmarshalBlockData: %s", err)
		}
		total += blockInfo.TransCnt
		blockInfoWithTxs = append(blockInfoWithTxs, blockInfo)
		transactionInfos = append(transactionInfos, blockInfo.TransactionInfos...)
		blockBasicInfos = append(blockBasicInfos, blockInfo.GetBasicInfo())
		curBlockHash = block.Header.PreviousHash
		if len(curBlockHash) == 0 {
			break
		}
	}
	channelInfo.TransTotal = total
	channelInfo.BlockInfoWithTxs = blockInfoWithTxs
	channelInfo.TransactionInfos = transactionInfos
	channelInfo.BlockBasicInfos = blockBasicInfos
	return channelInfo, nil
}

// BrowseChannel 浏览通道数据
//  入参: ledgerClient 账本客户端实例
//  返回: ChannelInfo
func BrowseChannelWithConfig(ledgerClient *ledger.Client, config *BrowseChannelConfig) (*ChannelInfo, error) {
	if config == nil {
		return nil, fmt.Errorf("no config(*BrowseChannelConfig)")
	}
	browseLimitType := config.BrowseLimitType
	if browseLimitType < 0 || browseLimitType > 2 {
		return nil, fmt.Errorf("not supported browseLimitType")
	}
	lastBlockHash, _ := hex.DecodeString(config.LastBlockHash)

	blockChainInfo, err := ledgerClient.QueryInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get blockInfo: %s", err)
	}
	channelInfo := &ChannelInfo{
		BlockHeight: blockChainInfo.BCI.Height,
	}

	var total uint64 = 0
	curBlockHash := blockChainInfo.BCI.CurrentBlockHash
	blockInfoWithTxs := []*BlockInfoWithTx{}
	transactionInfos := []*TransactionInfo{}
	blockBasicInfos := []*BlockInfoBasic{}
	var blockCnt uint64 = 0
	for {
		blockCnt++
		// 当浏览参数为使用区块数量限制时，检查区块数量是否已超过区块数量上限
		if browseLimitType == 0 && config.BlockCountLimit <= blockCnt {
			break
		}
		// 当浏览参数为使用前回区块哈希限制时，检查本次遍历的区块哈希
		if browseLimitType == 1 && bytes.Equal(curBlockHash, lastBlockHash) {
			break
		}
		block, err := ledgerClient.QueryBlockByHash(curBlockHash)
		if err != nil {
			return nil, fmt.Errorf("failed to QueryBlockByHash: %s", err)
		}
		// 当浏览参数为使用前回区块编号限制时，检查本次遍历的区块编号
		if browseLimitType == 2 && config.LastBlockNum == block.Header.Number {
			break
		}
		blockInfo, err := UnmarshalBlockData(block, curBlockHash)
		if err != nil {
			return nil, fmt.Errorf("failed to UnmarshalBlockData: %s", err)
		}
		total += blockInfo.TransCnt
		blockInfoWithTxs = append(blockInfoWithTxs, blockInfo)
		transactionInfos = append(transactionInfos, blockInfo.TransactionInfos...)
		blockBasicInfos = append(blockBasicInfos, blockInfo.GetBasicInfo())
		curBlockHash = block.Header.PreviousHash
		if len(curBlockHash) == 0 {
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
//  返回: BlockInfo
func UnmarshalBlockData(block *common.Block, curBlockHash []byte) (*BlockInfoWithTx, error) {
	// 区块内交易数据集合
	tranDatas := block.Data.Data
	// 区块内交易数量
	transCnt := len(tranDatas)
	// 准备区块情报
	blockInfo := &BlockInfoWithTx{
		BlockInfoBasic: BlockInfoBasic{
			BlockNum:     block.Header.Number,
			BlockHash:    hex.EncodeToString(curBlockHash),
			PreBlockHash: hex.EncodeToString(block.Header.PreviousHash),
			TransCnt:     uint64(transCnt),
		},
	}
	zclog.Debugf("区块编号: %d, 交易数量: %d", blockInfo.BlockNum, blockInfo.TransCnt)
	transactionInfos := []*TransactionInfo{}
	// 遍历区块内所有交易
	for i := 0; i < transCnt; i++ {
		// 创建交易情报
		transactionInfo := &TransactionInfo{}
		zclog.Debugf("第 %d 条交易数据.", i+1)

		/*
		*初步反序列化区块里的本条交易数据，获取payload
		 */
		// 交易数据反序列化为 Envelope
		envelope := &common.Envelope{}
		err := proto.Unmarshal(tranDatas[i], envelope)
		if err != nil {
			return blockInfo, err
		}
		// 反序列化 envelope.Payload
		payload := &common.Payload{}
		err = proto.Unmarshal(envelope.Payload, payload)
		if err != nil {
			return blockInfo, err
		}
		// zclog.Debugf("第 %d 条交易数据 payload: %s", i+1, payload.String())

		/*
		*从payload的header里获取交易ID、交易创建时间、交易发起者的MSPID/CommonName/OU信息
		 */
		// 反序列化 payload.Header.ChannelHeader 交易ID、交易创建时间等
		channelHeader := &common.ChannelHeader{}
		err = proto.Unmarshal(payload.Header.ChannelHeader, channelHeader)
		if err != nil {
			return blockInfo, err
		}
		transactionInfo.TxID = channelHeader.TxId
		transactionInfo.TxCreateTime = time.Unix(channelHeader.Timestamp.Seconds, 0).Format("2006-01-02 15:04:05")
		// zclog.Debugf("第 %d 条交易数据 channelHeader: %s", i+1, channelHeader.String())
		// 反序列化 payload.Header.SignatureHeader 发起交易请求的身份信息(字节数组)
		signatureHeader := &common.SignatureHeader{}
		err = proto.Unmarshal(payload.Header.SignatureHeader, signatureHeader)
		if err != nil {
			return blockInfo, err
		}
		// 反序列化 signatureHeader.Creator 发起交易请求的身份信息, 包括MSPID，以及身份字节数组IdBytes
		creator := &msp.SerializedIdentity{}
		err = proto.Unmarshal(signatureHeader.Creator, creator)
		if err != nil {
			return blockInfo, err
		}
		transactionInfo.CallerMspID = creator.GetMspid()
		// 证书的pem字节数组解析为x509证书结构
		cert, err := x509.ReadCertificateFromPem(creator.IdBytes)
		if err != nil {
			return blockInfo, err
		}
		zclog.Debugf("cert owner: %s", cert.Subject)
		transactionInfo.CallerName = cert.Subject.CommonName
		transactionInfo.CallerOU = cert.Subject.OrganizationalUnit[0]

		/*
		*从payload的payload.Data里进一步获取 ChaincodeActionPayload
		 */
		// 反序列化 payload.Data
		transaction := &peer.Transaction{}
		err = proto.Unmarshal(payload.Data, transaction)
		if err != nil {
			return blockInfo, err
		}
		// zclog.Debugf("transaction: %s", transaction.String())
		// 反序列化 transaction.Actions[0].Payload
		chaincodeActionPayload := &peer.ChaincodeActionPayload{}
		err = proto.Unmarshal(transaction.Actions[0].Payload, chaincodeActionPayload)
		if err != nil {
			return blockInfo, err
		}
		// zclog.Debugf("chaincodeActionPayload: %s", chaincodeActionPayload.String())

		/*
		*从ChaincodeActionPayload里获取 链码以及本次合约调用的入参
		 */
		// 反序列化 chaincodeActionPayload.ChaincodeProposalPayload
		chaincodeProposalPayload := &peer.ChaincodeProposalPayload{}
		err = proto.Unmarshal(chaincodeActionPayload.ChaincodeProposalPayload, chaincodeProposalPayload)
		if err != nil {
			return blockInfo, err
		}
		zclog.Debugf("chaincodeProposalPayload: %s", chaincodeProposalPayload.String())
		// 反序列化 chaincodeProposalPayload.Input
		chaincodeInvocationSpec := &peer.ChaincodeInvocationSpec{}
		err = proto.Unmarshal(chaincodeProposalPayload.Input, chaincodeInvocationSpec)
		if err != nil {
			return blockInfo, err
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
		}

		/*
		*从ChaincodeActionPayload里获取 本次合约调用的读写集
		 */
		proposalResponsePayloadTmp := string(chaincodeActionPayload.Action.ProposalResponsePayload)
		if proposalResponsePayloadTmp != "Application" {
			// 反序列化 chaincodeActionPayload.Action 数据
			proposalResponsePayload := &peer.ProposalResponsePayload{}
			err = proto.Unmarshal(chaincodeActionPayload.Action.ProposalResponsePayload, proposalResponsePayload)
			if err != nil {
				return blockInfo, err
			}
			// zclog.Debugf("proposalResponsePayload: %s", proposalResponsePayload.String())
			// 反序列化 proposalResponsePayload.Extension
			chaincodeAction := &peer.ChaincodeAction{}
			err = proto.Unmarshal(proposalResponsePayload.Extension, chaincodeAction)
			if err != nil {
				return blockInfo, err
			}
			// zclog.Debugf("chaincodeAction: %s", chaincodeAction.String())
			// 反序列化 chaincodeAction.Results
			txReadWriteSet := &rwset.TxReadWriteSet{}
			err = proto.Unmarshal(chaincodeAction.Results, txReadWriteSet)
			if err != nil {
				return blockInfo, err
			}
			// zclog.Debugf("txReadWriteSet: %s", txReadWriteSet.String())
			transactionReadInfos := []*TransactionReadInfo{}
			transactionWriteInfos := []*TransactionWriteInfo{}
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
					return blockInfo, err
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
			transactionInfo.TxReads = transactionReadInfos
			transactionInfo.TxWrites = transactionWriteInfos
		}

		transactionInfos = append(transactionInfos, transactionInfo)
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
