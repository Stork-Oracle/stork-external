package stork_publisher_agent

import (
	"github.com/rs/zerolog"
	"math"
	"time"
)

type DeltaTick struct{}

type ClockTick struct{}

const SigningSpeedBatchSize = 100000

type PriceUpdateProcessor[T Signature] struct {
	priceUpdateCh             chan PriceUpdate
	signedPriceUpdateBatchCh  chan SignedPriceUpdateBatch[T]
	signer                    Signer[T]
	priceUpdates              map[AssetId]PriceUpdate
	lastReportedPrice         map[AssetId]float64
	clockPeriod               time.Duration
	deltaCheckPeriod          time.Duration
	changeThresholdProportion float64 // 0-1
	logger                    zerolog.Logger
	totalSignatures           int
	totalSigningNs            int64
}

func NewPriceUpdateProcessor[T Signature](
	signer Signer[T],
	clockPeriod time.Duration,
	deltaCheckPeriod time.Duration,
	changeThresholdProportion float64,
	priceUpdateCh chan PriceUpdate,
	signedPriceUpdateBatchCh chan SignedPriceUpdateBatch[T],
	logger zerolog.Logger,
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
		logger:                    logger,
	}
}

func (p *PriceUpdateProcessor[T]) DeltaUpdate() PriceUpdatesWithTrigger {
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
	return PriceUpdatesWithTrigger{updates: significantUpdates, TriggerType: DeltaTriggerType}
}

func (p *PriceUpdateProcessor[T]) ClockUpdate() PriceUpdatesWithTrigger {
	updates := make(map[AssetId]PriceUpdate)
	// make a copy of the
	for asset, priceUpdate := range p.priceUpdates {
		updates[asset] = priceUpdate
	}

	return PriceUpdatesWithTrigger{updates: updates, TriggerType: ClockTriggerType}
}

func (p *PriceUpdateProcessor[T]) SignBatch(updates PriceUpdatesWithTrigger) SignedPriceUpdateBatch[T] {
	signedPriceUpdateBatch := make(SignedPriceUpdateBatch[T])
	startTime := time.Now()
	for asset, priceUpdate := range updates.updates {
		signedPriceUpdateBatch[asset] = p.signer.GetSignedPriceUpdate(priceUpdate, updates.TriggerType)
	}
	elapsedNs := time.Since(startTime).Nanoseconds()
	p.totalSigningNs += elapsedNs
	p.totalSignatures += len(updates.updates)

	if p.totalSignatures > SigningSpeedBatchSize {
		nsPerSignature := float64(p.totalSigningNs) / float64(p.totalSignatures)
		p.logger.Info().Msgf("Average signing speed for last %v signatures: %f ns/signature", p.totalSignatures, nsPerSignature)
		p.totalSigningNs = 0
		p.totalSignatures = 0
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
		for range time.Tick(p.clockPeriod) {
			q <- ClockTick{}
		}
	}(queue)

	// delta check thread
	go func(q chan any) {
		for range time.Tick(p.deltaCheckPeriod) {
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
		var updates PriceUpdatesWithTrigger
		switch msg := val.(type) {
		case DeltaTick:
			updates = p.DeltaUpdate()
		case ClockTick:
			updates = p.ClockUpdate()
		case PriceUpdate:
			p.priceUpdates[msg.Asset] = msg
		}

		if len(updates.updates) > 0 {
			priceUpdatesToSignCh <- updates

			for asset, priceUpdate := range updates.updates {
				p.lastReportedPrice[asset] = priceUpdate.Price
			}
		}
	}
}
