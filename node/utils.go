package node

import (
	"../WETH"
	"../config"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math"
	"math/big"
	"os"
	"strconv"
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
func toFloat(Amount *big.Int, decimals int) float64 {
	Big := new(big.Float)
	Big.SetString(Amount.String())
	Float := new(big.Float).Quo(Big, big.NewFloat(math.Pow10(decimals)))
	Float64, _ := Float.Float64()
	return Float64
}

func floatToString(Float float64) (String string) {
	String = strconv.FormatFloat(Float, 'f', -1, 64)
	return
}

// toBigInt convertit la valeur de la transaction en big.Int
func toBigInt(Amount float64, decimals int) *big.Int {
	Float := new(big.Float).Mul(big.NewFloat(Amount), big.NewFloat(math.Pow10(decimals)))
	BigInt, _ := Float.Int(nil)
	return BigInt
}

// getDecimals récupère le nombre de décimales d'un token
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

// getDecimalsWithCache récupère le nombre de décimales d'un token à partir de la base de données
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

// DisplayTokensAndDecimals affiche les tokens et leurs décimales
func DisplayTokensAndDecimals() {
	// Utilisez la méthode DatabaseCopy() pour obtenir une copie de la base de données
	dbCopy := tokenDecimalsDB.DatabaseCopy()

	// Parcourez la copie de la base de données et affichez les tokens et leurs décimales
	fmt.Println("Tokens and their decimals:")
	for tokenAddress, decimals := range dbCopy {
		fmt.Printf("Token Address: %s, Decimals: %d\n", tokenAddress, decimals)
	}
}

func SaveTokensToFile() error {
	dbCopy := tokenDecimalsDB.DatabaseCopy()
	file, err := os.OpenFile(config.TokenDBFileName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(&dbCopy); err != nil {
		return err
	}
	return nil
}

func LoadTokensFromFile() error {
	file, err := os.Open(config.TokenDBFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", config.TokenDBFileName)
		}
		return err
	}
	defer file.Close()
	var dbCopy map[string]int
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&dbCopy); err != nil {
		return err
	}

	for tokenAddress, decimals := range dbCopy {
		tokenDecimalsDB.WriteTokenDecimals(tokenAddress, decimals)
	}

	return nil
}

//   _____ ____  ____  ____    _     _     ____  _  __
//  /  __//  _ \/  _ \/  _ \  / \   / \ /\/   _\/ |/ /
//  | |  _| / \|| / \|| | \|  | |   | | |||  /  |   /
//  | |_//| \_/|| \_/|| |_/|  | |_/\| \_/||  \_ |   \
//  \____\\____/\____/\____/  \____/\____/\____/\_|\_\
//
