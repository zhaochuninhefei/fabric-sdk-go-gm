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
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
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
	BlockHeight uint64       // 区块高度
	TransTotal  uint64       // 交易总数
	BlockInfos  []*BlockInfo // 区块集合
}

// 区块情报
type BlockInfo struct {
	BlockNum         uint64             // 区块编号
	BlockHash        string             // 区块哈希(16进制字符串)
	PreBlockHash     string             // 前一区块哈希(16进制字符串)
	TransCnt         uint64             // 区块内交易数量
	TransactionInfos []*TransactionInfo // 区块内交易情报集合
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

// 交易读取数据情报
type TransactionReadInfo struct {
	NameSpace        string // 所属链码
	ReadKey          string // 交易读取Key
	ReadBlockNum     uint64 // 交易读取区块编号
	ReadTxNumInBlock uint64 // 交易读取交易编号(区块内部)
}

// 交易写入数据情报
type TransactionWriteInfo struct {
	NameSpace  string // 所属链码
	WriteKey   string // 交易写入Key
	WriteValue string // 交易写入数据
	IsDelete   bool   // 是否删除
}

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
	blockInfos := []*BlockInfo{}
	for {
		block, err := ledgerClient.QueryBlockByHash(curBlockHash)
		if err != nil {
			return nil, fmt.Errorf("failed to QueryBlockByHash: %s", err)
		}
		blockInfo, err := UnmarshalBlockData(block)
		if err != nil {
			return nil, fmt.Errorf("failed to UnmarshalBlockData: %s", err)
		}
		total += blockInfo.TransCnt
		blockInfos = append(blockInfos, blockInfo)
		curBlockHash = block.Header.PreviousHash
		if len(curBlockHash) == 0 {
			break
		}
	}
	channelInfo.TransTotal = total
	channelInfo.BlockInfos = blockInfos
	return channelInfo, nil
}

