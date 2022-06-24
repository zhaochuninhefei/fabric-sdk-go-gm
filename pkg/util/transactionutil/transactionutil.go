package transactionutil

import (
	"fmt"
	"time"

	"gitee.com/zhaochuninhefei/fabric-protos-go-gm/peer"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/internal/gitee.com/zhaochuninhefei/fabric-gm/protoutil"
	"gitee.com/zhaochuninhefei/gmgo/x509"
)

// 交易信息
type TransactionInfo struct {
	TxID         string   // 交易ID
	TxType       int32    // 交易类型
	TxCreateTime string   // 交易创建时间
	TxCcID       string   // 交易调用链码ID
	TxCcVersion  string   // 交易调用链码版本
	TxArgs       []string // 输入参数
	CallerMspID  string   // 交易发起者MSPID
	CallerName   string   // 交易发起者名称
	CallerOU     string   // 交易发起者OU分组
	Nonce        []byte   // 随机数
}

func (t *TransactionInfo) ToString() string {
	str := fmt.Sprintf("TxID: %s, TxType: %d, TxCreateTime: %s, TxCcID: %s, TxCcVersion: %s, TxArgs: %q, CallerMspID: %s, CallerName: %s, CallerOU: %s, Nonce: %x",
		t.TxID, t.TxType, t.TxCreateTime, t.TxCcID, t.TxCcVersion, t.TxArgs, t.CallerMspID, t.CallerName, t.CallerOU, t.Nonce)
	return str
}

// UnmarshalTransaction 反序列化交易的payload字节数组
//  payloadRaw : 某条交易的Payload字节数组，一般由`network.GetLedgerClient().QueryTransaction()`获取。
func UnmarshalTransaction(payloadRaw []byte) (*TransactionInfo, error) {
	result := &TransactionInfo{}
	// 反序列化为`common.Payload`
	payload, err := protoutil.UnmarshalPayload(payloadRaw)
	if err != nil {
		return nil, err
	}
	// 将 payload.Header.ChannelHeader 字节数组反序列化为`common.ChannelHeader`（包含通道ID、交易ID及交易创建时间等)
	channelHeader, err := protoutil.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		return nil, err
	}
	// 将 payload.Header.SignatureHeader 字节数组反序列化为`common.SignatureHeader`（包含创建者和nonce)
	signHeader, err := protoutil.UnmarshalSignatureHeader(payload.Header.SignatureHeader)
	if err != nil {
		return nil, err
	}
	// 将 signHeader.GetCreator() 反序列化为`msp.SerializedIdentity`（包含证书和MSPID）
	identity, err := protoutil.UnmarshalSerializedIdentity(signHeader.GetCreator())
	if err != nil {
		return nil, err
	}
	// 解析x509证书
	cert, err := x509.ReadCertificateFromPem(identity.GetIdBytes())
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %s", err)
	}
	// 解析用户名和OU分组
	uname := cert.Subject.CommonName
	outypes := cert.Subject.OrganizationalUnit

	// 将 payload.Data 字节数组反序列化为`peer.Transaction`
	tx, err := protoutil.UnmarshalTransaction(payload.Data)
	if err != nil {
		return nil, err
	}
	// 将 tx.Actions[0].Payload 字节数组反序列化为`peer.ChaincodeActionPayload`
	chaincodeActionPayload, err := protoutil.UnmarshalChaincodeActionPayload(tx.Actions[0].Payload)
	if err != nil {
		return nil, err
	}
	// 将 chaincodeActionPayload.ChaincodeProposalPayload 字节数组反序列化为`peer.ChaincodeProposalPayload`
	chaincodeProposalPayload, err := protoutil.UnmarshalChaincodeProposalPayload(chaincodeActionPayload.ChaincodeProposalPayload)
	if err != nil {
		return nil, err
	}
	// 将 chaincodeProposalPayload.Input 字节数组反序列化为`peer.ChaincodeInvocationSpec`
	chaincodeInvocationSpec, err := protoutil.UnmarshalChaincodeInvocationSpec(chaincodeProposalPayload.Input)
	if err != nil {
		return nil, err
	}
	// 获取链码ID，版本等信息
	result.TxCcID = chaincodeInvocationSpec.ChaincodeSpec.ChaincodeId.Name
	result.TxCcVersion = chaincodeInvocationSpec.ChaincodeSpec.ChaincodeId.Version
	// 获取本次交易的链码调用输入参数
	var args []string
	for _, v := range chaincodeInvocationSpec.ChaincodeSpec.Input.Args {
		args = append(args, string(v))
	}
	result.TxArgs = args
	result.Nonce = signHeader.GetNonce()
	result.TxType = channelHeader.GetType()
	result.TxID = channelHeader.GetTxId()
	result.CallerMspID = identity.GetMspid()
	result.CallerName = uname
	result.CallerOU = outypes[0]
	result.TxCreateTime = time.Unix(channelHeader.Timestamp.Seconds, 0).Format("2006-01-02 15:04:05")

	// TODO chaincodeProposalPayload
	fmt.Printf("chaincodeProposalPayload: %s\n", chaincodeProposalPayload.String())

	return result, nil
}

func UnmarshalPeerTransaction(tx *peer.Transaction) error {
	//进一步从transaction中解析成ChaincodeActionPayload
	chaincodeActionPayload, err := protoutil.UnmarshalChaincodeActionPayload(tx.Actions[0].Payload)
	if err != nil {
		return err
	}
	//进一步解析成proposalPayload
	proposalPayload, err := protoutil.UnmarshalChaincodeProposalPayload(chaincodeActionPayload.ChaincodeProposalPayload)
	if err != nil {
		return err
	}
	//得到交易调用的链码信息
	chaincodeInvocationSpec, err := protoutil.UnmarshalChaincodeInvocationSpec(proposalPayload.Input)
	if err != nil {
		return err
	}
	fmt.Printf("----- %s\n", chaincodeInvocationSpec.ChaincodeSpec.ChaincodeId.Name)
	return nil
}
