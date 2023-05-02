package node

import (
	"../WETH"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math"
	"math/big"
)

func (l *LocalTx) ToString() string {
	return fmt.Sprintf("Hash: %v, From: %v, To: %v, Value: %v eth, GasLimit: %v, Nonce: %v, Logs: %v",
		l.Hash, l.From, l.To, l.Value, l.GasLimit /*l.GasUsed,*/, l.Nonce /*l.Index, */, l.Logs /*, l.Revert, l.ContractAddress*/)
}

func (q *QueryData) ToString() string {
	return fmt.Sprintf("Protocol: %v, Type: %v, TokenIn: %v, TokenOut: %v, Amount: %v, MinMax: %v",
		q.Protocol, q.Type, q.TokenIn, q.TokenOut, q.Amount, q.MinMax)
}

func (e *Event) ToString() string {
	return fmt.Sprintf("Protocol: %v, Contract: %v, Type: %v, Pool: %v, TokenIn: %v, TokenOut: %v, AmountIn: %v, AmountOut: %v",
		e.Protocol, e.Contract, e.Type, e.Pool, e.TokenIn, e.TokenOut, e.AmountIn, e.AmountOut)
}

func (Tx *DecodedTx) Log() {
	var Events string
	for _, Event := range Tx.Events {
		Events = Events + Event.ToString() + ","
	}
	log.Println(Tx.Tx.ToString(), "\n", "| QueryData:", Tx.Query.ToString(), "\n", "| Events:", Events, "\n")
}

// toFloat convertit la valeur de la transaction en Ether
// Ethereum = 18 décimales
// USDC = 6 décimales
// Comment trouver les bonnes décimales ? :)
// 1.5 Ether :
// big.Int = 150000000000000000
// float64 = 1.5
func toFloat(Amount *big.Int, decimals int) float64 {
	Big := new(big.Float)
	Big.SetString(Amount.String())
	Float := new(big.Float).Quo(Big, big.NewFloat(math.Pow10(decimals)))
	Float64, _ := Float.Float64()
	return Float64
}

func getDecimals(client *ethclient.Client, contractAddress common.Address) (int, error) {
	contract, err := WETH.NewWETH(contractAddress, client)
	if err != nil {
		return 0, err
	}

	callOpts := &bind.CallOpts{Context: context.Background()}
	decimals, err := contract.Decimals(callOpts)
	if err != nil {
		return 0, err
	}

	return int(decimals), nil
}

func getDecimalsWithCache(client *ethclient.Client, contractAddress common.Address) (int, error) {
	// Vérifiez si l'adresse du token existe dans la base de données
	decimals, err := tokenDecimalsDB.ReadTokenDecimals(contractAddress.String())
	if err == nil {
		// Si les décimales existent, renvoyez-les
		return decimals, nil
	}

	// Sinon, récupérez les décimales à l'aide de la fonction getDecimals
	decimals, err = getDecimals(client, contractAddress)
	if err != nil {
		return 0, err
	}

	// Enregistrez les décimales dans la base de données pour une utilisation ultérieure
	tokenDecimalsDB.WriteTokenDecimals(contractAddress.String(), decimals)

	return decimals, nil
}

//   _____ ____  ____  ____    _     _     ____  _  __
//  /  __//  _ \/  _ \/  _ \  / \   / \ /\/   _\/ |/ /
//  | |  _| / \|| / \|| | \|  | |   | | |||  /  |   /
//  | |_//| \_/|| \_/|| |_/|  | |_/\| \_/||  \_ |   \
//  \____\\____/\____/\____/  \____/\____/\____/\_|\_\
//
