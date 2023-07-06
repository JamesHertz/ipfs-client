package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/JamesHertz/ipfs-client/client"
)

func main(){
	log.SetPrefix("bootstrap-ipfs-client")

	out_dir := os.Getenv("EXP_BOOT_DIR")
	if out_dir == "" {
		log.Fatal("Variable EXP_BOOT_DIR not set...")
	}


	log.Println("Starting boostrap client...")

	ipfs, err := client.NewClient()

	if err != nil {
		log.Fatalf("Error initializing client: %v", err)
	}

	peer, err := ipfs.SuitableAddresses()
	if err != nil {
		log.Fatalf("Error getting address: %v", err)
	}

	builder := strings.Builder{}

	for _, addr := range peer.Addrs {
		builder.WriteString(
			fmt.Sprintf("%s/p2p/%s\n", addr, peer.ID.Pretty()),
		)
	}

	filename := fmt.Sprintf("%s/%s", out_dir, peer.ID.Pretty())
	err = os.WriteFile(filename, []byte(builder.String()), 0666)
	if err != nil {
		log.Fatalf("Error writing address to file: %v", err)
	}
}