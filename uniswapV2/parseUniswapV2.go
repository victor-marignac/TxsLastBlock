package uniswapV2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"log"
	"os"
)

func LoadUniswapV2ABI() abi.ABI {

	// charger l'interface ABI d'un fichier JSON
	abiFile, err := os.Open("./uniswapV2/ABIUniswapV2.json")
	if err != nil {
		log.Fatal(err)
	}

	// lire l'interface ABI
	var AbiUniswapV2Router abi.ABI
	if err := json.NewDecoder(abiFile).Decode(&AbiUniswapV2Router); err != nil {
		log.Fatal(err)
	}

	return AbiUniswapV2Router
}

func DecodeTransactionInput(input string, abiContract abi.ABI) (string, []interface{}, error) {

	parsed, err := hexutil.Decode(input)
	if err != nil {
		return "", nil, err
	}

	for name, method := range abiContract.Methods {
		if bytes.HasPrefix(parsed, method.ID) {
			args, err := method.Inputs.UnpackValues(parsed[len(method.ID):])
			if err != nil {
				return "", nil, err
			}
			return name, args, nil
		}
	}

	return "", nil, fmt.Errorf("unknown method ID: %x", parsed[:4])
}
