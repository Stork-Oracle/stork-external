package stork_publisher_agent

import "math/big"

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
	OracleId          string
	AssetId           string
	EvmPublisherKey   string
	EvmPrivateKey     string
	StarkPublisherKey string
	StarkPrivateKey   string
	PublisherKey      string
	SignatureType     string
	QuantizedPrice    string
)

const EvmSignatureType = SignatureType("evm")
const StarkSignatureType = SignatureType("stark")

// Incoming
type PriceUpdatePullWebsocket struct {
	PublishTimestamp int64   `json:"t"`
	Asset            AssetId `json:"a"`
	Price            float64 `json:"p"`
}

type ValueUpdatePushWebsocket struct {
	PublishTimestamp int64       `json:"t"`
	Asset            AssetId     `json:"a"`
	Value            interface{} `json:"v"`
}

// Intermediate
type TriggerType string

const ClockTriggerType = TriggerType("clock")
const DeltaTriggerType = TriggerType("delta")
const UnspecifiedTriggerType = TriggerType("unspecified")

type ValueUpdate struct {
	PublishTimestamp int64
	Asset            AssetId
	Value            *big.Float
}

type ValueUpdateWithTrigger struct {
	ValueUpdate ValueUpdate
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

type BrokerPublishUrl string
type BrokerConnectionConfig struct {
	PublishUrl BrokerPublishUrl `json:"publish_url"`
	AssetIds   []AssetId        `json:"asset_ids"`
}
