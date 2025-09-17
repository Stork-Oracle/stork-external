package shared

type (
	AssetID        string
	EncodedAssetID string
	PublisherKey   string
	QuantizedPrice string

	AuthToken string

	SignatureType string
)

const (
	EvmSignatureType   = SignatureType("evm")
	StarkSignatureType = SignatureType("stark")
)

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
	Signature     T      `json:"signature"`
	TimestampNano uint64 `json:"timestamp"` //nolint:tagliatelle // this can't change for legacy reasons
	MsgHash       string `json:"msg_hash"`
}
