package main

import (
	"regexp"
	shell "github.com/ipfs/go-ipfs-api"
)

var (
	RE = regexp.MustCompile( 
		"^/ip4/([.0-9]+)/(tcp|udp)/\\d+(/quic-v1|/quic)?/p2p/\\w+$",
	)
	LOCALHOST = "127.0.0.1"
)


type IpfsClientNode struct {
	*shell.Shell
}


func NewClient() IpfsClientNode {
	return IpfsClientNode{
		Shell: shell.NewShell("localhost:5001"),
	}
}


func (ipfs * IpfsClientNode) BootstrapNode() error {
	pi, err := ipfs.ID()
	if err != nil {
		return err
	}

	myaddrs := []string{}

	for _, addr := range pi.Addresses {
		if suitableMultiAddress(addr){
			myaddrs = append(myaddrs, addr)
		}
	}

	return nil
}


// address that does not have the localhost as ip
// aren't address for webtransport or webrtc stuffs
func suitableMultiAddress(maddr string) bool{
	res := RE.FindStringSubmatch(maddr)
	return res != nil && res[1] != LOCALHOST
}

// notify server
