package main

import "aschar/cli"

func main() {
	// bc := blockmodule.OpenBlockchain()
	// uTXOSet := blockmodule.UTXOSet{bc}
	// uTXOSet.Reindex()
	// vaschal, err := customdata.DecodeCache("322689524619366a16fdaa5ef902289fb9064ee182b1a9a604d2866d5c89845a.data")
	// if err != nil {
	// 	log.Panic(err)
	// }
	// outs := blockmodule.DeserializeOutputs(vaschal)
	// for _, out := range outs.Outputs {
	// 	fmt.Println(out)
	// }
	// fmt.Println()
	cli := cli.CLI{}
	cli.Run()
	// cli.send("1sRUh86mArm1X7dnM9mtf4PJC33nFyDwL", "12JUPN9rRtndYDamherF11pM8ytSR3CbAv", 1)
}
