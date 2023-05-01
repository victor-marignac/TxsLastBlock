package main

import (
	"errors"
	"github.com/ethereum/go-ethereum/core/types"
	"sync"
)

type DatabaseStruct struct {
	Db    map[string]*types.Transaction
	Mutex *sync.RWMutex
}

func (ThisDatabase *DatabaseStruct) ReadTokenDecimals(address string) (decimals int, Error error) {
	ThisDatabase.Mutex.RLock()
	defer ThisDatabase.Mutex.RUnlock()
	var Exist bool
	decimals, Exist = ThisDatabase.Db[address]
	if !Exist {
		Error = errors.New("Tx doesnt exist.")
	}
	return
}

func (ThisDatabase *DatabaseStruct) WriteTransaction(Tx *types.Transaction) (Exist bool) {
	ThisDatabase.Mutex.Lock()
	defer ThisDatabase.Mutex.Unlock()
	_, Exist = ThisDatabase.Db[Tx.Hash().String()]
	ThisDatabase.Db[Tx.Hash().String()] = Tx
	return
}

func (ThisDatabase *DatabaseStruct) DeleteTransaction(Hash string) (Done bool) {
	ThisDatabase.Mutex.Lock()
	defer ThisDatabase.Mutex.Unlock()
	_, Done = ThisDatabase.Db[Hash]
	delete(ThisDatabase.Db, Hash)

	return
}

func (ThisDatabase *DatabaseStruct) DatabaseCopy() (Copy map[string]*types.Transaction) {
	ThisDatabase.Mutex.RLock()
	defer ThisDatabase.Mutex.RUnlock()

	// Copy :
	Copy = make(map[string]*types.Transaction)
	for hash, value := range ThisDatabase.Db {
		Copy[hash] = value
	}
	return
}

func (ThisDatabase *DatabaseStruct) Fork() (Fork DatabaseStruct) {
	Fork.Db = make(map[string]*types.Transaction)
	Fork.Mutex = &sync.RWMutex{}

	ThisDatabase.Mutex.RLock()
	defer ThisDatabase.Mutex.RUnlock()
	for a, b := range ThisDatabase.Db {
		Fork.Db[a] = b
	}
	return
}
