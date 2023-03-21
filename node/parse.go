package node

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
)

// Fais la routine qui recevra les DecodedTxs pour les traiter (FAIRE LE MEEEV)
type DecodedTxStruct struct {
	Tx      SimpleTxStruct  `json:"Tx basic infos"`
	Receipt TxReceiptStruct `json:"Tx receipt"`
	Events  Event           `json:"Tx Events"`  // Ce qu'il s'est passé, depuis les logs du receipt
	Queries Query           `json:"Tx Queries"` // Ce que la tx a voulu faire, depuis l'input data
}

type Query struct {
	Protocol  string  `json:"Protocol"`  // Uniswap v2
	Contract  string  `json:"Contract"`  // L'adresse du contrat du routeur Uniswap v2
	Type      string  `json:"Type"`      // SwapAmountIn / SwapAmountOut
	TokenIn   string  `json:"TokenIn"`   // Le token in
	TokenOut  string  `json:"TokenOut"`  // Le token out
	AmountIn  float64 `json:"AmountIn"`  // Amount In/Out
	AmountOut float64 `json:"AmountOut"` // Amount In/Out
	MinMax    float64 `json:"MinMax"`    // Montant Max/Min
}

type Event struct {
	Protocol  string  `json:"Protocol"`  // Uniswap v2
	Contract  string  `json:"Contract"`  // L'adresse du contrat du routeur Uniswap v2
	Type      string  `json:"Type"`      // Swap
	Pool      string  `json:"Pool"`      // L'adresse du pool
	TokenIn   string  `json:"TokenIn"`   // Le token in
	TokenOut  string  `json:"TokenOut"`  // Le token out
	AmountIn  float64 `json:"AmountIn"`  // Le montant qui est arrivé dans le pool
	AmountOut float64 `json:"AmountOut"` // Le montant qui est sorti du pool
}

// SimpleTxStruct est une structure d'une transaction
type SimpleTxStruct struct {
	Hash             string                 `json:"hash"`
	Value            float64                `json:"value"`
	From             string                 `json:"from"`
	To               string                 `json:"to"`
	Input            string                 `json:"InputData"`
	DecodedInputData DecodedInputDataStruct `json:"DecodedInputData"`
}

type DecodedInputDataStruct struct {
	Name         string
	ExactIn      bool
	Amount       *big.Int
	MinMaxAmount *big.Int
	TokenIn      string
	TokenOut     string
}

type TxReceiptStruct struct {
	CumulativeGasUsed uint64       `json:"Cumulative Gas Used: "`
	GasUsed           uint64       `json:"Gas Used: "`
	ContractAddress   string       `json:"Contract Address: "`
	Logs              []*types.Log `json:"Logs: "`
	Status            bool         `json:"Status: "`
}

func TxDecoder(TxsFeed chan *types.Transaction, DecodedTxsFeed chan DecodedTxStruct, Client *ethclient.Client) {

	// fonction anonyme dans une routine pour print le json de decodedstruct
	go func() {
		for decodedTx := range DecodedTxsFeed {
			RawJson, err := json.Marshal(decodedTx)
			if err != nil {
				log.Println("Unable to marshal")
				continue
			}
			log.Println("Tx :", string(RawJson))
			println("\n")
		}
	}()

	// Cette boucle va écrire une DecodedTx dans un channel DecodedTxsFeed
	for Tx := range TxsFeed {

		ThisTx := TxToSimpleStruct(Tx)
		ThisReceipt := ParseReceipt(Client, Tx)
		//ThisEvents := ParseEventsFromReceipt(ThisReceipt, "UniswapV2")
		//ThisQueries := ParseQueriesFromDecodedInputData(ThisTx.DecodedInputData)

		DecodedTx := DecodedTxStruct{
			Tx:      ThisTx,
			Receipt: ThisReceipt,
			//Events:  ThisEvents,
			//Queries: ThisQueries,
		}

		// Envoyez la structure DecodedTx au canal DecodedTxsFeed
		DecodedTxsFeed <- DecodedTx
	}
}

// TxToSimpleStruct prend un objet de type Transaction et retourne un objet de type SimpleTx
func TxToSimpleStruct(Tx *types.Transaction) (Simple SimpleTxStruct) {
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
			Simple.DecodedInputData = DecodeInputData("UniswapV2", Tx)

		case "0xE592427A0AEce92De3Edee1F18E0157C05861564":
			Simple.DecodedInputData = DecodeInputData("UniswapV3", Tx)
		default:
			Simple.Input = hexutil.Encode(Tx.Data())
		}

	} else {
		// Si l'adresse est nulle, cela signifie que la transaction crée un contrat
		Simple.To = "Contract creation"
	}

	return
}

func ParseInputDataTransaction(Method abi.Method, Input []interface{}, Tx *types.Transaction) (Result DecodedInputDataStruct) {
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
	/*case "multicall":
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
	}*/

	default:
		Result.Name = "Unsupported Method"
	}
	return
}

func ParseReceipt(client *ethclient.Client, Tx *types.Transaction) (TxReceipt TxReceiptStruct) {

	receipt, err := client.TransactionReceipt(context.Background(), Tx.Hash())
	if err != nil {
		log.Println(err)
	}

	TxReceipt.CumulativeGasUsed = receipt.CumulativeGasUsed
	TxReceipt.GasUsed = receipt.GasUsed
	TxReceipt.ContractAddress = receipt.ContractAddress.Hex()
	TxReceipt.Logs = receipt.Logs
	TxReceipt.Status = receipt.Status == types.ReceiptStatusSuccessful

	return
}

/*func ParseEventsFromReceipt(Receipt TxReceiptStruct, ProtocolType string) (events Event) {

	return events
}

func ParseQueriesFromDecodedInputData(decodedInputData DecodedInputDataStruct) (queries Query) {
	// Extraire les queries à partir du `DecodedInputData`
	// Implémentez la logique en fonction du protocole (Uniswap v2 ou v3) et des données décodées
	return queries
}*/
