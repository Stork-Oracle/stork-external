package sources

import (
	"embed"
	"fmt"
	"strings"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
)

const (
	maxQueryAttempts = 3
	baseRetryDelay   = 500 * time.Millisecond
)

func GetEthereumContract(
	contractAddress string,
	abiFileName string,
	httpProviderUrl string,
	apiKey string,
	resourcesFs embed.FS,
) (*bind.BoundContract, error) {
	address := common.HexToAddress(contractAddress)

	abiJson, err := resourcesFs.ReadFile("resources/abis/" + abiFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read ABI file %s: %v", abiFileName, err)
	}
	abi, err := abi.JSON(strings.NewReader(string(abiJson)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI file %s: %v", abiFileName, err)
	}

	httpClient, err := ethclient.Dial(httpProviderUrl + apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the http client: %v", err)
	}

	contract := bind.NewBoundContract(
		address,
		abi,
		httpClient,
		httpClient,
		httpClient,
	)
	return contract, nil
}

func CallEthereumFunction(
	contract *bind.BoundContract,
	functionName string,
	valueId types.ValueId,
	logger zerolog.Logger,
) ([]interface{}, error) {
	// retry with exponential backoff
	var result []interface{}
	delay := baseRetryDelay
	var queryError error
	for attempt := 0; attempt < maxQueryAttempts; attempt++ {
		queryError = contract.Call(nil, &result, functionName)
		if queryError != nil {
			logger.Warn().Err(queryError).Msgf("Failed to query contract method %s for value id %s (attempt %v): %v", functionName, valueId, attempt, queryError)
			time.Sleep(delay)
			delay = delay * 2
		} else {
			break
		}
	}

	if queryError != nil {
		return nil, fmt.Errorf("failed to hit contract method %s for value id %s: %v", functionName, valueId, queryError)
	}

	return result, nil
}
