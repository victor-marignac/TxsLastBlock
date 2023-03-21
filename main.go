package main

import (
	"./config"
	"./node"
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

	TxsFeed := ReadNewBlocks(Sub)
	DecodedTxsFeed := make(chan node.DecodedTxStruct, 500)
	go node.TxDecoder(TxsFeed, DecodedTxsFeed, Client)

	select {}
}

func ReadNewBlocks(Sub chan *types.Header) (TxsFeed chan *types.Transaction) {
	TxsFeed = make(chan *types.Transaction, 500)
	Reader := func(Feed chan *types.Transaction) {
		for {
			NewBlock := <-Sub
			Txs, err := node.GetBlockTxs(Client, NewBlock.Number)
			if err != nil {
				log.Println("Unable to get txs for block", NewBlock.Number.Int64(), ":", err)
				continue
			}
			log.Println(len(Txs), "txs in block", NewBlock.Number.Int64(), ":")
			for _, Tx := range Txs {
				select {
				case Feed <- Tx:
					//log.Println("Tx sent to TxsFeed")
				default:
					log.Println("Tx not sent to TxsFeed")
				}
			}
		}
	}
	go Reader(TxsFeed)
	return
}

func Init() (err error) {
	Client, err = node.Dial(config.InfuraBaseURI + config.InfuraKey)
	if err != nil {
		err = errors.New(fmt.Sprint("node.Dial():", err))
	}
	return
}
