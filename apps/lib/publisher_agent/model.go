package publisher_agent

import (
	"math/big"

	"github.com/Stork-Oracle/stork-external/apps/lib/signer"
)

type MessageType string

const WildcardSubscriptionAsset = "*"

type (
	ConnectionId string
	AuthToken    string
)

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
	QuantizedPrice string
)

const (
	EvmSignatureType   = signer.SignatureType("evm")
	StarkSignatureType = signer.SignatureType("stark")
)

// Incoming
type (
	Metadata                 map[string]any
	PriceUpdatePullWebsocket struct {
		PublishTimestampNano int64    `json:"t"`
		Asset                AssetId  `json:"a"`
		Price                float64  `json:"p"`
		Metadata             Metadata `json:"m,omitempty"`
	}
)

type ValueUpdatePushWebsocket struct {
	PublishTimestampNano int64    `json:"t"`
	Asset                AssetId  `json:"a"`
	Value                any      `json:"v"`
	Metadata             Metadata `json:"m,omitempty"`
}

// Intermediate
type TriggerType string

const (
	ClockTriggerType       = TriggerType("clock")
	DeltaTriggerType       = TriggerType("delta")
	UnspecifiedTriggerType = TriggerType("unspecified")
)

type ValueUpdate struct {
	PublishTimestampNano int64
	Asset                AssetId
	Value                *big.Float
	Metadata             Metadata
}

type ValueUpdateWithTrigger struct {
	ValueUpdate ValueUpdate
	TriggerType TriggerType
}

// Outgoing
type SignedPrice[T signer.Signature] struct {
	PublisherKey         signer.PublisherKey            `json:"publisher_key"`
	ExternalAssetId      string                         `json:"external_asset_id"`
	SignatureType        signer.SignatureType           `json:"signature_type"`
	QuantizedPrice       QuantizedPrice                 `json:"price"`
	TimestampedSignature signer.TimestampedSignature[T] `json:"timestamped_signature"`
	Metadata             Metadata                       `json:"metadata,omitempty"`
}

// SignedPriceUpdate represents a signed price from a publisher
type SignedPriceUpdate[T signer.Signature] struct {
	OracleId    OracleId       `json:"oracle_id"`
	AssetId     AssetId        `json:"asset_id"`
	Trigger     TriggerType    `json:"trigger"`
	SignedPrice SignedPrice[T] `json:"signed_price"`
}

type SignedPriceUpdateBatch[T signer.Signature] map[AssetId]SignedPriceUpdate[T]

type SubscriptionRequest struct {
	Assets []AssetId `json:"assets"`
}

type (
	BrokerPublishUrl       string
	BrokerConnectionConfig struct {
		PublishUrl BrokerPublishUrl `json:"publish_url"`
		AssetIds   []AssetId        `json:"asset_ids"`
	}
)

type RegistryErrorResponse struct {
	Error string `json:"error"`
}
