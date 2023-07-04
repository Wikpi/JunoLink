package blockchain

import (
	"fmt"
	"log"

	"github.com/dgraph-io/badger"
)

const (
	dbPath = "./tools/blocks"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func InitBlockChain() *BlockChain {
	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)

	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	if err := db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found")
			genesis := Genesis()
			fmt.Println("Genesis proved")
			if err = txn.Set(genesis.Hash, genesis.Serialize()); err != nil {
				log.Panic(err)
			}

			err = txn.Set([]byte("lh"), genesis.Hash)

			lastHash = genesis.Hash

			return err
		} else {
			item, err := txn.Get([]byte("lh"))
			if err != nil {
				log.Panic(err)
			}

			err = item.Value(func(val []byte) error {
				lastHash = val
				return nil
			})

			return err
		}
	}); err != nil {
		log.Panic(err)
	}

	blockChain := BlockChain{lastHash, db}

	return &blockChain
}

func (chain *BlockChain) AddBlock(data string) {
	var lastHash []byte

	if err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			log.Panic(err)
		}

		err = item.Value(func(val []byte) error {
			lastHash = val

			return nil
		})

		return err

	}); err != nil {
		log.Panic(err)
	}

	newBlock := CreateBlock(data, lastHash)

	if err := chain.Database.Update(func(txn *badger.Txn) error {
		if err := txn.Set(newBlock.Hash, newBlock.Serialize()); err != nil {
			log.Panic(err)
		}

		err := txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash

		return err
	}); err != nil {
		log.Panic(err)
	}
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{chain.LastHash, chain.Database}

	return iter
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block

	if err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		if err != nil {
			log.Panic(err)
		}

		var encodedBlock []byte
		err = item.Value(func(val []byte) error {
			encodedBlock = val

			return nil
		})
		block = Deserialize(encodedBlock)

		return err
	}); err != nil {
		log.Panic(err)
	}

	iter.CurrentHash = block.PrevHash

	return block
}
