package node

import (
	"../uniswapV2"
	"../uniswapV3"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math"
	"math/big"
	"strings"
)

// SubscribeNewBlock permet de recevoir les informations sur les nouveaux blocs
func SubscribeNewBlock(Client *ethclient.Client) (Feed chan *types.Header, err error) {
	Feed = make(chan *types.Header, 2)
	_, err = Client.SubscribeNewHead(context.Background(), Feed)
	return
}

// GetBlockTxs retourne les transactions associées à un bloc donné
func GetBlockTxs(Client *ethclient.Client, BlockNumber *big.Int) (Txs []*types.Transaction, err error) {
	var Block *types.Block
	Block, err = Client.BlockByNumber(context.Background(), BlockNumber)
	if err != nil {
		return
	}
	// Récupère les transactions associées au bloc en utilisant la méthode "Transactions()"
	Txs = Block.Transactions()
	return
}

func DecodeInputData(ProtocolType string, Tx *types.Transaction) (DecodedInputData DecodedInputDataStruct) {
	// Extraire le premier octet des données de transaction pour obtenir l'identifiant de méthode
	HexMethod := Tx.Data()[:4]

	// Extraire les données de transaction restantes
	HexData := Tx.Data()[4:]

	// Obtenir l'ABI (Application Binary Interface) du protocole Uniswap correspondant
	ABI, err := abiFromProtocolType(ProtocolType)
	if err != nil {
		return
	}

	// Obtenir la méthode et les arguments correspondants à l'identifiant de méthode
	Method, Args, err := getMethodAndArgs(ABI, HexMethod)
	if err != nil {
		return
	}

	// Décompresser les données de transaction à l'aide des arguments correspondants pour obtenir les paramètres de méthode
	Input, err := Args.Unpack(HexData)
	if err != nil {
		return
	}

	// Analyser les paramètres de méthode et extraire les informations minimales et maximales pour l'échange
	DecodedInputData = ParseInputDataTransaction(*Method, Input, Tx)
	return
}

func abiFromProtocolType(ProtocolType string) (ABI abi.ABI, err error) {
	switch ProtocolType {
	case "UniswapV2":
		ABI, err = abi.JSON(strings.NewReader(uniswapV2.UniswapV2MetaData.ABI))
	case "UniswapV3":
		ABI, err = abi.JSON(strings.NewReader(uniswapV3.UniswapV3MetaData.ABI))
	default:
		err = fmt.Errorf("unsupported protocol type %s", ProtocolType)
	}
	return
}

func getMethodAndArgs(ABI abi.ABI, HexMethod []byte) (*abi.Method, abi.Arguments, error) {
	Method, err := ABI.MethodById(HexMethod)
	if err != nil {
		return nil, nil, err
	}

	Args := ABI.Methods[Method.Name].Inputs
	return Method, Args, nil
}

// Wei2Float convertit la valeur de la transaction en Ether
func Wei2Float(Amount *big.Int, decimals int) float64 {
	Big := new(big.Float)
	Big.SetString(Amount.String())
	Float := new(big.Float).Quo(Big, big.NewFloat(math.Pow10(decimals)))
	Float64, _ := Float.Float64()
	return Float64
}
