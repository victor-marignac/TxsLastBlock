package main

/*tx1 := "toto"
tx2 := "tata"

// Tu vas lancer plusieurs routines, chacune devra spam lecture/ecriture, on s'en fiche de ce que tu passes
// Tu te prends la tête, passes un random string
// je me prends la tête psq write prend un *types.transaction du coup je voulais récupérer une vraie tx partout
go func() {
	for {
		var Tx *types.Transaction
		Database.WriteTransaction(Tx)
		//ah
		time.Sleep(time.Millisecond * 10)
	}
}()

go func() {
	for {
		_, _ = Database.ReadTransaction(tx1)
		time.Sleep(time.Millisecond * 10)
	}
}()

go func() {
	for {
		Database.DeleteTransaction(tx2)
		time.Sleep(time.Millisecond * 10)
	}
}()

//NewDatabase := Database.Fork()

// Ok niquel
// Maintenant fais une fonction qui "fork" la database, donc qui l'initialise dans un DatabaseStruct avec un mutex, pour créer un autre DatabaseStruct indépendant
}

//go Dumper()
/*
var Mutex = &sync.RWMutex{}

func Dumper() {
	for {
		time.Sleep(time.Second* 30)
		Mutex.Lock()
		//
		Mutex.Unlock()
	}
}

func Dumper2() {
	for {
		time.Sleep(time.Second* 30)
		Mutex.Lock()
		//
		Mutex.Unlock()
	}
}*/

// Tag moi ce que tu n'as pas compris :
//c'est tout bon j'ai juste afk bio

