package publisher_agent

import (
	"math"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/signer"
	"github.com/rs/zerolog"
)

type DeltaTick struct{}

type ClockTick struct{}

const SignedMessageBatchPeriod = 1 * time.Millisecond

type ValueUpdateProcessor[T signer.Signature] struct {
	valueUpdateCh             chan ValueUpdate
	signedPriceUpdateBatchCh  chan SignedPriceUpdateBatch[T]
	signer                    signer.Signer[T]
	oracleId                  OracleId
	numRunners                int
	valueUpdates              map[AssetId]ValueUpdate
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

func NewPriceUpdateProcessor[T signer.Signature](
	signer signer.Signer[T],
	oracleId OracleId,
	numRunners int,
	clockPeriod time.Duration,
	deltaCheckPeriod time.Duration,
	changeThresholdProportion float64,
	signEveryUpdate bool,
	valueUpdateCh chan ValueUpdate,
	signedPriceUpdateBatchCh chan SignedPriceUpdateBatch[T],
	logger zerolog.Logger,
) *ValueUpdateProcessor[T] {
	return &ValueUpdateProcessor[T]{
		valueUpdateCh:             valueUpdateCh,
		signedPriceUpdateBatchCh:  signedPriceUpdateBatchCh,
		signer:                    signer,
		oracleId:                  oracleId,
		numRunners:                numRunners,
		valueUpdates:              make(map[AssetId]ValueUpdate),
		lastReportedPrice:         make(map[AssetId]float64),
		clockPeriod:               clockPeriod,
		deltaCheckPeriod:          deltaCheckPeriod,
		changeThresholdProportion: changeThresholdProportion,
		signEveryUpdate:           signEveryUpdate,
		logger:                    logger,
	}
}

func (vup *ValueUpdateProcessor[T]) DeltaUpdate() []ValueUpdateWithTrigger {
	significantUpdates := make([]ValueUpdateWithTrigger, 0)
	for asset, valueUpdate := range vup.valueUpdates {
		// float imprecision is ok for change threshold computation
		currentValue, _ := valueUpdate.Value.Float64()
		lastReportedValue, exists := vup.lastReportedPrice[asset]
		if exists {
			if math.Abs((currentValue-lastReportedValue)/lastReportedValue) > vup.changeThresholdProportion {
				significantUpdates = append(
					significantUpdates,
					ValueUpdateWithTrigger{
						ValueUpdate: valueUpdate,
						TriggerType: DeltaTriggerType,
					},
				)
			}
		} else {
			// the first update for an asset is always significant
			significantUpdates = append(
				significantUpdates,
				ValueUpdateWithTrigger{
					ValueUpdate: valueUpdate,
					TriggerType: DeltaTriggerType,
				},
			)
		}
	}
	return significantUpdates
}

func (vup *ValueUpdateProcessor[T]) ClockUpdate() []ValueUpdateWithTrigger {
	updates := make([]ValueUpdateWithTrigger, 0)

	for _, valueUpdate := range vup.valueUpdates {
		currentTimeValueUpdate := ValueUpdate{
			PublishTimestamp: time.Now().UnixNano(),
			Value:            valueUpdate.Value,
			Asset:            valueUpdate.Asset,
			Metadata:         valueUpdate.Metadata,
		}
		updates = append(
			updates,
			ValueUpdateWithTrigger{
				ValueUpdate: currentTimeValueUpdate,
				TriggerType: ClockTriggerType,
			},
		)
	}

	return updates
}

func (vup *ValueUpdateProcessor[T]) Run() {
	queue := make(chan any, 4096)
	priceUpdatesToSignCh := make(chan ValueUpdateWithTrigger, 4096)
	signedPriceUpdateCh := make(chan SignedPriceUpdate[T], 4096)

	// update price map thread
	go func(q chan any) {
		for valueUpdate := range vup.valueUpdateCh {
			q <- valueUpdate
		}
	}(queue)

	// clock thread if configured
	if vup.clockPeriod.Nanoseconds() > 0 {
		go func(q chan any) {
			for range time.Tick(vup.clockPeriod) {
				q <- ClockTick{}
			}
		}(queue)
	}

	if !vup.signEveryUpdate {
		// delta check thread
		go func(q chan any) {
			for range time.Tick(vup.deltaCheckPeriod) {
				q <- DeltaTick{}
			}
		}(queue)
	}

	numSignerThreads := max(runtime.NumCPU()/vup.numRunners, 1)
	vup.logger.Debug().Msgf("Starting %v signer threads", numSignerThreads)
	// start a signing thread for each CPU core
	for i := 0; i < numSignerThreads; i++ {
		go func(updates chan ValueUpdateWithTrigger, signedUpdates chan SignedPriceUpdate[T], threadNum int) {
			for update := range updates {
				start := time.Now()

				quantizedPrice := FloatToQuantizedPrice(update.ValueUpdate.Value)
				timestampedSig, externalAssetId, err := vup.signer.SignPublisherPrice(update.ValueUpdate.PublishTimestamp, string(update.ValueUpdate.Asset), string(quantizedPrice))
				if err != nil {
					vup.logger.Error().Err(err).Msg("Failed to sign update")
					continue
				}

				priceUpdate := SignedPriceUpdate[T]{
					OracleId: vup.oracleId,
					AssetId:  update.ValueUpdate.Asset,
					Trigger:  update.TriggerType,
					SignedPrice: SignedPrice[T]{
						PublisherKey:         signer.PublisherKey(vup.signer.GetPublisherKey()),
						ExternalAssetId:      externalAssetId,
						SignatureType:        signer.SignatureType(vup.signer.GetSignatureType()),
						QuantizedPrice:       quantizedPrice,
						TimestampedSignature: *timestampedSig,
						Metadata:             update.ValueUpdate.Metadata,
					},
				}

				signedUpdates <- priceUpdate
				elapsed := time.Since(start).Microseconds()
				atomic.AddInt32(&vup.signQueueSize, -1)
				ageMs := (time.Now().UnixNano() - update.ValueUpdate.PublishTimestamp) / 1_000_000
				vup.logger.Debug().Msgf("Signing update on thread %v took %v microseconds (age %v ms, queue size: %v)", threadNum, elapsed, ageMs, vup.signQueueSize)
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
						vup.signedPriceUpdateBatchCh <- signedPriceUpdateBatch
						signedPriceUpdateBatch = make(SignedPriceUpdateBatch[T])
					}
				}
			}
		}
	}(signedPriceUpdateCh)

	for val := range queue {
		var valueUpdates []ValueUpdateWithTrigger
		switch msg := val.(type) {
		case DeltaTick:
			valueUpdates = vup.DeltaUpdate()
		case ClockTick:
			valueUpdates = vup.ClockUpdate()
		case ValueUpdate:
			if vup.signEveryUpdate {
				priceUpdatesToSignCh <- ValueUpdateWithTrigger{ValueUpdate: msg, TriggerType: UnspecifiedTriggerType}
				atomic.AddInt32(&vup.signQueueSize, 1)
			}
			vup.valueUpdates[msg.Asset] = msg
		}

		if len(valueUpdates) > 0 {
			for _, priceUpdate := range valueUpdates {
				priceUpdatesToSignCh <- priceUpdate
				atomic.AddInt32(&vup.signQueueSize, 1)
				lastReportedPrice, _ := priceUpdate.ValueUpdate.Value.Float64()
				vup.lastReportedPrice[priceUpdate.ValueUpdate.Asset] = lastReportedPrice
			}
		}
	}
}
