package node

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"strings"
)

// SimpleTx est une structure d'une transaction
type SimpleTx struct {
	Hash  string  `json:"hash"`
	Value float64 `json:"value"`
	From  string  `json:"from"`
	To    string  `json:"to"`
	Input string  `json:"InputData"`
}

type MinMaxResult struct {
	Name         string
	ExactIn      bool
	Amount       *big.Int
	MinMaxAmount *big.Int
	TokenIn      string
	TokenOut     string
}

// TxToSimple prend un objet de type Transaction et retourne un objet de type SimpleTx
// La fonction effectue les conversions nécessaires pour extraire les informations
//	nécessaires et les stocker dans un format plus simple

func TxToSimple(Tx *types.Transaction) (Simple SimpleTx) {
	// Convertit l'identificateur de transaction en chaîne de caractères
	Simple.Hash = Tx.Hash().String()

	// Convertit la valeur de la transaction en Ether
	Simple.Value = Wei2Float(Tx.Value(), 18)

	// Récupère l'adresse de l'émetteur de la transaction
	From, _ := types.Sender(types.NewLondonSigner(Tx.ChainId()), Tx)
	Simple.From = From.String()

	// Récupère l'adresse du destinataire de la transaction et check s'il s'agit d'un contract Uniswap v2/v3
	if Tx.To() != nil {
		Simple.To = Tx.To().String()

		switch Simple.To {

		case "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D":
			Simple.To = "Uniswap v2 Router Contract"
			Result := GetMinMaxAmount("UniswapV2", Tx)
			Simple.Input += fmt.Sprintf("Method name: %s, Token In: %s, Token Out: %s, Exact In: %t, Amount: %s, Min/Max Amount: %s", Result.Name, Result.TokenIn, Result.TokenOut, Result.ExactIn, Result.Amount.String(), Result.MinMaxAmount.String())

		case "0xE592427A0AEce92De3Edee1F18E0157C05861564":
			Simple.To = "Uniswap v3 Router Contract"
			Result := GetMinMaxAmount("UniswapV3", Tx)
			Simple.Input += fmt.Sprintf("Method name: %s, Token In: %s, Token Out: %s, Exact In: %t, Amount: %s, Min/Max Amount: %s", Result.Name, Result.TokenIn, Result.TokenOut, Result.ExactIn, Result.Amount.String(), Result.MinMaxAmount.String())

		default:
			Simple.Input = hexutil.Encode(Tx.Data())
		}

	} else {
		// Si l'adresse est nulle, cela signifie que la transaction crée un contrat
		Simple.To = "Contract creation"
	}

	return
}

