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

// SubscribeNewBlock souscrit à la notification de nouveau bloc
// La fonction prend un objet de type ethclient.Client en entrée et retourne un canal (Feed) qui
// permet de recevoir les informations sur les nouveaux blocs
// Si la souscription échoue, la fonction retourne une erreur
func SubscribeNewBlock(Client *ethclient.Client) (Feed chan *types.Header, err error) {
	// Crée un canal avec une capacité de 2 éléments
	Feed = make(chan *types.Header, 2)
	// Souscrit à la notification de nouveau bloc en utilisant la méthode "SubscribeNewHead()"
	_, err = Client.SubscribeNewHead(context.Background(), Feed)
	return
}

// GetBlockTxs retourne les transactions associées à un bloc donné
// La fonction prend un objet de type ethclient.Client et le numéro de bloc en entrée
// Elle retourne une liste de transactions et une erreur s'il y a lieu
func GetBlockTxs(Client *ethclient.Client, BlockNumber *big.Int) (Txs []*types.Transaction, err error) {
	var Block *types.Block
	// Récupère le bloc en utilisant la méthode "BlockByNumber()"
	Block, err = Client.BlockByNumber(context.Background(), BlockNumber)
	if err != nil {
		return
	}
	// Récupère les transactions associées au bloc en utilisant la méthode "Transactions()"
	Txs = Block.Transactions()
	return
}

func GetMinMaxAmount(ProtocolType string, Tx *types.Transaction) (Result MinMaxResult) {
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
	Result = parseTransaction(*Method, Input, Tx, ABI)
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
