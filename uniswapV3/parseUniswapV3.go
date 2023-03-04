package uniswapV3

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"log"
	"os"
)

func LoadUniswapV3ABI() abi.ABI {

	// charger l'interface ABI d'un fichier JSON
	abiFile, err := os.Open("./uniswapV3/ABIUniswapV3.json")
	if err != nil {
		log.Fatal(err)
	}

	// lire l'interface ABI
	var AbiUniswapV3Router abi.ABI
	if err := json.NewDecoder(abiFile).Decode(&AbiUniswapV3Router); err != nil {
		log.Fatal(err)
	}

	return AbiUniswapV3Router
}
