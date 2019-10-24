package node

import (
	"fmt"
	"net"
)

type version struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

var nodeAddress string
var knownNodes = []string{"localhost:3000"}

func startServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddress
	ln, err := net.Listen(protocol, nodeAddress)
	der ln.Close()
	bc
}