func parseTransaction(Method abi.Method, Input []interface{}, Tx *types.Transaction, ABI abi.ABI) (Result MinMaxResult) {
	switch Method.Name {
	// Uniswap v2
	case "swapExactTokensForTokens":
		Result.Name = Method.Name
		Result.ExactIn = true
		Result.Amount = Input[0].(*big.Int)
		Result.MinMaxAmount = Input[1].(*big.Int)
		Path := Input[2].([]common.Address)
		Result.TokenIn = Path[0].String()
		Result.TokenOut = Path[len(Path)-1].String()
	case "swapTokensForExactTokens":
		Result.Name = Method.Name
		Result.ExactIn = false
		Result.Amount = Input[0].(*big.Int)
		Result.MinMaxAmount = Input[1].(*big.Int)
		Path := Input[2].([]common.Address)
		Result.TokenIn = Path[0].String()
		Result.TokenOut = Path[len(Path)-1].String()
	case "swapExactETHForTokens":
		Result.Name = Method.Name
		Result.ExactIn = true
		Result.Amount = Tx.Value()
		Result.MinMaxAmount = Input[0].(*big.Int)
		Path := Input[1].([]common.Address)
		Result.TokenIn = Path[0].String()
		Result.TokenOut = Path[len(Path)-1].String()
	case "swapTokensForExactETH":
		Result.Name = Method.Name
		Result.ExactIn = false
		Result.Amount = Input[0].(*big.Int)
		Result.MinMaxAmount = Tx.Value()
		Path := Input[1].([]common.Address)
		Result.TokenIn = Path[0].String()
		Result.TokenOut = Path[len(Path)-1].String()
	case "swapExactTokensForETH":
		Result.Name = Method.Name
		Result.ExactIn = true
		Result.Amount = Input[0].(*big.Int)
		Result.MinMaxAmount = Input[1].(*big.Int)
		Path := Input[2].([]common.Address)
		Result.TokenIn = Path[0].String()
		Result.TokenOut = Path[len(Path)-1].String()
	case "swapETHForExactTokens":
		Result.Name = Method.Name
		Result.ExactIn = false
		Result.Amount = Input[0].(*big.Int)
		Result.MinMaxAmount = Tx.Value()
		Path := Input[1].([]common.Address)
		Result.TokenIn = Path[0].String()
		Result.TokenOut = Path[len(Path)-1].String()
	// Uniswap v3
	case "exactInputSingle":
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
		Result.Name = Method.Name
		Result.ExactIn = true
		Result.Amount = Params.AmountIn
		Result.MinMaxAmount = Params.AmountOutMinimum
		Result.TokenIn = Params.TokenIn.String()
		Result.TokenOut = Params.TokenOut.String()
	case "exactOutputSingle":
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
		Result.Name = Method.Name
		Result.ExactIn = false
		Result.Amount = Params.AmountOut
		Result.MinMaxAmount = Params.AmountInMaximum
		Result.TokenIn = Params.TokenIn.String()
		Result.TokenOut = Params.TokenOut.String()
	case "exactInput":
		Params := Input[0].(struct {
			Path             []uint8        "json:\"path\""
			Recipient        common.Address "json:\"recipient\""
			Deadline         *big.Int       "json:\"deadline\""
			AmountIn         *big.Int       "json:\"amountIn\""
			AmountOutMinimum *big.Int       "json:\"amountOutMinimum\""
		})
		Result.Name = Method.Name
		Result.ExactIn = true
		Result.Amount = Params.AmountIn
		Result.MinMaxAmount = Params.AmountOutMinimum
		Result.TokenIn = common.HexToAddress(hexutil.Encode(Params.Path[:20])).String()
		Result.TokenOut = common.HexToAddress(hexutil.Encode(Params.Path[len(Params.Path)-20:])).String()
	case "exactOutput":
		Params := Input[0].(struct {
			Path            []uint8        "json:\"path\""
			Recipient       common.Address "json:\"recipient\""
			Deadline        *big.Int       "json:\"deadline\""
			AmountOut       *big.Int       "json:\"amountOut\""
			AmountInMaximum *big.Int       "json:\"amountInMaximum\""
		})
		Result.Name = Method.Name
		Result.ExactIn = false
		Result.Amount = Params.AmountOut
		Result.MinMaxAmount = Params.AmountInMaximum
		Result.TokenIn = common.HexToAddress(hexutil.Encode(Params.Path[:20])).String()
		Result.TokenOut = common.HexToAddress(hexutil.Encode(Params.Path[len(Params.Path)-20:])).String()
	case "multicall":
		Calls := Input[0].([][]byte)
		var Method abi.Method
		var Args abi.Arguments
		for _, Call := range Calls {
			RawCallMethod := Call[:4]
			RawCallData := Call[4:]
			CallMethod, err := ABI.MethodById(RawCallMethod)
			if err != nil {
				continue
			}
			if strings.Contains(CallMethod.Name, "exact") {
				Method = *CallMethod
				Args = ABI.Methods[Method.Name].Inputs
				Input, _ = Args.Unpack(RawCallData)
				Result = parseTransaction(Method, Input, Tx, ABI)
				break
			}
		}
	}
	return
}
