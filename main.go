package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	client "github.com/JamesHertz/ipfs-client/client"
	"github.com/JamesHertz/webmaster/record"
)

const (
	NORMAL = "normal"
	SECURE = "secure"
	NONE   = "default"
)

var (
	mode string
	bootstrap bool
) 


func parseMode() record.IpfsMode {
	flag.StringVar(&mode, "mode", NONE, "choose the node mode (used for publish cids on webmaster)")
	flag.BoolVar(&bootstrap, "init", false, "if node should add cids and peers or not")

	flag.Parse()

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

func main() {
	nodeMode      := parseMode()
	duration, err := strconv.Atoi( os.Getenv("EXP_DURATION") )

	if err != nil {
		fmt.Printf("Error parsing EXP_DURATION value: %s\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(duration) * time.Minute)

	defer cancel()

	rand.Seed(time.Now().Unix())
	log.SetPrefix(mode + "-ipfs-client ")

	log.Print("Running ipfs-client")

	opts := []client.Option{
		client.Mode(nodeMode),
	}

	if bootstrap {
		opts = append(opts, client.Bootstrap())
	}

	ipfs, err := client.NewClient(opts...)

	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	if err := ipfs.RunExperiment(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("Done...")
}
