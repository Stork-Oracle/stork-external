package main

import (
	"math"
	"time"
)

type DeltaTick struct{}

type ClockTick struct{}

type PriceUpdateProcessor[T Signature] struct {
	priceUpdateCh             chan PriceUpdate
	signedPriceUpdateBatchCh  chan SignedPriceUpdateBatch[T]
	signer                    Signer[T]
	priceUpdates              map[AssetId]PriceUpdate
	lastReportedPrice         map[AssetId]float64
	clockPeriod               time.Duration
	deltaCheckPeriod          time.Duration
	changeThresholdProportion float64 // 0-1
}

func NewPriceUpdateProcessor[T Signature](
	signer Signer[T],
	clockPeriod time.Duration,
	deltaCheckPeriod time.Duration,
	changeThresholdProportion float64,
	priceUpdateCh chan PriceUpdate,
	signedPriceUpdateBatchCh chan SignedPriceUpdateBatch[T],
) *PriceUpdateProcessor[T] {
	return &PriceUpdateProcessor[T]{
		priceUpdateCh:             priceUpdateCh,
		signedPriceUpdateBatchCh:  signedPriceUpdateBatchCh,
		signer:                    signer,
		priceUpdates:              make(map[AssetId]PriceUpdate),
		lastReportedPrice:         make(map[AssetId]float64),
		clockPeriod:               clockPeriod,
		deltaCheckPeriod:          deltaCheckPeriod,
		changeThresholdProportion: changeThresholdProportion,
	}
}

func (p *PriceUpdateProcessor[T]) DeltaUpdate() *PriceUpdatesWithTrigger {
	significantUpdates := make(map[AssetId]PriceUpdate)
	for asset, priceUpdate := range p.priceUpdates {
		currentPrice := priceUpdate.Price
		lastReportedPrice, exists := p.lastReportedPrice[asset]
		if exists {
			if math.Abs((currentPrice-lastReportedPrice)/lastReportedPrice) > p.changeThresholdProportion {
				significantUpdates[asset] = priceUpdate
			}
		}
	}
	return &PriceUpdatesWithTrigger{updates: significantUpdates, TriggerType: DeltaTriggerType}
}

func (p *PriceUpdateProcessor[T]) ClockUpdate() *PriceUpdatesWithTrigger {
	return &PriceUpdatesWithTrigger{updates: p.priceUpdates, TriggerType: ClockTriggerType}
}

func (p *PriceUpdateProcessor[T]) SignBatch(updates PriceUpdatesWithTrigger) SignedPriceUpdateBatch[T] {
	signedPriceUpdateBatch := make(SignedPriceUpdateBatch[T])
	for asset, priceUpdate := range updates.updates {
		signedPriceUpdateBatch[asset] = p.signer.Sign(priceUpdate, updates.TriggerType)
	}
	return signedPriceUpdateBatch
}

func (p *PriceUpdateProcessor[T]) Run() {
	queue := make(chan any, 4096)
	priceUpdatesToSignCh := make(chan PriceUpdatesWithTrigger, 4096)

	// update price map thread
	go func(q chan any) {
		for priceUpdate := range p.priceUpdateCh {
			q <- priceUpdate
		}
	}(queue)

	// clock thread
	go func(q chan any) {
		for _ = range time.Tick(p.clockPeriod) {
			q <- ClockTick{}
		}
	}(queue)

	// delta check thread
	go func(q chan any) {
		for _ = range time.Tick(p.deltaCheckPeriod) {
			q <- DeltaTick{}
		}
	}(queue)

	// signing thread
	go func(q chan PriceUpdatesWithTrigger) {
		for update := range q {
			p.signedPriceUpdateBatchCh <- p.SignBatch(update)
		}
	}(priceUpdatesToSignCh)

	for val := range queue {
		var updates *PriceUpdatesWithTrigger
		switch msg := val.(type) {
		case DeltaTick:
			updates = p.DeltaUpdate()
		case ClockTick:
			updates = p.ClockUpdate()
		case PriceUpdate:
			p.priceUpdates[msg.Asset] = msg
		}

		if updates != nil && len(updates.updates) > 0 {
			priceUpdatesToSignCh <- *updates

			for asset, priceUpdate := range updates.updates {
				p.lastReportedPrice[asset] = priceUpdate.Price
			}
		}
	}
}
