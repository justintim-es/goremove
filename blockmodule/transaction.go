package blockmodule

import (
	"aschar/customdata"
	"aschar/wallet"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/mr-tron/base58"
)

type TXOutputs struct {
	Outputs []TXOutput
}

func (outs TXOutputs) Serialize() []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}
	return outputs
}

type UTXOSet struct {
	Blockchain *Blockchain
}

func (u UTXOSet) CountTransactions() int {
	counter := 0
	vaschal, err := customdata.LoadCache()
	if err != nil {
		log.Panic(err)
	}
	for range vaschal {
		counter++
	}
	return counter
}
func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	files, err := customdata.LoadCache()
	if err != nil {
		log.Panic(err)
	}
	for _, file := range files {
		txID := file.Name()
		data, err := customdata.DecodeCache(file.Name())
		if err != nil {
			log.Panic(err)
		}
		outs := DeserializeOutputs(data)
		for outIdx, out := range outs.Outputs {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
			}
		}
	}
	return accumulated, unspentOutputs
}
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	files, err := customdata.LoadCache()
	if err != nil {
		log.Panic(err)
	}
	for _, file := range files {
		vaschal, err := customdata.DecodeCache(file.Name())
		if err != nil {
			log.Panic(err)
		}
		outs := DeserializeOutputs(vaschal)
		for _, out := range outs.Outputs {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs

	// unspentTransactions := bc.FindUnspentTransactions(pubKeyHash)
	// for _, tx := range unspentTransactions {
	// 	for _, out := range tx.Vout {
	// 		if out.IsLockedWithKey(pubKeyHash) {
	// 			UTXOs = append(UTXOs, out)
	// 		}
	// 	}
	// }
	// return UTXOs
}

func (u UTXOSet) Update(block *Block) {
	fmt.Println("GOT IN HERE")
	for _, tx := range block.Transactions {
		fmt.Println("Iscoinbase", tx.IsCoinbase())
		if tx.IsCoinbase() == false {
			for _, vin := range tx.Vin {
				updatedOuts := TXOutputs{}
				outsBytes, err := customdata.DecodeCache(hex.EncodeToString(vin.TxId) + ".data")
				if err != nil {
					log.Panic(err)
				}
				outs := DeserializeOutputs(outsBytes)
				for outIdx, out := range outs.Outputs {
					if outIdx != vin.Vout {
						updatedOuts.Outputs = append(updatedOuts.Outputs, out)
					}
				}
				if len(updatedOuts.Outputs) == 0 {
					fmt.Println("SHOULD REMOVE")
					err := customdata.DeleteExplicitFileCache(hex.EncodeToString(vin.TxId) + ".data")
					if err != nil {
						log.Panic(err)
					}
				} else {
					err := customdata.SaveCache(hex.EncodeToString(vin.TxId), updatedOuts.Serialize())
					if err != nil {
						log.Panic(err)
					}
				}
			}
		}
		newOutputs := TXOutputs{}
		for _, out := range tx.Vout {
			newOutputs.Outputs = append(newOutputs.Outputs, out)
		}
		err := customdata.SaveCache(hex.EncodeToString(tx.ID), newOutputs.Serialize())
		if err != nil {
			log.Panic(err)
		}
	}
}

func (u UTXOSet) Reindex() {
	err := customdata.DeleteFullCache()
	if err != nil {
		log.Panic(err)
	}
	UTXO := u.Blockchain.FindUTXO()
	for txID, outs := range UTXO {
		err := customdata.SaveCache(txID, outs.Serialize())
		if err != nil {
			log.Panic(err)
		}
	}
}

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

func (tx *Transaction) Serialize() []byte {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes()
}
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()
	for inID, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.TxId)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.Signature)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}
	return true
}
func (tx *Transaction) Hash() []byte {
	var hash [32]byte
	txCopy := *tx
	txCopy.ID = []byte{}
	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:]
}
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.TxId, vin.Vout, nil, nil})
	}
	// kijk of je deze onderste forloop weg kan halen want volgens mij is hij niet nodig
	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}
	txCopy := Transaction{tx.ID, inputs, outputs}
	return txCopy
}
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}
	txCopy := tx.TrimmedCopy()
	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.TxId)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[inID].Signature = signature
	}
}
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].TxId) == 0 && tx.Vin[0].Vout == -1
}
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

type TXInput struct {
	TxId      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

func (in *TXInput) UseKey(pubKeyHash []byte) bool {
	lockingHash := wallet.HashPubKey(in.PubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

func (out *TXOutput) Lock(address []byte) {
	pubKeyHash, err := base58.Decode(string(address))
	if err != nil {
		log.Panic(err)
	}
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))
	return txo
}
func NewUTXOTransaction(from, to string, amount int, UTXOSet *UTXOSet) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	wallets, err := wallet.CheckWallets()
	if err != nil {
		log.Panic(err)
	}
	singleWallet := wallets.GetWallet(from)
	pubKeyHash := wallet.HashPubKey(singleWallet.PublicKey)
	acc, validOutputs := UTXOSet.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("ERROR: not enough funds")
	}
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(strings.Replace(txid, ".data", "", 1))
		if err != nil {
			log.Panic(err)
		}
		for _, out := range outs {
			input := TXInput{txID, out, nil, singleWallet.PublicKey}
			inputs = append(inputs, input)
		}
	}
	outputs = append(outputs, *NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}
	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	UTXOSet.Blockchain.SignTransaction(&tx, singleWallet.PrivateKey)
	return &tx

}
func NewCoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s", to)
	}
	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(10, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.SetID()
	return &tx
}
