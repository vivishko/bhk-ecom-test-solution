package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

var aggregatorV3InterfaceABI = `[{"inputs":[],"name":"decimals","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"description","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"version","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint80","name":"_roundId","type":"uint80"}],"name":"getRoundData","outputs":[{"internalType":"uint80","name":"roundId","type":"uint80"},{"internalType":"int256","name":"answer","type":"int256"},{"internalType":"uint256","name":"startedAt","type":"uint256"},{"internalType":"uint256","name":"updatedAt","type":"uint256"},{"internalType":"uint80","name":"answeredInRound","type":"uint80"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"latestRoundData","outputs":[{"internalType":"uint80","name":"roundId","type":"uint80"},{"internalType":"int256","name":"answer","type":"int256"},{"internalType":"uint256","name":"startedAt","type":"uint256"},{"internalType":"uint256","name":"updatedAt","type":"uint256"},{"internalType":"uint80","name":"answeredInRound","type":"uint80"}],"stateMutability":"view","type":"function"}]`

var erc20ABI = `[{"constant":true,"inputs":[{"name":"account","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"type":"function"},
{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"payable":false,"type":"function"}]`

var (
	ethUsdAggregator = common.HexToAddress("0x5f4ec3df9cbd43714fe2740f5e3616155c5b8419")
	wethToken        = common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27ead9083C756Cc2")
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go [ETH_ADDRESS]")
		return
	}

	address := common.HexToAddress(os.Args[1])

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	infuraApiKey := os.Getenv("INFURA_APIKEY")
    if infuraApiKey == "" {
        log.Println("INFURA_APIKEY not set")
    }
	rpcURL := fmt.Sprintf("https://mainnet.infura.io/v3/%s", infuraApiKey)
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum node: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	balanceWei, err := client.BalanceAt(ctx, address, nil)
	if err != nil {
		log.Fatalf("Error getting ETH balance: %v", err)
	}
	balanceEth := new(big.Float).Quo(new(big.Float).SetInt(balanceWei), big.NewFloat(1e18))

	priceEthUsd, err := getLatestPrice(ctx, client, ethUsdAggregator)
	if err != nil {
		log.Fatalf("Error getting ETH price: %v", err)
	}

	ethInUsd := new(big.Float).Mul(balanceEth, priceEthUsd)

	wethBalance, wethDecimals, err := getERC20Balance(ctx, client, wethToken, address)
	if err != nil {
		log.Fatalf("Error getting WETH balance: %v", err)
	}
	wethBalanceFloat := new(big.Float).Quo(new(big.Float).SetInt(wethBalance), big.NewFloat(math.Pow10(int(wethDecimals))))

	wethInUsd := new(big.Float).Mul(wethBalanceFloat, priceEthUsd)

	fmt.Printf("Address: %s\n", address.Hex())
	fmt.Printf("ETH balance: %s ETH (~%s USD)\n", balanceEth.String(), ethInUsd.String())
	fmt.Printf("WETH balance: %s WETH (~%s USD)\n", wethBalanceFloat.String(), wethInUsd.String())

	totalUsd := new(big.Float).Add(ethInUsd, wethInUsd)
	fmt.Printf("Total value in USD: %s\n", totalUsd.String())
}

func getLatestPrice(ctx context.Context, client *ethclient.Client, aggregatorAddr common.Address) (*big.Float, error) {
	parsedABI, err := abi.JSON(strings.NewReader(aggregatorV3InterfaceABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse aggregator ABI: %w", err)
	}

	contract := bind.NewBoundContract(aggregatorAddr, parsedABI, client, client, client)

	var decimals uint8
	outDecimals := []interface{}{&decimals}
	err = contract.Call(nil, &outDecimals, "decimals")
	if err != nil {
		return nil, fmt.Errorf("failed to get decimals: %w", err)
	}

	var roundId, answer, startedAt, updatedAt, answeredInRound big.Int
	out := []interface{}{&roundId, &answer, &startedAt, &updatedAt, &answeredInRound}
	err = contract.Call(nil, &out, "latestRoundData")
	if err != nil {
		return nil, fmt.Errorf("failed to call latestRoundData: %w", err)
	}

	price := new(big.Float).Quo(new(big.Float).SetInt(&answer), big.NewFloat(math.Pow10(int(decimals))))

	return price, nil
}

func getERC20Balance(ctx context.Context, client *ethclient.Client, token common.Address, holder common.Address) (*big.Int, uint8, error) {
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse erc20 ABI: %w", err)
	}

	contract := bind.NewBoundContract(token, parsedABI, client, client, client)

	var balance *big.Int
	outBalance := []interface{}{&balance}
	err = contract.Call(nil, &outBalance, "balanceOf", holder)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get balanceOf: %w", err)
	}

	var decimals uint8
	outDecimals := []interface{}{&decimals}
	err = contract.Call(nil, &outDecimals, "decimals")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get decimals: %w", err)
	}

	return balance, decimals, nil
}