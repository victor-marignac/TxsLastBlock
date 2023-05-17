package main

import (
	"./config"
	"./node"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"log"
)

// 0x15618650000000f0zef0zef0ezf0zefze1 = func UneFonction(UnArgument string)
// 1561865 = Nom de la fonction
// 0000000f0zef0zef0ezf0zefze1 = Argument
// router.SwapExactTokensForTokens(

// https://etherscan.io/tx/0x5a704adb810af8bd777a6b6289bd8e3e00447396686ecb2a037cea6d2a088694
// 0xb6f9de95000000000000000000000000000000000000000000000000000000f8740fc833000000000000000000000000000000000000000000000000000000000000008000000000000000000000000098355f02ce847a286e9dd2271b98896ab17d8201000000000000000000000000000000000000000000000000000000006426efc90000000000000000000000000000000000000000000000000000000000000002000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000548c14df2ab003ef18ea6b4d2c0d9706953c732c
// 0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D = router
// 0xb6f9de95 = swapExactETHForTokensSupportingFeeOnTransferTokens(uint256 amountOutMin, address[] path, address to, uint256 deadline)
// [0]:  000000000000000000000000000000000000000000000000000000f8740fc833 = 1067099080755
// [1]:  0000000000000000000000000000000000000000000000000000000000000080 // <-
// [2]:  00000000000000000000000098355f02ce847a286e9dd2271b98896ab17d8201 = 0x98355F02CE847a286e9dD2271b98896Ab17D8201
// [3]:  000000000000000000000000000000000000000000000000000000006426efc9 = 1680273353
// [4]:  0000000000000000000000000000000000000000000000000000000000000002 // <-
// [5]:  000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2 = 0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2
// [6]:  000000000000000000000000548c14df2ab003ef18ea6b4d2c0d9706953c732c = 0x548c14df2AB003eF18ea6b4d2C0D9706953C732c

// 0xb6f9de95000000000000000000000000000000000000000000000000000000f8740fc833000000000000000000000000000000000000000000000000000000000000008000000000000000000000000098355f02ce847a286e9dd2271b98896ab17d8201000000000000000000000000000000000000000000000000000000006426efc90000000000000000000000000000000000000000000000000000000000000002000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000548c14df2ab003ef18ea6b4d2c0d9706953c732c
// =
// router.swapExactETHForTokensSupportingFeeOnTransferTokens(1067099080755, [0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2,0x548c14df2AB003eF18ea6b4d2C0D9706953C732c], 0x98355F02CE847a286e9dD2271b98896Ab17D8201, 1680273353)
// pool.swap(<args>)

// https://docs.uniswap.org/ -> https://docs.uniswap.org/contracts/v3/overview ->

func main() {
	node.Sync.Add(1)

	err := Init()
	if err != nil {
		log.Println("Initialization error:", err)
		return
	}

	node.LoadTokensFromFile()

	Sub, err := node.SubscribeNewBlock(node.Client)
	if err != nil {
		log.Println("Unable to subscribe:", err)
		return
	}

	TxsFeed := ReadNewBlocks(Sub)
	DecodedTxsFeed := make(chan node.DecodedTx, 500)
	go node.TxDecoder(TxsFeed, DecodedTxsFeed)

	node.Sync.Wait()
	Shutdown()
}

func Shutdown() {
	log.Println("Shutting down gracefully..")
	node.SaveTokensToFile()
	// save db
	// whatever..
}

func ReadNewBlocks(Sub chan *types.Header) chan node.LocalTx {
	var TxsFeed = make(chan node.LocalTx, 1000)
	Reader := func(Feed chan node.LocalTx) {
		for {
			NewBlock := <-Sub
			Txs, err := node.GetBlockTxs(node.Client, NewBlock.Number)
			if err != nil {
				log.Println("Unable to get txs for block", NewBlock.Number.Int64(), ":", err)
				continue
			}
			Logs, err := node.Client.FilterLogs(context.Background(), ethereum.FilterQuery{FromBlock: NewBlock.Number, ToBlock: NewBlock.Number})
			if err != nil {
				log.Println("Unable to get logs for block", NewBlock.Number.Int64(), ":", err)
				continue
			} else if len(Logs) == 0 {
				log.Println("Empty logs for block", NewBlock.Number.Int64())
			}
			log.Println(len(Txs), "txs in block", NewBlock.Number.Int64(), "(", len(Logs), "events )")
			//node.DisplayTokensAndDecimals()
			var LogsByTxs = make(map[common.Hash][]*types.Log)
			for _, l := range Logs {
				Log := l
				LogsByTxs[Log.TxHash] = append(LogsByTxs[Log.TxHash], &Log)
			}
			for _, Tx := range Txs {
				NewTx, err := node.TxToLocalTx(Tx, LogsByTxs[Tx.Hash()])
				if err != nil {
					log.Println("ParseInputData() error:", err, "(", NewTx.Hash, ")")
					continue
				}
				if len(NewTx.Logs) == 0 {
					continue
				}
				select {
				case Feed <- NewTx:
					//log.Println("Tx sent to TxsFeed")
				default:
					log.Println("Tx not sent to TxsFeed")
				}
			}
		}
	}
	go Reader(TxsFeed)
	return TxsFeed
}

func Init() (err error) {
	node.Client, err = node.Dial(config.InfuraBaseURI + config.InfuraKey)
	if err != nil {
		err = errors.New(fmt.Sprint("node.Dial():", err))
	}
	return
}
