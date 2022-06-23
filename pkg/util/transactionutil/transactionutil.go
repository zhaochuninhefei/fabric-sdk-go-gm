package transactionutil

import (
	"encoding/pem"
	"fmt"
	"time"

	"gitee.com/zhaochuninhefei/fabric-protos-go-gm/peer"
	"gitee.com/zhaochuninhefei/fabric-sdk-go-gm/internal/gitee.com/zhaochuninhefei/fabric-gm/protoutil"
	"gitee.com/zhaochuninhefei/gmgo/x509"
)

type TransactionInfo struct {
	CreateTime       string   //交易创建时间
	ChaincodeID      string   //交易调用链码ID
	ChaincodeVersion string   //交易调用链码版本
	Nonce            []byte   //随机数
	Mspid            string   //交易发起者MSPID
	Name             string   //交易发起者名称
	OUTypes          string   //交易发起者OU分组
	Args             []string //输入参数
	Type             int32    //交易类型
	TxID             string   //交易ID
}

// UnmarshalTransaction 从某个交易的payload来解析它
func UnmarshalTransaction(payloadRaw []byte) (*TransactionInfo, error) {
	result := &TransactionInfo{}
	// 解析成payload
	payload, err := protoutil.UnmarshalPayload(payloadRaw)
	if err != nil {
		return nil, err
	}
	// 解析成ChannelHeader（包含通道ID、交易ID及交易创建时间等)
	channelHeader, err := protoutil.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		return nil, err
	}
	// 解析成SignatureHeader（包含创建者和nonce)
	signHeader, err := protoutil.UnmarshalSignatureHeader(payload.Header.SignatureHeader)
	if err != nil {
		return nil, err
	}
	// 解析成SerializedIdentity（包含证书和MSPID）
	identity, err := protoutil.UnmarshalSerializedIdentity(signHeader.GetCreator())
	if err != nil {
		return nil, err
	}
	// 下面为解析证书
	block, _ := pem.Decode(identity.GetIdBytes())
	if block == nil {
		return nil, fmt.Errorf("identity could not be decoded from credential")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %s", err)
	}
	// 解析用户名和OU分组
	uname := cert.Subject.CommonName
	outypes := cert.Subject.OrganizationalUnit

	// 解析成transaction
	tx, err := protoutil.UnmarshalTransaction(payload.Data)
	if err != nil {
		return nil, err
	}
	// 进一步从transaction中解析成ChaincodeActionPayload
	chaincodeActionPayload, err := protoutil.UnmarshalChaincodeActionPayload(tx.Actions[0].Payload)
	if err != nil {
		return nil, err
	}
	// 进一步解析成proposalPayload
	proposalPayload, err := protoutil.UnmarshalChaincodeProposalPayload(chaincodeActionPayload.ChaincodeProposalPayload)
	if err != nil {
		return nil, err
	}
	// 得到交易调用的链码信息
	chaincodeInvocationSpec, err := protoutil.UnmarshalChaincodeInvocationSpec(proposalPayload.Input)
	if err != nil {
		return nil, err
	}
	// 得到调用的链码的ID，版本和PATH（这里PATH省略了）
	result.ChaincodeID = chaincodeInvocationSpec.ChaincodeSpec.ChaincodeId.Name
	result.ChaincodeVersion = chaincodeInvocationSpec.ChaincodeSpec.ChaincodeId.Version
	// 得到输入参数
	var args []string
	for _, v := range chaincodeInvocationSpec.ChaincodeSpec.Input.Args {
		args = append(args, string(v))
	}
	result.Args = args
	result.Nonce = signHeader.GetNonce()
	result.Type = channelHeader.GetType()
	result.TxID = channelHeader.GetTxId()
	result.Mspid = identity.GetMspid()
	result.Name = uname
	result.OUTypes = outypes[0]
	result.CreateTime = time.Unix(channelHeader.Timestamp.Seconds, 0).Format("2006-01-02 15:04:05")
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
