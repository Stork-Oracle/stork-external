package stork_publisher_agent

import (
	"fmt"
	"math/big"
	"sort"
	"sync"
)

func FloatToQuantizedPrice(f float64) QuantizedPrice {
	// convert to string first to avoid rounding error
	strValue := fmt.Sprintf("%.18f", f)
	bigFloatValue, _, _ := big.ParseFloat(strValue, 10, 0, big.ToZero)
	multiplier := new(big.Float).SetInt64(1e18)
	bigFloatValue.Mul(bigFloatValue, multiplier)
	result := new(big.Int)
	bigFloatValue.Int(result)
	return StringifyQuantizedPrice(result)
}

func StringifyQuantizedPrice(price *big.Int) QuantizedPrice {
	// Convert the big.Int to a string
	valStr := price.String()

	if len(valStr) > 6 {
		// zero out last 6 digits
		valStr = valStr[:len(valStr)-6] + "000000"
	}

	return QuantizedPrice(valStr)
}

type SubscriptionTracker struct {
	assetsLock sync.RWMutex
	assets     map[AssetId]struct{}
	allAssets  bool
}

func NewSubscriptionTracker() *SubscriptionTracker {
	return &SubscriptionTracker{
		assets:     make(map[AssetId]struct{}),
		allAssets:  false,
		assetsLock: sync.RWMutex{},
	}
}

func (st *SubscriptionTracker) Subscribe(assets []AssetId) {
	st.assetsLock.Lock()
	defer st.assetsLock.Unlock()

	for _, asset := range assets {
		if asset == WildcardSubscriptionAsset {
			st.allAssets = true
			break
		}
		st.assets[asset] = struct{}{}
	}
}

func (st *SubscriptionTracker) Unsubscribe(assets []AssetId) {
	st.assetsLock.Lock()
	defer st.assetsLock.Unlock()
	for _, asset := range assets {
		if asset == WildcardSubscriptionAsset {
			st.allAssets = false
			clear(st.assets)
			break
		}
		delete(st.assets, asset)
	}
}

func (st *SubscriptionTracker) GetSortedAssets() []AssetId {
	st.assetsLock.RLock()
	defer st.assetsLock.RUnlock()
	assets := make([]AssetId, 0, len(st.assets))
	for asset, _ := range st.assets {
		assets = append(assets, asset)
	}
	sort.Slice(assets, func(i, j int) bool {
		return assets[i] < assets[j]
	})
	return assets
}

func (st *SubscriptionTracker) IsSubscribed(assetId AssetId) bool {
	if st.allAssets {
		return true
	}
	st.assetsLock.RLock()
	defer st.assetsLock.RUnlock()
	_, exists := st.assets[assetId]
	return exists
}
