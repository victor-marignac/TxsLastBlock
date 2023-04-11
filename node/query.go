package node

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
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