func UnmarshalBlockData(block *common.Block) (*BlockInfo, error) {
	// 区块内交易数据集合
	tranDatas := block.Data.Data
	// 区块内交易数量
	transCnt := len(tranDatas)
	// 准备区块情报
	blockInfo := &BlockInfo{
		BlockNum:     block.Header.Number,
		BlockHash:    hex.EncodeToString(block.Header.DataHash),
		PreBlockHash: hex.EncodeToString(block.Header.PreviousHash),
		TransCnt:     uint64(transCnt),
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
		// // 对 creator.IdBytes 做base64解码，得到证书的pem字节数组
		// idBase64Str := base64.URLEncoding.EncodeToString([]byte(creator.IdBytes))
		// certPem, err := base64.URLEncoding.DecodeString(idBase64Str)
		// if err != nil {
		// 	return blockInfo, err
		// }
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
						ReadKey:   r.Key,
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
						WriteKey:   w.GetKey(),
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

func showBlockData(ledgerClient *ledger.Client, blockNumber uint64) error {
	if blockNumber == 0 {
		blockInfo, err := ledgerClient.QueryInfo()
		if err != nil {
			zclog.Errorf("Failed to get blockInfo: %s", err)
			os.Exit(1)
		}
		blockNumber = blockInfo.BCI.Height - 1
	}

	block, err := ledgerClient.QueryBlock(blockNumber)
	if err != nil {
		return err
	}
	transCnt := len(block.Data.Data)
	zclog.Infof("区块编号: %d, 交易数: %d", block.Header.Number, transCnt)

	for i := 0; i < transCnt; i++ {
		zclog.Infof("第 %d 条交易数据.", i+1)
		// 交易数据反序列化为 Envelope
		envelope := &common.Envelope{}
		err = proto.Unmarshal(block.GetData().Data[i], envelope)
		if err != nil {
			return err
		}
		// zclog.Infof("第 %d 条交易数据 envelope: %s", i+1, envelope.String())
		// 反序列化 envelope.Payload
		payload := &common.Payload{}
		err = proto.Unmarshal(envelope.Payload, payload)
		if err != nil {
			return err
		}
		// zclog.Infof("第 %d 条交易数据 payload: %s", i+1, payload.String())

		// 反序列化 payload.Header.ChannelHeader 交易ID、通道ID、交易创建时间等
		channelHeader := &common.ChannelHeader{}
		err = proto.Unmarshal(payload.Header.ChannelHeader, channelHeader)
		if err != nil {
			return err
		}
		// zclog.Infof("第 %d 条交易数据 channelHeader: %s", i+1, channelHeader.String())
		// 反序列化 payload.Header.SignatureHeader 发起交易请求的身份信息(字节数组)
		signatureHeader := &common.SignatureHeader{}
		err = proto.Unmarshal(payload.Header.SignatureHeader, signatureHeader)
		if err != nil {
			return err
		}
		// 反序列化 signatureHeader.Creator 发起交易请求的身份信息, 包括MSPID，以及身份字节数组IdBytes
		creator := &msp.SerializedIdentity{}
		err = proto.Unmarshal(signatureHeader.Creator, creator)
		if err != nil {
			return err
		}
		// 对 creator.IdBytes 做base64解码，得到证书的pem字节数组
		idBase64Str := base64.URLEncoding.EncodeToString([]byte(creator.IdBytes))
		certPem, err := base64.URLEncoding.DecodeString(idBase64Str)
		if err != nil {
			return err
		}
		// 证书的pem字节数组解析为x509证书结构
		cert, err := x509.ReadCertificateFromPem(certPem)
		if err != nil {
			return err
		}
		zclog.Infof("cert owner: %s", cert.Subject)

		// 反序列化 payload.Data
		transaction := &peer.Transaction{}
		err = proto.Unmarshal(payload.Data, transaction)
		if err != nil {
			return err
		}
		// zclog.Infof("transaction: %s", transaction.String())

		chaincodeActionPayload := &peer.ChaincodeActionPayload{}
		err = proto.Unmarshal(transaction.Actions[0].Payload, chaincodeActionPayload)
		if err != nil {
			return err
		}
		// zclog.Infof("chaincodeActionPayload: %s", chaincodeActionPayload.String())

		// TODO 尝试获取 chaincodeActionPayload.Action 数据
		proposalResponsePayload := &peer.ProposalResponsePayload{}
		err = proto.Unmarshal(chaincodeActionPayload.Action.ProposalResponsePayload, proposalResponsePayload)
		if err != nil {
			return err
		}
		// zclog.Infof("proposalResponsePayload: %s", proposalResponsePayload.String())

		chaincodeAction := &peer.ChaincodeAction{}
		err = proto.Unmarshal(proposalResponsePayload.Extension, chaincodeAction)
		if err != nil {
			return err
		}
		// zclog.Infof("chaincodeAction: %s", chaincodeAction.String())

		txReadWriteSet := &rwset.TxReadWriteSet{}
		err = proto.Unmarshal(chaincodeAction.Results, txReadWriteSet)
		if err != nil {
			return err
		}
		// zclog.Infof("txReadWriteSet: %s", txReadWriteSet.String())

		for _, v := range txReadWriteSet.NsRwset {
			if v.Namespace == "_lifecycle" {
				continue
			}
			zclog.Infof("Namespace: %s", v.Namespace)
			readWriteSet := &kvrwset.KVRWSet{}
			err = proto.Unmarshal(v.Rwset, readWriteSet)
			if err != nil {
				return err
			}

			for _, r := range readWriteSet.Reads {
				// readSetJsonStr, err := json.Marshal(r)
				// if err != nil {
				// 	return err
				// }
				// zclog.Infof("readSetJsonStr: %s", readSetJsonStr)
				// zclog.Infof("读集 key:: %v", []byte(r.Key))
				zclog.Infof("读集 key: %s, 区块编号: %d, 交易编号: %d", TrimHiddenCharacter(r.Key), r.Version.BlockNum, r.Version.TxNum)
			}
			for _, w := range readWriteSet.Writes {
				// writeSetItem := map[string]interface{}{
				// 	"Key":      w.GetKey(),
				// 	"Value":    string(w.GetValue()),
				// 	"IsDelete": w.GetIsDelete(),
				// }
				// writeSetJsonStr, err := json.Marshal(writeSetItem)
				// if err != nil {
				// 	return err
				// }
				// zclog.Infof("writeSetJsonStr: %s", writeSetJsonStr)
				zclog.Infof("写集 key: %s, value: %s, IsDelete: %v", w.GetKey(), string(w.GetValue()), w.GetIsDelete())
			}

		}

		// 获取交易对应的链码调用参数
		chaincodeProposalPayload := &peer.ChaincodeProposalPayload{}
		err = proto.Unmarshal(chaincodeActionPayload.ChaincodeProposalPayload, chaincodeProposalPayload)
		if err != nil {
			return err
		}
		// zclog.Infof("chaincodeProposalPayload: %s", chaincodeProposalPayload.String())

		chaincodeInvocationSpec := &peer.ChaincodeInvocationSpec{}
		err = proto.Unmarshal(chaincodeProposalPayload.Input, chaincodeInvocationSpec)
		if err != nil {
			return err
		}
		// zclog.Infof("chaincodeInvocationSpec: %s", chaincodeInvocationSpec.String())
		zclog.Infof("ChaincodeId.Name: %s, Input.Args: %q", chaincodeInvocationSpec.ChaincodeSpec.ChaincodeId.Name, chaincodeInvocationSpec.ChaincodeSpec.Input.Args)

		// transactionAction := &peer.TransactionAction{}
		// err = proto.Unmarshal(payload.Data, transactionAction)
		// if err != nil {
		// 	return err
		// }
		// zclog.Infof("transactionAction: %s", transactionAction.String())
		// TODO transactionAction.Payload 没有数据。。。
		// TODO 调查 TransactionAction.Header 对应的proto类型，尚未找到
		// transactionAction.Header 用 common.SignatureHeader 反序列化的结果不对劲，nonce里放着入参与交易后数据。。。
		// signatureHeaderInTransactionAction := &common.SignatureHeader{}
		// err = proto.Unmarshal(transactionAction.Header, signatureHeaderInTransactionAction)
		// if err != nil {
		// 	return err
		// }
		// zclog.Infof("signatureHeaderInTransactionAction: %s", signatureHeaderInTransactionAction.String())
		// chainCodeActionPayload := &peer.ChaincodeActionPayload{}
		// // err = proto.Unmarshal(transactionAction.Payload, chainCodeActionPayload)
		// err = proto.Unmarshal(transactionAction.Header, chainCodeActionPayload)
		// if err != nil {
		// 	return err
		// }
		// zclog.Infof("chainCodeActionPayload: %s", chainCodeActionPayload.String())

		// chaincodeProposalPayload := &peer.ChaincodeProposalPayload{}
		// err = proto.Unmarshal(chainCodeActionPayload.ChaincodeProposalPayload, chaincodeProposalPayload)
		// if err != nil {
		// 	return err
		// }
		// zclog.Infof("chaincodeProposalPayload: %s", chaincodeProposalPayload.String())

		// break
	}

	return nil
}

// TODO
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
