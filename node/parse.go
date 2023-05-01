package node

import (
	"../WETH"
	"../config"
	"../uniswapV2/pool"
	"../uniswapV2/router"
	"../uniswapV3"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"log"
	"math/big"
	"strings"
)

func TxDecoder(TxsFeed chan LocalTx, DecodedTxsFeed chan DecodedTx) {
	// fonction anonyme dans une routine pour print le log de DecodedTx
	go func() {
		for {
			if Tx, Closed := <-DecodedTxsFeed; !Closed {
				log.Println("Channel closed (DecodedTxsFeed)")
			} else if len(Tx.Events) != 0 && len(Tx.Query.Protocol) != 0 {
				Tx.Log()
			}
		}
	}()

	// Cette boucle va écrire une DecodedTx dans un channel DecodedTxsFeed
	for {
		if Tx, Closed := <-TxsFeed; !Closed {
			log.Println("Channel closed (TxsFeed)")
			break
		} else {
			var NewTx DecodedTx
			NewTx.Tx = Tx

			err := NewTx.ParseInputData()
			if err != nil {
				log.Println("ParseInputData() error:", err, "(", NewTx.Tx.Hash, ")")
			}
			NewTx.ParseReceipt()

			// Envoyez la structure DecodedTx au canal DecodedTxsFeed
			DecodedTxsFeed <- NewTx
		}
	}
}

func TxToLocalTx(Tx *types.Transaction, Logs []*types.Log) (NewTx LocalTx, err error) {
	NewTx.Hash = Tx.Hash().String()
	NewTx.Value = toFloat(Tx.Value(), 18)
	From, _ := types.Sender(types.NewLondonSigner(Tx.ChainId()), Tx)
	NewTx.From = From.String()
	NewTx.Nonce = Tx.Nonce()
	NewTx.GasLimit = Tx.Gas()
	if Tx.To() != nil {
		NewTx.To = Tx.To().String()
	} else {
		// Si l'adresse est nulle, cela signifie que la transaction crée un contrat
		NewTx.To = "Contract creation"
	}
	/*NewTx.Tx.GasUsed = receipt.GasUsed
	NewTx.Tx.ContractAddress = receipt.ContractAddress.Hex()*/
	NewTx.Logs = Logs
	/*NewTx.Tx.Revert = receipt.Status != types.ReceiptStatusSuccessful*/
	//NewTx.Index =
	NewTx.RawTx = *Tx
	return
}

func (Tx *DecodedTx) ParseInputData() (err error) {
	var ProtocolType string
	switch Tx.Tx.To {
	case config.UniswapV2ContractRouter:
		ProtocolType = "UniswapV2Router"
	case config.UniswapV3ContractRouter:
		ProtocolType = "UniswapV3Router"
	default:

	}
	if len(ProtocolType) != 0 {
		var Result QueryData
		if Result, err = DecodeInputData(ProtocolType, &Tx.Tx.RawTx); err == nil {
			Tx.Query = Result
		}
	}
	return
}

func (Tx *DecodedTx) ParseReceipt() {
	ABI, err := abiFromProtocolType("UniswapV2Pool")
	if err != nil {
		log.Fatalf("Erreur lors de la création de l'ABI: %v", err)
	}

	for _, l := range Tx.Tx.Logs {
		var SwapEvent pool.PoolSwap
		if len(l.Topics) == 0 {
			log.Printf("Aucun topic trouvé dans les logs")
			continue
		}
		// Amount0In = Token 0 In
		// Amount1In = Token 1 In
		// Amount0Out = Token 0 Out
		// Amount1Out = Token 1 Out
		// Token 0 = WETH
		// Token 1 = USDC
		// Si Amount0In = 1, Amount1In = 0, Amount0Out = 0, Amount1Out=1700
		// Ca veut dire que c'est un trade de 1 WETH -> 1700 USDC
		// Si Amount0In = 0, Amount1In = 1700, Amount0Out = 1, Amount1Out=0
		// Ca veut dire que c'est un trade de 1700 USDC -> 1 WETH
		switch l.Topics[0].Hex() {
		case config.UniswapV2EventSwap:
			err = ABI.UnpackIntoInterface(&SwapEvent, "Swap", l.Data)
			if err != nil {
				log.Printf("Erreur lors de l'extraction des données de l'événement Swap: %v", err)
				continue
			}

			amountIn, amountOut := new(big.Int), new(big.Int)

			if SwapEvent.Amount0In != nil && SwapEvent.Amount1In != nil {
				amountIn.Add(SwapEvent.Amount0In, SwapEvent.Amount1In)
			}

			if SwapEvent.Amount0Out != nil && SwapEvent.Amount1Out != nil {
				amountOut.Add(SwapEvent.Amount0Out, SwapEvent.Amount1Out)
			}

			tokenInAddress := common.HexToAddress(Tx.Query.TokenIn)
			tokenOutAddress := common.HexToAddress(Tx.Query.TokenOut)

			decimalsIn, _ := getDecimals(Client, tokenInAddress)
			decimalsOut, _ := getDecimals(Client, tokenOutAddress)

			Tx.Events = append(Tx.Events, Event{
				Protocol:  "Uniswap V2",
				Contract:  Tx.Tx.To,
				Type:      "Swap",
				Pool:      l.Address.String(),
				TokenIn:   Tx.Query.TokenIn,
				TokenOut:  Tx.Query.TokenOut,
				AmountIn:  toFloat(amountIn, decimalsIn),
				AmountOut: toFloat(amountOut, decimalsOut),
			})
		default:
			//log.Printf("Event with topic %s not supported", l.Topics[0].Hex())
		}
	}
}

