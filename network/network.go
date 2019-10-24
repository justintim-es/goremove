package network

import (
	"aschar/blockmodule"
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
)

type version struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

var nodeAddress string
var knownNodes = []string{"localhost:3000"}
var miningAddress string

func startServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddress
	ln, err := net.Listen(protocol, nodeAddress)
	ln.Close()
	bc := blockmodule.OpenBlockchain()
	if nodeAddress != knownNodes[0] {
		sendVersion(knownNodes[0], bc)
	}
	for {
		conn, err := ln.Accept()
		go handleConnection(conn, bc)
	}
}
func sendVersion(addr string, bc *blockmodule.Blockchain) {
	bestheight := bc.GetBestHeight()
	payload := gobEncode(version{nodeVersion, bestHeight, nodeAddress})
	request := append(commandToBytes("version"), payload...)
	sendData(addr, request)
}
func commandToBytes(command string) []byte {
	var bytes [commandLength]byte
	for i, c := range command {
		bytes[i] = byte(c)
	}
	return bytes[:]
}
func bytesToCommand(bytes []byte) string {
	var command []byte
	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return fmt.Sprintf("%s", command)
}
func handleConnection(conn net.Conn, bc *blockmodule.Blockchain) {
	request, err := ioutil.ReadAll(conn)
	command := bytesToCommand(request[:commandLenght])
	fmt.Printf("Recieved %s command\n", command)
	switch command {
	case version:
		handleVersion(request, bc)
	default:
		fmt.Println("Unknown command")
	}
	conn.Close()
}
func handleVersion(request []byte, bc *blockmodule.Blockchain) {
	var buff bytes.Buffer
	var payload version

	buff.Write(request[commandLength:])
	dec := gob.NewEncoder(&buff)
	err := dec.Decode(&payload)
	myBestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight
	if myBestHeight < foreignerBestHeight {
		sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, bc)
	}
	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}
}

type getBlocks struct {
	AddrFrom string
}

func handleGetBlocks(request []byte, bc *Blockchain) {
	blocks := bc.GetBlockHashes()
	sendInv(payload.AddrFromm, "block", blocks)
}

type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

func handleInv(request []byte, bc *Blockchain) {
	fmt.Printf("Recieved inventory with %d %s\n", len(payload.Items), payload.Type)
	if payload.Type == "block" {
		blockInTransit = payload.Items
		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, "block", blockHash)
		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blockInTransit = newInTransit
	}
	if payload.Type == "tx" {
		txID := payload.Items[0]
		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

type getdata struct {
	AddrFrom string
	Type     string
	ID       []byte
}

func handleGetData(request []byte, bc *Blockchain) {
	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		sendBlock(payload.AddrFrom, &block)
	}
	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]
		sendTx(payload.AddrFrom, &tx)
	}
}

type block struct {
	AddrFrom string
	Block    []byte
}

func handleBlock(request []byte, bc *blockmodule.Blockchain) {
	blockData := payload.Block
	block := DeserializeBlock(blockData)
	fmt.Println("Recieved a new block!")
	bc.AddBlock(block)
	fmt.Printf("Added Block %x\n", block.Hash)
	if len(blockInTransit > 0) {
		blockHash := blockInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockHash)
		blockInTransit = blockInTransit[1:]
	} else {
		UTXOSet := UTXOSet{bc}
		UTCOSet.Reindex()
	}
}

type tx struct {
	AddFrom     string
	Transaction []byte
}

func handleTx(request []byte, bc *blockmodule.Blockchain) {
	txData := payload.Transaction
	tx := DeserializeTransaction(txData)
	mempool[hex.EncodeToString(tx.ID)] = tx
	if nodeAddress == knownNodes[0] {
		for _, node := range knownNodes {
			if node != nodeAddress && node != payload.AddrFrom {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		if len(mempool) >= 2 && len(miningAddress) > 0 {
		MineTransactions:
			var txs []*Transaction
			for id := range mempool {
				tx := mempool[id]
				if bc.VerifyTransaction(&tx) {
					txs = append(txs, &tx)
				}
			}
			if len(txs) == 0 {
				fnt.Println("All transactions are invalid! Waiting for new ones")
				return
			}
			cbTx := NewCoinBaseTX(miningAddress, "")
			txs = append(txs, cbTx)
			newBlock := bc.MineBlock(txs)
			UTXOSet := UTXOSet{bc}
			UTCOSet.Reindex()
			fmt.Prinln("New block is mined!")
			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(mempool, txID)
			}
			for _, node := range knownNodes {
				if node != nodeAddress {
					sendInv(node, "block", [][]byte{newBlock.Hash})
				}
			}
			if len(memorypool) > 0 {
				goto MineTransactions
			}
		}
	}
}
