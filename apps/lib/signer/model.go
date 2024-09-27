package signer

type (
	SignatureType     string
	PublisherKey      string
	EvmPublisherKey   string
	EvmPrivateKey     string
	StarkPublisherKey string
	StarkPrivateKey   string
)

const EvmSignatureType = SignatureType("evm")
const StarkSignatureType = SignatureType("stark")

type Signature interface {
	*StarkSignature | *EvmSignature
}

type StarkSignature struct {
	R string `json:"r"`
	S string `json:"s"`
}

type EvmSignature struct {
	R string `json:"r"`
	S string `json:"s"`
	V string `json:"v"`
}

type TimestampedSignature[T Signature] struct {
	Signature T      `json:"signature"`
	Timestamp int64  `json:"timestamp"`
	MsgHash   string `json:"msg_hash"`
}
