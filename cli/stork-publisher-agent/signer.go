package main

import "time"

type Signer[T Signature] struct {
	config StorkPublisherAgentConfig
}

func (s *Signer[T]) Sign(priceUpdate PriceUpdate, triggerType TriggerType) SignedPriceUpdate[T] {
	// todo: sign this for real
	timestampedSignature := TimestampedSignature[T]{
		Signature: nil,
		Timestamp: time.Now().UnixNano(),
		MsgHash:   "fake_hash",
	}
	quantizedPrice := FloatToQuantizedPrice(priceUpdate.Price)
	return SignedPriceUpdate[T]{
		OracleId: s.config.oracleId,
		AssetId:  priceUpdate.Asset,
		Trigger:  triggerType,
		SignedPrice: SignedPrice[T]{
			PublisherKey:         s.config.publisherKey,
			ExternalAssetId:      "TODO",
			SignatureType:        s.config.signatureType,
			QuantizedPrice:       quantizedPrice,
			TimestampedSignature: timestampedSignature,
		},
	}
}
