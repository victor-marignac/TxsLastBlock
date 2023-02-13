package node

import "github.com/ethereum/go-ethereum/ethclient"

func Dial(URL string) (Client *ethclient.Client, Error error) {
	Client, Error = ethclient.Dial(URL)
	return
}
