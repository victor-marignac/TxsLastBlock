package node

import (
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"sync"
)

type DecodedTx struct {
	Tx     LocalTx
	Events []Event   // logs (receipt)
	Query  QueryData // input data (transaction)
}

// types.Transaction + types.Receipt =
type LocalTx struct {
	Hash     string            // transaction
	From     string            // transaction
	To       string            // transaction
	Value    float64           // transaction
	Fees     LocalTxFees       // transaction + receipt
	GasLimit uint64            // transaction
	GasUsed  uint64            // receipt
	Nonce    uint64            // transaction
	Index    uint              // receipt
	Revert   bool              // receipt
	Logs     []*types.Log      // receipt
	RawTx    types.Transaction // = transaction
}

type LocalTxFees struct {
	GasPrice *big.Int
	TipCap   *big.Int
	FeeCap   *big.Int
}

type QueryData struct {
	Protocol string  `json:"Protocol"` // Uniswap v2
	Contract string  `json:"Contract"` // L'adresse du contrat du routeur Uniswap v2
	Type     string  `json:"Type"`     // SwapAmountIn / SwapAmountOut
	TokenIn  string  `json:"TokenIn"`  // Le token in
	TokenOut string  `json:"TokenOut"` // Le token out
	Amount   float64 `json:"Amount"`   // Amount In/Out
	MinMax   float64 `json:"MinMax"`   // Montant Max/Min
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

type DatabaseStruct struct {
	Db    map[string]int
	Mutex *sync.RWMutex
}
