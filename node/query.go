package node

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
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
