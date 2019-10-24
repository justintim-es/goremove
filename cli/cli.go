package cli

import (
	"aschar/blockmodule"
	"aschar/wallet"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/mr-tron/base58"
)

type CLI struct{}

func (cli *CLI) reindexUTXO() {
	bc := blockmodule.OpenBlockchain()
	UTXOSet := blockmodule.UTXOSet{bc}
	UTXOSet.Reindex()
	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! there are %d transactions in the UTXO set.\n", count)
}
func (cli *CLI) listAddresses() {
	wallets, err := wallet.CheckWallets()
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()
	for _, address := range addresses {
		fmt.Println(address)
	}
}
func (cli *CLI) createWallet() {
	wallets, _ := wallet.CheckWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()
	fmt.Printf("Your new address %s\n", address)
}
func (cli *CLI) createBlockchain(address string) {
	bc := blockmodule.CreateBlockchain(address)
	UTXOSet := blockmodule.UTXOSet{bc}
	UTXOSet.Reindex()

	fmt.Println("Done!")
}
func (cli *CLI) send(from, to string, amount int) {
	if !wallet.ValidateAddress(from) {
		log.Panic("ERROR: sender address is not valid")
	}
	if !wallet.ValidateAddress(to) {
		log.Panic("ERROR: reciever address is not valid")
	}
	bc := blockmodule.OpenBlockchain()
	uTXOSet := blockmodule.UTXOSet{bc}
	tx := blockmodule.NewUTXOTransaction(from, to, amount, &uTXOSet)
	cbTx := blockmodule.NewCoinbaseTx(from, "")
	txs := []*blockmodule.Transaction{cbTx, tx}
	newBlock := bc.MineBlock(txs)
	uTXOSet.Update(newBlock)
	fmt.Println("Success!")
}
func (cli *CLI) getBalance(address string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR invalid address")
	}
	bc := blockmodule.OpenBlockchain()
	UTXOSet := blockmodule.UTXOSet{bc}
	balance := 0
	pubKeyHash, err := base58.Decode(address)
	if err != nil {
		log.Panic(err)
	}
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of '%s': %d\n", address, balance)
}
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println(" reindexutxo - Rebuilds the utxo set")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}
func (cli *CLI) Run() {
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockChainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddresses := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "the address to get balance for")
	createBlockchainAddress := createBlockChainCmd.String("address", "", "the address to send genesis reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	switch os.Args[1] {
	case "reindexutxo":
		err := reindexUTXOCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddresses":
		err := listAddresses.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	default:
		cli.printUsage()
		os.Exit(1)
	}
	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}
	if createBlockChainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockChainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}
	if createWalletCmd.Parsed() {
		cli.createWallet()
	}
	if listAddresses.Parsed() {
		cli.listAddresses()
	}
	if printChainCmd.Parsed() {
		cli.printChain()
	}
	if reindexUTXOCmd.Parsed() {
		cli.reindexUTXO()
	}
	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}
func (cli *CLI) printChain() {
	bc := blockmodule.OpenBlockchain()
	bci := bc.Iterator()
	for {
		blockInstance := bci.Next()
		fmt.Printf("Prev hash: %x\n", blockInstance.PrevBlockHash)
		fmt.Printf("Hash: %x\n", blockInstance.Hash)
		pow := blockmodule.NewProofOfWork(blockInstance)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
		if len(blockInstance.PrevBlockHash) == 0 {
			break
		}
	}
}
