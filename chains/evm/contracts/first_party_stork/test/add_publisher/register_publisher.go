package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Stork-Oracle/stork-external/apps/first_party_pusher/pkg/evm/bindings"
)

// TODO: temp util for testing
func main() {
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal("Failed to connect to Ethereum client:", err)
	}

	contractAddress := common.HexToAddress("0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512")
	publisherAddress := common.HexToAddress("0x99e295e85cb07C16B7BB62A44dF532A7F2620237")
	fee := big.NewInt(0)

	// default private key for the first account in hardhat as that is the default owner account when deploying locally
	privateKeyHex := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatal("Failed to parse private key:", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(31337)) // Hardhat chain ID
	if err != nil {
		log.Fatal("Failed to create transactor:", err)
	}

	contract, err := bindings.NewFirstPartyStorkContract(contractAddress, client)
	if err != nil {
		log.Fatal("Failed to create contract instance:", err)
	}

	fmt.Printf("Registering publisher: %s\n", publisherAddress.Hex())
	fmt.Printf("Contract address: %s\n", contractAddress.Hex())
	fmt.Printf("Fee: %s\n", fee.String())

	tx, err := contract.CreatePublisherUser(auth, publisherAddress, fee)
	if err != nil {
		log.Fatal("Failed to call createPublisherUser:", err)
	}

	fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())

	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatal("Failed to wait for transaction:", err)
	}

	fmt.Printf("Transaction mined in block: %d\n", receipt.BlockNumber.Uint64())
	fmt.Println("Publisher registered successfully!")
}
