package stork_publisher_agent

import (
	"math"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
)

type DeltaTick struct{}

type ClockTick struct{}

const SignedMessageBatchPeriod = 1 * time.Millisecond

type PriceUpdateProcessor[T Signature] struct {
	priceUpdateCh             chan PriceUpdate
	signedPriceUpdateBatchCh  chan SignedPriceUpdateBatch[T]
	signer                    Signer[T]
	numRunners                int
	priceUpdates              map[AssetId]PriceUpdate
	lastReportedPrice         map[AssetId]float64
	clockPeriod               time.Duration
	deltaCheckPeriod          time.Duration
	changeThresholdProportion float64 // 0-1
	signEveryUpdate           bool
	logger                    zerolog.Logger
	totalSignatures           int
	totalSigningNs            int64
	signQueueSize             int32
}

func NewPriceUpdateProcessor[T Signature](
	signer Signer[T],
	numRunners int,
	clockPeriod time.Duration,
	deltaCheckPeriod time.Duration,
	changeThresholdProportion float64,
	signEveryUpdate bool,
	priceUpdateCh chan PriceUpdate,
	signedPriceUpdateBatchCh chan SignedPriceUpdateBatch[T],
	logger zerolog.Logger,
) *PriceUpdateProcessor[T] {
	return &PriceUpdateProcessor[T]{
		priceUpdateCh:             priceUpdateCh,
		signedPriceUpdateBatchCh:  signedPriceUpdateBatchCh,
		signer:                    signer,
		numRunners:                numRunners,
		priceUpdates:              make(map[AssetId]PriceUpdate),
		lastReportedPrice:         make(map[AssetId]float64),
		clockPeriod:               clockPeriod,
		deltaCheckPeriod:          deltaCheckPeriod,
		changeThresholdProportion: changeThresholdProportion,
		signEveryUpdate:           signEveryUpdate,
		logger:                    logger,
	}
}

func (p *PriceUpdateProcessor[T]) DeltaUpdate() []PriceUpdateWithTrigger {
	significantUpdates := make([]PriceUpdateWithTrigger, 0)
	for asset, priceUpdate := range p.priceUpdates {
		// float imprecision is ok for change threshold computation
		currentPrice, _ := priceUpdate.Value.Float64()
		lastReportedPrice, exists := p.lastReportedPrice[asset]
		if exists {
			if math.Abs((currentPrice-lastReportedPrice)/lastReportedPrice) > p.changeThresholdProportion {
				significantUpdates = append(
					significantUpdates,
					PriceUpdateWithTrigger{
						PriceUpdate: priceUpdate,
						TriggerType: DeltaTriggerType,
					},
				)
			}
		}
	}
	return significantUpdates
}

func (p *PriceUpdateProcessor[T]) ClockUpdate() []PriceUpdateWithTrigger {
	updates := make([]PriceUpdateWithTrigger, 0)

	for _, priceUpdate := range p.priceUpdates {
		updates = append(
			updates,
			PriceUpdateWithTrigger{
				PriceUpdate: priceUpdate,
				TriggerType: ClockTriggerType,
			},
		)
	}

	return updates
}

func (p *PriceUpdateProcessor[T]) Run() {
	queue := make(chan any, 4096)
	priceUpdatesToSignCh := make(chan PriceUpdateWithTrigger, 4096)
	signedPriceUpdateCh := make(chan SignedPriceUpdate[T], 4096)

	// update price map thread
	go func(q chan any) {
		for priceUpdate := range p.priceUpdateCh {
			q <- priceUpdate
		}
	}(queue)

	if !p.signEveryUpdate {
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
	}

	numSignerThreads := runtime.NumCPU() / p.numRunners
	p.logger.Debug().Msgf("Starting %v signer threads", numSignerThreads)
	// start a signing thread for each CPU core
	for i := 0; i < numSignerThreads; i++ {
		go func(updates chan PriceUpdateWithTrigger, signedUpdates chan SignedPriceUpdate[T], threadNum int) {
			for update := range updates {
				start := time.Now()
				signedUpdates <- p.signer.GetSignedPriceUpdate(update.PriceUpdate, update.TriggerType)
				elapsed := time.Since(start).Microseconds()
				atomic.AddInt32(&p.signQueueSize, -1)
				ageMs := (time.Now().UnixNano() - update.PriceUpdate.PublishTimestamp) / 1_000_000
				p.logger.Debug().Msgf("Signing update on thread %v took %v microseconds (age %v ms, queue size: %v)", threadNum, elapsed, ageMs, p.signQueueSize)
			}
		}(priceUpdatesToSignCh, signedPriceUpdateCh, i)
	}

	// batch the signed updates together into outgoing messages
	go func(signedUpdates chan SignedPriceUpdate[T]) {
		ticker := time.NewTicker(SignedMessageBatchPeriod)
		signedPriceUpdateBatch := make(SignedPriceUpdateBatch[T])
		for {
			select {
			// add incoming signed updates into a map
			case signedUpdate := <-signedUpdates:
				signedPriceUpdateBatch[signedUpdate.AssetId] = signedUpdate

			case <-ticker.C:
				{
					if len(signedPriceUpdateBatch) > 0 {
						p.signedPriceUpdateBatchCh <- signedPriceUpdateBatch
						signedPriceUpdateBatch = make(SignedPriceUpdateBatch[T])
					}
				}
			}
		}
	}(signedPriceUpdateCh)

	for val := range queue {
		var priceUpdates []PriceUpdateWithTrigger
		switch msg := val.(type) {
		case DeltaTick:
			priceUpdates = p.DeltaUpdate()
		case ClockTick:
			priceUpdates = p.ClockUpdate()
		case PriceUpdate:
			if p.signEveryUpdate {
				priceUpdatesToSignCh <- PriceUpdateWithTrigger{PriceUpdate: msg, TriggerType: DeltaTriggerType}
				atomic.AddInt32(&p.signQueueSize, 1)
			} else {
				p.priceUpdates[msg.Asset] = msg
			}
		}

		if len(priceUpdates) > 0 {
			for _, priceUpdate := range priceUpdates {
				priceUpdatesToSignCh <- priceUpdate
				atomic.AddInt32(&p.signQueueSize, 1)
				lastReportedPrice, _ := priceUpdate.PriceUpdate.Value.Float64()
				p.lastReportedPrice[priceUpdate.PriceUpdate.Asset] = lastReportedPrice
			}
		}
	}
}
