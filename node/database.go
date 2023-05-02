package node

import (
	"errors"
	"sync"
)

type TokenDecimalsDBStruct struct {
	DatabaseStruct
}

var tokenDecimalsDB = TokenDecimalsDBStruct{
	DatabaseStruct{
		Db:    make(map[string]int),
		Mutex: &sync.RWMutex{},
	},
}

func (ThisDatabase *TokenDecimalsDBStruct) ReadTokenDecimals(address string) (decimals int, Error error) {
	ThisDatabase.Mutex.RLock()
	defer ThisDatabase.Mutex.RUnlock()
	var Exist bool
	decimals, Exist = ThisDatabase.Db[address]
	if !Exist {
		Error = errors.New("Token address not found.")
	}
	return
}

func (ThisDatabase *TokenDecimalsDBStruct) WriteTokenDecimals(address string, decimals int) (Exist bool) {
	ThisDatabase.Mutex.Lock()
	defer ThisDatabase.Mutex.Unlock()
	_, Exist = ThisDatabase.Db[address]
	ThisDatabase.Db[address] = decimals
	return
}

func (ThisDatabase *TokenDecimalsDBStruct) DeleteTokenDecimals(address string) (Done bool) {
	ThisDatabase.Mutex.Lock()
	defer ThisDatabase.Mutex.Unlock()
	_, Done = ThisDatabase.Db[address]
	delete(ThisDatabase.Db, address)

	return
}

func (ThisDatabase *TokenDecimalsDBStruct) DatabaseCopy() (Copy map[string]int) {
	ThisDatabase.Mutex.RLock()
	defer ThisDatabase.Mutex.RUnlock()

	// Copy :
	Copy = make(map[string]int)
	for hash, value := range ThisDatabase.Db {
		Copy[hash] = value
	}
	return
}

func (ThisDatabase *TokenDecimalsDBStruct) Fork() (Fork DatabaseStruct) {
	Fork.Db = make(map[string]int)
	Fork.Mutex = &sync.RWMutex{}

	ThisDatabase.Mutex.RLock()
	defer ThisDatabase.Mutex.RUnlock()
	for a, b := range ThisDatabase.Db {
		Fork.Db[a] = b
	}
	return
}
