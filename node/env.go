package node

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"sync"
)

var Client *ethclient.Client
var Sync = &sync.WaitGroup{}
