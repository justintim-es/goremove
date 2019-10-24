package blockmodule

import (
	"aschar/customdata"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
)

type BlockChainIterator struct {
	currentHash []byte
}

func (i *BlockChainIterator) Next() *Block {
	var block *Block
	encodedBlock, err := customdata.LoadHex(hex.EncodeToString(i.currentHash))
	if err != nil {
		log.Panic(err)
	}
	block = DeserializeBlock(encodedBlock)
	i.currentHash = block.PrevBlockHash
	return block
}

type Blockchain struct {
	tip []byte
}

func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return UTXO
}
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("Transaction is not found")
}
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.TxId)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	tx.Sign(privKey, prevTXs)
}
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		fmt.Println("vinid 570", vin.TxId)
		prevTX, err := bc.FindTransaction(vin.TxId)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	return tx.Verify(prevTXs)
}

func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.IsLockedWithKey(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.UseKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.TxId)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return unspentTXs
}

func (bc *Blockchain) Iterator() *BlockChainIterator {
	bci := &BlockChainIterator{bc.tip}
	return bci
}
func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	lastHash, err := customdata.LoadHashes()
	if err != nil {
		log.Panic(err)
	}
	for _, tx := range transactions {
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}
	newBlock := NewBlock(transactions, lastHash)
	err = customdata.SaveHex(hex.EncodeToString(newBlock.Hash), newBlock.Serialize())
	if err != nil {
		log.Panic(err)
	}
	err = customdata.SaveHashes(newBlock.Hash)
	if err != nil {
		log.Panic(err)
	}
	return newBlock
}
func OpenBlockchain() *Blockchain {
	tip, err := customdata.LoadHashes()
	if err != nil {
		log.Panic(err)
	}
	if err != nil {
		log.Panic(err)
	}
	bc := Blockchain{tip}
	return &bc
}
func CreateBlockchain(address string) *Blockchain {

	tip, err := customdata.LoadHashes()
	if err != nil {
		cbtx := NewCoinbaseTx(address, "FIRST COINBASE TX")
		genesis := NewGenesisBlock(cbtx)
		err = customdata.SaveHex(hex.EncodeToString(genesis.Hash), genesis.Serialize())
		tip = genesis.Hash
		if err != nil {
			log.Panic(err)
		}
		err = customdata.SaveHashes(genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
	}
	bc := Blockchain{tip}
	return &bc
}
