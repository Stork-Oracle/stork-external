package publisher_agent

import (
	"math/big"

	"github.com/Stork-Oracle/stork-external/shared"
)

type MessageType string

const WildcardSubscriptionAsset = "*"

type (
	ConnectionID string
)

type WebsocketMessage[T any] struct {
	Type    string `json:"type"`
	Error   string `json:"error,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
	Data    T      `json:"data,omitempty"`
}

type ErrorMessage struct {
	Error string `json:"error"`
}

type (
	OracleID string
)

// Incoming
type (
	Metadata                 map[string]any
	PriceUpdatePullWebsocket struct {
		PublishTimestampNano int64          `json:"t"`
		Asset                shared.AssetID `json:"a"`
		Price                float64        `json:"p"`
		Metadata             Metadata       `json:"m,omitempty"`
	}
)

type ValueUpdatePushWebsocket struct {
	PublishTimestampNano int64          `json:"t"`
	Asset                shared.AssetID `json:"a"`
	Value                any            `json:"v"`
	Metadata             Metadata       `json:"m,omitempty"`
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
	Asset                shared.AssetID
	Value                *big.Float
	Metadata             Metadata
}

type ValueUpdateWithTrigger struct {
	ValueUpdate ValueUpdate
	TriggerType TriggerType
}

// Outgoing
type SignedPrice[T shared.Signature] struct {
	PublisherKey         shared.PublisherKey            `json:"publisher_key"`
	ExternalAssetID      string                         `json:"external_asset_id"`
	SignatureType        shared.SignatureType           `json:"signature_type"`
	QuantizedPrice       shared.QuantizedPrice          `json:"price"`
	TimestampedSignature shared.TimestampedSignature[T] `json:"timestamped_signature"`
	Metadata             Metadata                       `json:"metadata,omitempty"`
}

// SignedPriceUpdate represents a signed price from a publisher
type SignedPriceUpdate[T shared.Signature] struct {
	OracleID    OracleID       `json:"oracle_id"`
	AssetID     shared.AssetID `json:"asset_id"`
	Trigger     TriggerType    `json:"trigger"`
	SignedPrice SignedPrice[T] `json:"signed_price"`
}

type SignedPriceUpdateBatch[T shared.Signature] map[shared.AssetID]SignedPriceUpdate[T]

type SubscriptionRequest struct {
	Assets []shared.AssetID `json:"assets"`
}

type (
	BrokerPublishUrl       string
	BrokerConnectionConfig struct {
		PublishUrl BrokerPublishUrl `json:"publish_url"`
		AssetIDs   []shared.AssetID `json:"asset_ids"`
	}
)

type RegistryErrorResponse struct {
	Error string `json:"error"`
}
