package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"time"
	"fmt"

	client "github.com/JamesHertz/ipfs-client/client"
	"github.com/JamesHertz/webmaster/record"
)

var (
	NORMAL = "normal"
	SECURE = "secure"
	NONE   = "default"
)

var mode string

func parseMode() record.IpfsMode {
	flag.StringVar(&mode, "mode", NONE, "choose the node mode (used for publish cids on webmaster)")

	switch mode {
	case NONE:
		return record.NONE
	case SECURE:
		return record.SECURE_IPFS
	case NORMAL:
		return record.NORMAL_IPFS
	default:
		fmt.Printf("Invalid mode: \"%s\"\n", mode)
		fmt.Printf("Should've been none or one of: %v", []string{NORMAL, SECURE, NONE})
		os.Exit(1)
	}

	return record.NONE
}

func main(){
	nodeMode := parseMode()

	rand.Seed( time.Now().Unix() )
	log.SetPrefix(mode + "-ipfs-client ")

	log.Print("Running ipfs-client")

	ipfs := client.NewClient(nodeMode)
	if err := ipfs.BootstrapNode() ; err != nil {
		log.Fatalf("Error bootstrap the client: %v", err)
	}

	log.Println("Bootstrap complete. Waiting for one minute")

	time.Sleep(time.Minute * 1)

	if err := ipfs.UploadFiles(); err != nil {
		log.Fatalf("Error uploading files: %v", err)
	}

	log.Println("Uploading complete. Waitint 5 minutes")
	time.Sleep(time.Minute * 5)

	// start experiment
	ipfs.StartExperiment()
}
