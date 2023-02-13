package main

import (
	"./config"
	"./node"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"log"
)

var Client *ethclient.Client

func main() {
	err := Init()
	if err != nil {
		log.Println("Initialization error:", err)
		return
	}

	Sub, err := node.SubscribeNewBlock(Client)
	if err != nil {
		log.Println("Unable to subscribe:", err)
		return
	}

	ReadNewBlocks(Sub)
}

func ReadNewBlocks(Sub chan *types.Header) {
	for {
		NewBlock := <-Sub
		Txs, err := node.GetBlockTxs(Client, NewBlock.Number)
		if err != nil {
			log.Println("Unable to get txs for block", NewBlock.Number.Int64(), ":", err)
			continue
		}
		log.Println(len(Txs), "txs in block", NewBlock.Number.Int64(), ":")
		for Index, Tx := range Txs {
			SimpleTx := node.TxToSimple(Tx)
			RawJson, err := json.Marshal(SimpleTx)
			if err != nil {
				log.Println("Unable to marshal", Tx.Hash().String())
				continue
			}
			log.Println("Tx", Index+1, ":", string(RawJson))
			println("\n")
		}
	}
}

func Init() (err error) {
	Client, err = node.Dial(config.InfuraBaseURI + config.InfuraKey)
	if err != nil {
		err = errors.New(fmt.Sprint("node.Dial():", err))
	}
	return
}