/*RawJson, err := json.Marshal(SimpleTx)
if err != nil {
	log.Println("Unable to marshal", Tx.Hash().String())
	continue
}
txInfoJson, err := json.Marshal(Receipt)
if err != nil {
	log.Println("Unable to marshal", Tx.Hash().String())
	continue
}
log.Println("Tx :", string(RawJson), "Receipt :", string(txInfoJson))
println("\n")


func ProcessDecodedTxs(DecodedTxsFeed chan DecodedTxStruct) {
	for {
		decodedTx := <-DecodedTxsFeed

		fmt.Println("Transaction hash:", decodedTx.Tx.Hash)
		fmt.Println("Gas used:", decodedTx.GasUsed)

		fmt.Println("Queries:")
		for _, query := range decodedTx.Queries {
			fmt.Printf("  Protocol: %s, Contract: %s, Type: %s, TokenIn: %s, TokenOut: %s, In: %f, Out: %f, MinMax: %f\n",
				query.Protocol, query.Contract, query.Type, query.TokenIn, query.TokenOut, query.AmountIn, query.AmountOut, query.MinMax)
		}

		fmt.Println("Events:")
		for _, event := range decodedTx.Events {
			fmt.Printf("  Protocol: %s, Contract: %s, Type: %s, Pool: %s, TokenIn: %s, TokenOut: %s, AmountIn: %f, AmountOut: %f\n",
				event.Protocol, event.Contract, event.Type, event.Pool, event.TokenIn, event.TokenOut, event.AmountIn, event.AmountOut)
		}

		fmt.Println("--------------------------------------------------")
	}
}

// SimpleTxStruct est une structure d'une transaction
type SimpleTxStruct struct {
	Hash             string                 `json:"Hash"`
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

// TxToSimpleStruct prend un objet de type Transaction et retourne un objet de type SimpleTx
func TxToSimpleStruct(Tx *types.Transaction) (Simple SimpleTx, err error) {
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

		var ProtocolType string
		switch Simple.To {
		case config.UniswapV2ContractRouter:
			ProtocolType = "UniswapV2"
		case config.UniswapV3ContractRouter:
			ProtocolType = "Uniswapv3"
		default:
			Simple.Input = hexutil.Encode(Tx.Data())
		}
		if len(ProtocolType) != 0 {
			var DecodedInputData DecodedInputDataStruct
			if DecodedInputData, err = DecodeInputData(ProtocolType, Tx); err == nil {
				Simple.DecodedInputData = DecodedInputData
			}
		}
	} else {
		// Si l'adresse est nulle, cela signifie que la transaction crée un contrat
		Simple.To = "Contract creation"
	}

	return
}

// Obtenir l'ABI (Application Binary Interface) du protocole Uniswap correspondant
ABI, err = abiFromProtocolType(ProtocolType)
if err != nil {
return
}

// Obtenir la méthode et les arguments correspondants à l'identifiant de méthode
Method, Args, err = getMethodAndArgs(ABI, HexMethod)
if err != nil {
return
}

// Décompresser les données de transaction à l'aide des arguments correspondants pour obtenir les paramètres de méthode
Input, err = Args.Unpack(HexData)
if err != nil {
return
}

// Analyser les paramètres de méthode et extraire les informations minimales et maximales pour l'échange
DecodedInputData = ParseInputDataTransaction(*Method, Input, Tx)
return


func (d *DecodedTx) Log() {
    log.Printf("Transaction hash: %s\n", d.Tx.Hash)
    log.Printf("Transaction from: %s\n", d.Tx.From)
    log.Printf("Transaction to: %s\n", d.Tx.To)
    log.Printf("Transaction value: %f\n", d.Tx.Value)

    log.Println("Transaction query:")
    log.Printf("  Protocol: %s\n", d.Query.Protocol)
    log.Printf("  Contract: %s\n", d.Query.Contract)
    log.Printf("  Type: %s\n", d.Query.Type)
    log.Printf("  TokenIn: %s\n", d.Query.TokenIn)
    log.Printf("  TokenOut: %s\n", d.Query.TokenOut)
    log.Printf("  Amount: %s\n", d.Query.Amount.String())
    log.Printf("  MinMax: %s\n", d.Query.MinMax.String())

    log.Println("Transaction events:")
    for _, event := range d.Events {
        log.Printf("  Protocol: %s\n", event.Protocol)
        log.Printf("  Contract: %s\n", event.Contract)
        log.Printf("  Type: %s\n", event.Type)
        log.Printf("  Pool: %s\n", event.Pool)
        log.Printf("  TokenIn: %s\n", event.TokenIn)
        log.Printf("  TokenOut: %s\n", event.TokenOut)
        log.Printf("  AmountIn: %s\n", event.AmountIn.String())
        log.Printf("  AmountOut: %s\n", event.AmountOut.String())
    }
}

func ParseReceipt(receipt *types.Receipt) (LocalTx LocalTx, Events []Event) {
	ABI, err := abiFromProtocolType("UniswapV2")
	if err != nil {
		log.Fatalf("Erreur lors de la création de l'ABI: %v", err)
	}

	LocalTx.GasUsed = receipt.GasUsed
	LocalTx.ContractAddress = receipt.ContractAddress.Hex()
	LocalTx.Logs = receipt.Logs
	LocalTx.Status = receipt.Status == types.ReceiptStatusSuccessful
	LocalTx.Index = receipt.TransactionIndex

	for _, l := range receipt.Logs {
		event := Event{
			Protocol: "Uniswap V2",
			Contract: l.Address.Hex(),
		}

		switch l.Topics[0].Hex() {
		case config.UniswapV2EventSwap:
			err = ABI.UnpackIntoInterface(&event, Method.Name, l.Data)
			if err != nil {
				log.Printf("Erreur lors de l'extraction des données de l'événement Swap: %v", err)
				continue
			}
			Events = append(Events, event)
		/*case config.UniswapV2EventMint:
			err = ABI.UnpackIntoInterface(&event, "Mint", l.Data)
			if err != nil {
				log.Printf("Erreur lors de l'extraction des données de l'événement Mint: %v", err)
				continue
			}
			Events = append(Events, event)
		case config.UniswapV2EventBurn:
			err = ABI.UnpackIntoInterface(&event, "Burn", l.Data)
			if err != nil {
				log.Printf("Erreur lors de l'extraction des données de l'événement Burn: %v", err)
				continue
			}
			Events = append(Events, event)
		case config.UniswapV2EventApproval:
			err = ABI.UnpackIntoInterface(&event, "Approval", l.Data)
			if err != nil {
				log.Printf("Erreur lors de l'extraction des données de l'événement Approval: %v", err)
				continue
			}
			Events = append(Events, event)
default:
//log.Printf("Event with topic %s not supported", l.Topics[0].Hex())
}
}

return LocalTx, Events
}

*/
