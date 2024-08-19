package stork_publisher_agent

type MessageType string

const WildcardSubscriptionAsset = "*"

type ConnectionId string
type AuthToken string

type WebsocketMessage[T any] struct {
	connId  ConnectionId
	Type    string `json:"type"`
	Error   string `json:"error,omitempty"`
	TraceId string `json:"trace_id,omitempty"`
	Data    T      `json:"data,omitempty"`
}

type ErrorMessage struct {
	Error string `json:"error"`
}

type (
	OracleId       string
	AssetId        string
	PublisherKey   string
	PrivateKey     string
	SignatureType  string
	QuantizedPrice string
)

const EvmSignatureType = SignatureType("evm")
const StarkSignatureType = SignatureType("stark")

// Incoming

type PriceUpdate struct {
	PublishTimestamp int64   `json:"t"`
	Asset            AssetId `json:"a"`
	Price            float64 `json:"p"` // todo: consider making this a string
}

// Intermediate
type TriggerType string

const ClockTriggerType = TriggerType("clock")
const DeltaTriggerType = TriggerType("delta")

type PriceUpdatesWithTrigger struct {
	updates     map[AssetId]PriceUpdate
	TriggerType TriggerType
}

// Outgoing

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

type SignedPrice[T Signature] struct {
	PublisherKey         PublisherKey            `json:"publisher_key"`
	ExternalAssetId      string                  `json:"external_asset_id"`
	SignatureType        SignatureType           `json:"signature_type"`
	QuantizedPrice       QuantizedPrice          `json:"price"`
	TimestampedSignature TimestampedSignature[T] `json:"timestamped_signature"`
}

// SignedPriceUpdate represents a signed price from a publisher
type SignedPriceUpdate[T Signature] struct {
	OracleId    OracleId       `json:"oracle_id"`
	AssetId     AssetId        `json:"asset_id"`
	Trigger     TriggerType    `json:"trigger"`
	SignedPrice SignedPrice[T] `json:"signed_price"`
}

type SignedPriceUpdateBatch[T Signature] map[AssetId]SignedPriceUpdate[T]

type SubscriptionRequest struct {
	Assets []AssetId `json:"assets"`
}
