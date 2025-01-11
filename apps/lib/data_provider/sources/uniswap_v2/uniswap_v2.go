package uniswap_v2

import (
	"fmt"
	"math"
	"math/big"
	"os"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
)

const uniswapV2AbiFileName = "uniswap_v2.json"
const getUniswapV2ContractFunction = "getReserves"

type uniswapV2DataSource struct {
	logger          zerolog.Logger
	uniswapConfig   uniswapV2Config
	apiKey          string
	updateFrequency time.Duration
	valueId         types.ValueId
	contract        *bind.BoundContract
}

func newUniswapV2DataSource(sourceConfig types.DataProviderSourceConfig) *uniswapV2DataSource {
	var uniswapConfig uniswapV2Config
	mapstructure.Decode(sourceConfig.Config, &uniswapConfig)

	updateFrequency, err := time.ParseDuration(uniswapConfig.UpdateFrequency)
	if err != nil {
		panic("unable to parse update frequency: " + uniswapConfig.UpdateFrequency)
	}

	apiKey := ""
	if len(uniswapConfig.ProviderApiKeyEnvVar) > 0 {
		var exists bool
		apiKey, exists = os.LookupEnv(uniswapConfig.ProviderApiKeyEnvVar)
		if !exists {
			panic("env var with name " + uniswapConfig.ProviderApiKeyEnvVar + " is not set")
		}
	}

	return &uniswapV2DataSource{
		uniswapConfig:   uniswapConfig,
		valueId:         sourceConfig.Id,
		apiKey:          apiKey,
		updateFrequency: updateFrequency,
		logger:          utils.DataSourceLogger(UniswapV2DataSourceId),
	}
}

func (c *uniswapV2DataSource) RunDataSource(updatesCh chan types.DataSourceUpdateMap) {
	updater := func() (types.DataSourceUpdateMap, error) { return c.getUpdate() }
	scheduler := sources.NewScheduler(
		c.updateFrequency,
		updater,
		sources.GetErrorLogHandler(c.logger, zerolog.WarnLevel),
	)
	scheduler.RunScheduler(updatesCh)
}

func (c *uniswapV2DataSource) getUpdate() (types.DataSourceUpdateMap, error) {
	if c.contract == nil {
		err := c.initializeBoundContract()
		if err != nil {
			return nil, fmt.Errorf("failed to bind to contract: %v", err)
		}
	}
	updateValue, err := c.getPrice()
	if err != nil {
		return nil, fmt.Errorf("failed to get price: %v", err)
	}

	updates := make(types.DataSourceUpdateMap)

	updateTime := time.Now().UTC().UnixMilli()
	updates[c.valueId] = types.DataSourceValueUpdate{
		Timestamp:    time.UnixMilli(updateTime),
		ValueId:      c.valueId,
		Value:        updateValue,
		DataSourceId: UniswapV2DataSourceId,
	}

	return updates, nil
}

func (c *uniswapV2DataSource) initializeBoundContract() error {
	contract, err := sources.GetEthereumContract(
		c.uniswapConfig.ContractAddress,
		uniswapV2AbiFileName,
		c.uniswapConfig.HttpProviderUrl,
		c.apiKey,
		resourcesFS,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize contract: %v", err)
	}
	c.contract = contract
	return nil
}

// hit the contract and compute the price
func (c *uniswapV2DataSource) getPrice() (float64, error) {
	result, err := sources.CallEthereumFunction(
		c.contract,
		getUniswapV2ContractFunction,
		c.valueId,
		c.logger,
	)
	if err != nil {
		return -1, fmt.Errorf("failed to call ethereum contract: %v", err)
	}

	return calculatePrice(
		result,
		c.uniswapConfig.BaseTokenIndex,
		c.uniswapConfig.QuoteTokenIndex,
		c.uniswapConfig.BaseTokenDecimals,
		c.uniswapConfig.QuoteTokenDecimals,
	)
}

// helper function to convert the result object to a useful price
func calculatePrice(result []interface{}, baseTokenIndex int8, quoteTokenIndex int8, baseTokenDecimals int8, quoteTokenDecimals int8) (float64, error) {
	reserveBase, ok := result[baseTokenIndex].(*big.Int)
	if !ok {
		return -1, fmt.Errorf("failed to convert reserveBase size to big int: %v", ok)
	}
	reserveQuote, ok := result[quoteTokenIndex].(*big.Int)
	if !ok {
		return -1, fmt.Errorf("failed to convert reserveQuote size to big int: %v", ok)
	}

	if reserveBase.Cmp(big.NewInt(0)) == 0 || reserveQuote.Cmp(big.NewInt(0)) == 0 {
		return -1, fmt.Errorf("pool had zero reserve tokens for some coin")
	}

	reserveBaseFloat := new(big.Float).SetInt(reserveBase)
	reserveQuoteFloat := new(big.Float).SetInt(reserveQuote)

	tokenA := new(big.Float).Quo(reserveQuoteFloat, reserveBaseFloat)
	price, _ := tokenA.Float64()

	exponent := float64(baseTokenDecimals - quoteTokenDecimals)
	price = price * math.Pow(10, exponent)

	return price, nil
}
