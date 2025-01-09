package data_provider

import (
	"fmt"
	"math"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
)

// Define the DataSourceId which will be referenced in the config file and the data_source_registry
const UniswapV2DataSourceId = "UNISWAP_V2"

// Define the structure of the config needed to pull the appropriate data
type uniswapDataSourceConfig struct {
	HttpProviderUrl      string `json:"httpProviderUrl"`
	WsProviderUrl        string `json:"wsProviderUrl"`
	ProviderApiKeyEnvVar string `json:"providerApiKeyEnvVar"`
	ContractAddress      string `json:"contractAddress"`
	BaseTokenIndex       int8   `json:"baseTokenIndex"`
	QuoteTokenIndex      int8   `json:"quoteTokenIndex"`
	BaseTokenDecimals    int8   `json:"baseTokenDecimals,omitempty"`
	QuoteTokenDecimals   int8   `json:"quoteTokenDecimals,omitempty"`
}

const UniswapV2AbiFileName = "uniswap_v2.json"

const getUniswapV2ContractFunction = "getReserves"

func GetUniswapV2DataSources(sourceConfigs []DataProviderSourceConfig) []dataSource {
	dataSources := make([]dataSource, 0)
	for _, sourceConfig := range sourceConfigs {
		var UniswapPullConfig uniswapDataSourceConfig
		mapstructure.Decode(sourceConfig.Config, &UniswapPullConfig)
		connector := NewUniswapV2AlchemyConnector(UniswapPullConfig)
		dataSource := NewEthLogListenerDataSource(sourceConfig, connector)
		dataSources = append(dataSources, dataSource)
	}

	return dataSources
}

type UniswapV2AlchemyConnector struct {
	logger        zerolog.Logger
	uniswapConfig uniswapDataSourceConfig
}

func NewUniswapV2AlchemyConnector(uniswapConfig uniswapDataSourceConfig) *UniswapV2AlchemyConnector {
	return &UniswapV2AlchemyConnector{
		logger:        dataSourceLogger(UniswapV2DataSourceId),
		uniswapConfig: uniswapConfig,
	}
}

func (c *UniswapV2AlchemyConnector) GetUpdateValue(contract *bind.BoundContract, valueId ValueId) (float64, error) {
	// Get price from the contract
	var result []interface{}
	delay := BaseRetryDelay
	var queryError error
	for attempt := 0; attempt < MaxQueryAttempts; attempt++ {
		queryError = contract.Call(nil, &result, getUniswapV2ContractFunction)
		if queryError != nil {
			c.logger.Warn().Err(queryError).Msgf("Failed to query contract method %s for value id %s (attempt %v)", getUniswapV2ContractFunction, valueId, attempt)
			time.Sleep(delay)
			// exponentially increase delay
			delay = delay * 2
		} else {
			break
		}
	}

	if queryError != nil {
		return -1, fmt.Errorf("failed to hit contract method %s for value id %s : %v", getUniswapV2ContractFunction, valueId, queryError)
	}

	return c.calculatePrice(result)
}

// helper function to convert the result object to a useful price
func (c *UniswapV2AlchemyConnector) calculatePrice(result []interface{}) (float64, error) {
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

func (c *UniswapV2AlchemyConnector) GetEthLogListenerConnectionId(config DataProviderSourceConfig) EthLogListenerConnectionId {
	apiKey, exists := os.LookupEnv(c.uniswapConfig.ProviderApiKeyEnvVar)
	if !exists {
		panic("Must set an environment variable named " + c.uniswapConfig.ProviderApiKeyEnvVar)
	}
	return EthLogListenerConnectionId{
		wsProviderUrl:   c.uniswapConfig.WsProviderUrl + apiKey,
		httpProviderUrl: c.uniswapConfig.HttpProviderUrl + apiKey,
		abiFilename:     UniswapV2AbiFileName,
		apiKeyEnvVar:    c.uniswapConfig.ProviderApiKeyEnvVar,
	}
}

func (c *UniswapV2AlchemyConnector) GetContractId() common.Address {
	return common.HexToAddress(c.uniswapConfig.ContractAddress)
}

func (c *UniswapV2AlchemyConnector) GetDataSourceId() DataSourceId {
	return UniswapV2DataSourceId
}