func DecodeInputData(ProtocolType string, Tx *types.Transaction) (DecodedInputData QueryData, err error) {
	// Extraire le premier octet des données de transaction pour obtenir l'identifiant de méthode
	HexMethod := Tx.Data()[:4]
	// Extraire les données de transaction restantes
	HexData := Tx.Data()[4:]
	var ABI abi.ABI
	var Method *abi.Method
	var Args abi.Arguments
	var Input []interface{}
	if ABI, err = abiFromProtocolType(ProtocolType); err == nil {
		if Method, Args, err = getMethodAndArgs(ABI, HexMethod); err == nil {
			if Input, err = Args.Unpack(HexData); err == nil {
				DecodedInputData = ParseInputDataTransaction(*Method, Input, Tx)
			}
		}
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

func abiFromProtocolType(ProtocolType string) (ABI abi.ABI, err error) {
	switch ProtocolType {
	case "UniswapV2Router":
		ABI, err = abi.JSON(strings.NewReader(router.RouterMetaData.ABI))
	case "UniswapV3Router":
		ABI, err = abi.JSON(strings.NewReader(uniswapV3.UniswapV3MetaData.ABI))
	case "UniswapV2Pool":
		ABI, err = abi.JSON(strings.NewReader(pool.PoolMetaData.ABI))
	case "WETH":
		ABI, err = abi.JSON(strings.NewReader(WETH.WETHMetaData.ABI))
	default:
		err = fmt.Errorf("unsupported protocol type %s", ProtocolType)
	}
	return
}

func ParseInputDataTransaction(Method abi.Method, Input []interface{}, Tx *types.Transaction) (Q QueryData) {

	switch Method.Name {
	// Uniswap v2
	case "swapExactTokensForTokens":
		Path, err := Input[2].([]common.Address)
		if !err {
			log.Println("Erreur d'assertion Input[2] as []common.Address", err)
		}
		decimalsIn, _ := getDecimals(Client, Path[0])
		decimalsOut, _ := getDecimals(Client, Path[len(Path)-1])
		Q.Protocol = "Uniswap V2"
		Q.Type = Method.Name
		Q.Amount = toFloat(Input[0].(*big.Int), decimalsIn)
		Q.MinMax = toFloat(Input[1].(*big.Int), decimalsOut)

		log.Println("Query decimal In", decimalsIn)
		log.Println("Query decimal Out", decimalsOut)

		Q.TokenIn = Path[0].String()
		Q.TokenOut = Path[len(Path)-1].String()

	case "swapTokensForExactTokens":
		Path, err := Input[2].([]common.Address)
		if !err {
			log.Println("Erreur d'assertion Input[2] as []common.Address", err)
		}
		Q.Protocol = "Uniswap V2"
		decimals, _ := getDecimals(Client, Path[0])
		decimalsOut, _ := getDecimals(Client, Path[len(Path)-1])
		Q.Type = Method.Name
		Q.Amount = toFloat(Input[0].(*big.Int), decimals)
		Q.MinMax = toFloat(Input[1].(*big.Int), decimalsOut)

		log.Println("Query decimal In", decimals)
		log.Println("Query decimal Out", decimalsOut)

		Q.TokenIn = Path[0].String()
		Q.TokenOut = Path[len(Path)-1].String()
	case "swapExactETHForTokens":
		Path, err := Input[1].([]common.Address)
		if !err {
			log.Println("Erreur d'assertion Input[2] as []common.Address", err)
		}
		decimals, _ := getDecimals(Client, Path[0])
		decimalsOut, _ := getDecimals(Client, Path[len(Path)-1])
		Q.Protocol = "Uniswap V2"
		Q.Type = Method.Name
		Q.Amount = toFloat(Tx.Value(), decimals)
		Q.MinMax = toFloat(Input[0].(*big.Int), decimalsOut)

		log.Println("Query decimal In", decimals)
		log.Println("Query decimal Out", decimalsOut)

		Q.TokenIn = Path[0].String()
		Q.TokenOut = Path[len(Path)-1].String()
	case "swapTokensForExactETH":
		Path, err := Input[1].([]common.Address)
		if !err {
			log.Println("Erreur d'assertion Input[2] as []common.Address", err)
		}
		decimals, _ := getDecimals(Client, Path[0])
		decimalsOut, _ := getDecimals(Client, Path[len(Path)-1])
		Q.Protocol = "Uniswap V2"
		Q.Type = Method.Name
		Q.Amount = toFloat(Input[0].(*big.Int), decimals)
		Q.MinMax = toFloat(Tx.Value(), decimalsOut)

		log.Println("Query decimal In", decimals)
		log.Println("Query decimal Out", decimalsOut)

		Q.TokenIn = Path[0].String()
		Q.TokenOut = Path[len(Path)-1].String()
	case "swapExactTokensForETH":
		Path, err := Input[2].([]common.Address)
		if !err {
			log.Println("Erreur d'assertion Input[2] as []common.Address", err)
		}
		decimals, _ := getDecimals(Client, Path[0])
		decimalsOut, _ := getDecimals(Client, Path[len(Path)-1])
		Q.Protocol = "Uniswap V2"
		Q.Type = Method.Name
		Q.Amount = toFloat(Input[0].(*big.Int), decimals)
		Q.MinMax = toFloat(Input[1].(*big.Int), decimalsOut)

		Q.TokenIn = Path[0].String()
		Q.TokenOut = Path[len(Path)-1].String()
	case "swapETHForExactTokens":
		Path, err := Input[1].([]common.Address)
		if !err {
			log.Println("Erreur d'assertion Input[2] as []common.Address")
		}
		decimals, _ := getDecimals(Client, Path[0])
		decimalsOut, _ := getDecimals(Client, Path[len(Path)-1])

		Q.Protocol = "Uniswap V2"
		Q.Type = Method.Name
		Q.Amount = toFloat(Input[0].(*big.Int), decimals)
		Q.MinMax = toFloat(Tx.Value(), decimalsOut)

		Q.TokenIn = Path[0].String()
		Q.TokenOut = Path[len(Path)-1].String()
	// Uniswap v3
	case "exactInputSingle":
		Q.Protocol = "Uniswap V3"
		Params := Input[0].(struct {
			TokenIn           common.Address "json:\"tokenIn\""
			TokenOut          common.Address "json:\"tokenOut\""
			Fee               *big.Int       "json:\"fee\""
			Recipient         common.Address "json:\"recipient\""
			Deadline          *big.Int       "json:\"deadline\""
			AmountIn          *big.Int       "json:\"amountIn\""
			AmountOutMinimum  *big.Int       "json:\"amountOutMinimum\""
			SqrtPriceLimitX96 *big.Int       "json:\"sqrtPriceLimitX96\""
		})
		Q.Type = Method.Name
		Q.Amount = toFloat(Params.AmountIn, 18)
		Q.MinMax = toFloat(Params.AmountOutMinimum, 6)
		Q.TokenIn = Params.TokenIn.String()
		Q.TokenOut = Params.TokenOut.String()
	case "exactOutputSingle":
		Q.Protocol = "Uniswap V3"
		Params := Input[0].(struct {
			TokenIn           common.Address "json:\"tokenIn\""
			TokenOut          common.Address "json:\"tokenOut\""
			Fee               *big.Int       "json:\"fee\""
			Recipient         common.Address "json:\"recipient\""
			Deadline          *big.Int       "json:\"deadline\""
			AmountOut         *big.Int       "json:\"amountOut\""
			AmountInMaximum   *big.Int       "json:\"amountInMaximum\""
			SqrtPriceLimitX96 *big.Int       "json:\"sqrtPriceLimitX96\""
		})
		Q.Type = Method.Name
		Q.Amount = toFloat(Params.AmountOut, 18)
		Q.MinMax = toFloat(Params.AmountInMaximum, 6)
		Q.TokenIn = Params.TokenIn.String()
		Q.TokenOut = Params.TokenOut.String()
	case "exactInput":
		Q.Protocol = "Uniswap V3"
		Params := Input[0].(struct {
			Path             []uint8        "json:\"path\""
			Recipient        common.Address "json:\"recipient\""
			Deadline         *big.Int       "json:\"deadline\""
			AmountIn         *big.Int       "json:\"amountIn\""
			AmountOutMinimum *big.Int       "json:\"amountOutMinimum\""
		})
		Q.Type = Method.Name
		Q.Amount = toFloat(Params.AmountIn, 18)
		Q.MinMax = toFloat(Params.AmountOutMinimum, 6)
		Q.TokenIn = common.HexToAddress(hexutil.Encode(Params.Path[:20])).String()
		Q.TokenOut = common.HexToAddress(hexutil.Encode(Params.Path[len(Params.Path)-20:])).String()
	case "exactOutput":
		Q.Protocol = "Uniswap V3"
		Params := Input[0].(struct {
			Path            []uint8        "json:\"path\""
			Recipient       common.Address "json:\"recipient\""
			Deadline        *big.Int       "json:\"deadline\""
			AmountOut       *big.Int       "json:\"amountOut\""
			AmountInMaximum *big.Int       "json:\"amountInMaximum\""
		})
		Q.Type = Method.Name
		Q.Amount = toFloat(Params.AmountOut, 18)
		Q.MinMax = toFloat(Params.AmountInMaximum, 6)
		Q.TokenIn = common.HexToAddress(hexutil.Encode(Params.Path[:20])).String()
		Q.TokenOut = common.HexToAddress(hexutil.Encode(Params.Path[len(Params.Path)-20:])).String()
	default:
		Q.Type = "Unsupported Method"
	}
	return
}
