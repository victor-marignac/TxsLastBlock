package node

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"math"
	"math/big"
)

// SimpleTx est une structure de données qui représente une transaction simplifiée
// Elle contient les informations suivantes :
//   - Hash : identifiant de la transaction
//   - Value : valeur de la transaction en Ether
//   - From : adresse de l'expéditeur
//   - To : adresse du destinataire (ou "Contract creation" si la transaction crée un contrat)
type SimpleTx struct {
	Hash  string  `json:"hash"`
	Value float64 `json:"value"`
	From  string  `json:"from"`
	To    string  `json:"to"`
	Input string  `json:"Input"`
}

// TxToSimple prend un objet de type Transaction et retourne un objet de type SimpleTx
// La fonction effectue les conversions nécessaires pour extraire les informations
//	nécessaires et les stocker dans un format plus simple

func TxToSimple(Tx *types.Transaction) (Simple SimpleTx) {
	// Convertit l'identificateur de transaction en chaîne de caractères
	Simple.Hash = Tx.Hash().String()

	// Convertit la valeur de la transaction en Ether
	Simple.Value = Wei2Float(Tx.Value(), 18)

	// Récupère l'adresse de l'expéditeur de la transaction
	From, _ := types.Sender(types.NewLondonSigner(Tx.ChainId()), Tx)
	Simple.From = From.String()

	// Récupère l'adresse du destinataire de la transaction et check s'il s'agit d'un contract Uniswap v2/v3
	if Tx.To() != nil {
		Simple.To = Tx.To().String()
		if Simple.To == "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D" {
			Simple.To = "Uniswap v2 Router Contract"
		} else if Simple.To == "0xE592427A0AEce92De3Edee1F18E0157C05861564" {
			Simple.To = "Uniswap v3 Router Contract"
		}

	} else {
		// Si l'adresse est nulle, cela signifie que la transaction crée un contrat
		Simple.To = "Contract creation"
	}

	// Récupère l'input data de la Transaction
	Simple.Input = hexutil.Encode(Tx.Data())
	return
}

// Wei2Float convertit la valeur de la transaction en Ether
// La conversion implique la division de la valeur de la transaction (en Wei)
// par 10^décimales pour obtenir la valeur en Ether
func Wei2Float(Amount *big.Int, decimals int) float64 {
	Big := new(big.Float)
	Big.SetString(Amount.String())
	Float := new(big.Float).Quo(Big, big.NewFloat(math.Pow10(decimals)))
	Float64, _ := Float.Float64()
	return Float64
}
