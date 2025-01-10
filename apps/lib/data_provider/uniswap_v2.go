package data_provider

import (
	"embed"
	"fmt"
	"math"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
)

//go:embed resources
var resourcesFS embed.FS

// Define the DataSourceId which will be referenced in the config file and the data_source_registry
const UniswapV2DataSourceId = "UNISWAP_V2"

// Define the structure of the config needed to pull the appropriate data
type uniswapDataSourceConfig struct {
	UpdateFrequency      string `json:"updateFrequency"`
	HttpProviderUrl      string `json:"httpProviderUrl"`
	ProviderApiKeyEnvVar string `json:"providerApiKeyEnvVar"`
	ContractAddress      string `json:"contractAddress"`
	BaseTokenIndex       int8   `json:"baseTokenIndex"`
	QuoteTokenIndex      int8   `json:"quoteTokenIndex"`
	BaseTokenDecimals    int8   `json:"baseTokenDecimals,omitempty"`
	QuoteTokenDecimals   int8   `json:"quoteTokenDecimals,omitempty"`
}

const (
	maxQueryAttempts = 3
	baseRetryDelay   = 500 * time.Millisecond
)

const uniswapV2AbiFileName = "uniswap_v2.json"

const getUniswapV2ContractFunction = "getReserves"

type uniswapV2Connector struct {
	logger          zerolog.Logger
	uniswapConfig   uniswapDataSourceConfig
	apiKey          string
	updateFrequency time.Duration
	valueId         ValueId
	contract        *bind.BoundContract
}

func newUniswapV2AlchemyConnector(sourceConfig DataProviderSourceConfig) *uniswapV2Connector {
	var uniswapConfig uniswapDataSourceConfig
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

	return &uniswapV2Connector{
		logger:          dataSourceLogger(UniswapV2DataSourceId),
		uniswapConfig:   uniswapConfig,
		valueId:         sourceConfig.Id,
		apiKey:          apiKey,
		updateFrequency: updateFrequency,
	}
}

func (c *uniswapV2Connector) GetUpdate() (DataSourceUpdateMap, error) {
	if c.contract == nil {
		err := c.updateBoundContract()
		if err != nil {
			return nil, fmt.Errorf("failed to bind to contract: %v", err)
		}
	}
	updateValue, err := c.getPrice()
	if err != nil {
		return nil, fmt.Errorf("failed to get price: %v", err)
	}

	updates := make(DataSourceUpdateMap)

	updateTime := time.Now().UTC().UnixMilli()
	updates[c.valueId] = DataSourceValueUpdate{
		Timestamp:    time.UnixMilli(updateTime),
		ValueId:      c.valueId,
		Value:        updateValue,
		DataSourceId: c.GetDataSourceId(),
	}

	return updates, nil
}

func (c *uniswapV2Connector) GetUpdateFrequency() time.Duration {
	return c.updateFrequency
}

func (c *uniswapV2Connector) updateBoundContract() error {
	address := common.HexToAddress(c.uniswapConfig.ContractAddress)

	abiJson, err := resourcesFS.ReadFile("resources/abis/" + uniswapV2AbiFileName)
	if err != nil {
		c.logger.Fatal().Msgf("failed to read ABI file %s: %v", uniswapV2AbiFileName, err)
	}
	abi, err := abi.JSON(strings.NewReader(string(abiJson)))
	if err != nil {
		c.logger.Fatal().Msgf("failed to parse ABI file %s: %v", uniswapV2AbiFileName, err)
	}

	httpClient, err := ethclient.Dial(c.uniswapConfig.HttpProviderUrl + c.apiKey)
	if err != nil {
		return fmt.Errorf("failed to connect to the http client: %v", err)
	}

	contract := bind.NewBoundContract(
		address,
		abi,
		httpClient,
		httpClient,
		httpClient,
	)
	c.contract = contract
	return nil
}

// hit the contract and compute the price
func (c *uniswapV2Connector) getPrice() (float64, error) {
	// retry with exponential backoff
	var result []interface{}
	delay := baseRetryDelay
	var queryError error
	for attempt := 0; attempt < maxQueryAttempts; attempt++ {
		queryError = c.contract.Call(nil, &result, getUniswapV2ContractFunction)
		if queryError != nil {
			c.logger.Warn().Err(queryError).Msgf("Failed to query contract method %s for value id %s (attempt %v)", getUniswapV2ContractFunction, c.valueId, attempt)
			time.Sleep(delay)
			delay = delay * 2
		} else {
			break
		}
	}

	if queryError != nil {
		return -1, fmt.Errorf("failed to hit contract method %s for value id %s : %v", getUniswapV2ContractFunction, c.valueId, queryError)
	}

	return c.calculatePrice(result)
}

// helper function to convert the result object to a useful price
func (c *uniswapV2Connector) calculatePrice(result []interface{}) (float64, error) {
	reserveBase, ok := result[c.uniswapConfig.BaseTokenIndex].(*big.Int)
	if !ok {
		return -1, fmt.Errorf("failed to convert reserveBase size to big int: %v", ok)
	}
	reserveQuote, ok := result[c.uniswapConfig.QuoteTokenIndex].(*big.Int)
	if !ok {
		return -1, fmt.Errorf("failed to convert reserveQuote size to big int: %v", ok)
	}

	reserve0Float := new(big.Float).SetInt(reserveBase)
	reserve1Float := new(big.Float).SetInt(reserveQuote)

	tokenA := new(big.Float).Quo(reserve1Float, reserve0Float)
	price, _ := tokenA.Float64()
	if price == 0 {
		return -1, fmt.Errorf("failed to convert reserve data to tokenA price: %v", price)
	}

	exponent := float64(c.uniswapConfig.BaseTokenDecimals - c.uniswapConfig.QuoteTokenDecimals)
	price = price * math.Pow(10, exponent)

	return price, nil
}

func (c *uniswapV2Connector) GetDataSourceId() DataSourceId {
	return UniswapV2DataSourceId
}

func getUniswapV2DataSources(sourceConfigs []DataProviderSourceConfig) []dataSource {
	dataSources := make([]dataSource, 0)
	for _, sourceConfig := range sourceConfigs {
		connector := newUniswapV2AlchemyConnector(sourceConfig)
		dataSource := newScheduledDataSource(connector)
		dataSources = append(dataSources, dataSource)
	}

	return dataSources
}
