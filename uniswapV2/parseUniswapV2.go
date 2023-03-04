package uniswapV2

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/abi"
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
