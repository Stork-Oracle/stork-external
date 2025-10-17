package sources

import (
	"embed"
	"fmt"
	"strings"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/types"
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
	abiFilePath string,
	httpProviderUrl string,
	resourcesFs embed.FS,
) (*bind.BoundContract, error) {
	address := common.HexToAddress(contractAddress)

	abiJson, err := resourcesFs.ReadFile(abiFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read ABI file %s: %v", abiFilePath, err)
	}
	abi, err := abi.JSON(strings.NewReader(string(abiJson)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI file %s: %v", abiFilePath, err)
	}

	httpClient, err := ethclient.Dial(httpProviderUrl)
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
	valueID types.ValueID,
	logger zerolog.Logger,
) ([]any, error) {
	// retry with exponential backoff
	var result []any
	delay := baseRetryDelay
	var queryError error
	for attempt := 0; attempt < maxQueryAttempts; attempt++ {
		queryError = contract.Call(nil, &result, functionName)
		if queryError != nil {
			logger.Warn().
				Err(queryError).
				Msgf("Failed to query contract method %s for value id %s (attempt %v): %v", functionName, valueID, attempt, queryError)
			time.Sleep(delay)
			delay = delay * 2
		} else {
			break
		}
	}

	if queryError != nil {
		return nil, fmt.Errorf(
			"failed to hit contract method %s for value id %s: %v",
			functionName,
			valueID,
			queryError,
		)
	}

	return result, nil
}
