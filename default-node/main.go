package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	client "github.com/JamesHertz/ipfs-client/client"
	"github.com/JamesHertz/ipfs-client/experiments"
	"github.com/JamesHertz/webmaster/record"
)

const (
	NORMAL = "normal"
	SECURE = "secure"
	NONE   = "default"
)

var (
	mode string
) 


// TODO: set this envariables variables:
//       - NODE_SEQ_NUM
//       - EXP_TOTAL_NODES
//       - EXP_CIDS_PER_NODE"
//       - EXP_CIDS_FILE
//       - EXP_BOOT_FILE
//       - EXP_DURATION

func eprintf(format string, args ...any){
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func parseMode() record.IpfsMode {
	flag.StringVar(&mode, "mode", NONE, "choose the node mode (used for publish cids on webmaster)")
	flag.Parse()

	switch mode {
	case NONE:
		return record.NONE
	case SECURE:
		return record.SECURE_IPFS
	case NORMAL:
		return record.NORMAL_IPFS
	default:
		eprintf(
			"Invalid mode: \"%s\"\nShould've been none or one of: %v\n", 
			mode, []string{NORMAL, SECURE, NONE},
		)
	}

	return record.NONE
}

func boostrapNodes() []string{
	boot_file := os.Getenv("EXP_BOOT_FILE")

	if boot_file == "" {
		eprintf("variable EXP_BOOT_FILE not set...\n")
	}

	data, err := os.ReadFile(boot_file)
	if err != nil {
		eprintf("Error reading EXP_BOOT_FILE (%s): %s\n", boot_file, err)
	}

	var (
		nodes [][]string
		chosen  [] string
	)

	if err := json.Unmarshal(data, &nodes); err != nil {
		eprintf("Error unmarshalling EXP_BOOT_FILE (%s): %s\n", boot_file, err)
	}

	for _, node := range nodes {
		chosen = append(chosen, node[ rand.Intn(len(node)) ])
	}

	if len(chosen) == 0 {
		eprintf("Not nodes found in EXP_BOOT_FILE (%s)\n", boot_file)
	}

	return chosen
}

func experimentDuration() time.Duration {
	duration, err := strconv.Atoi( os.Getenv("EXP_DURATION") )

	if err != nil {
		eprintf("Error parsing EXP_DURATION value: %s\n", err)
	}
	return time.Duration(duration) * time.Minute
}

func main() {

	rand.Seed(time.Now().Unix())
	// get infos :)
	nodeMode  := parseMode()
	duration  := experimentDuration()
	nodes     := boostrapNodes()

	ctx, cancel := context.WithTimeout(context.Background(), duration)

	defer cancel()

	log.SetPrefix(mode + "-ipfs-client ")

	log.Print("Running ipfs-client")

	ipfs, err := client.NewClient(
		client.Mode(nodeMode),
		client.Bootstrap(nodes...),
	)

	log.Printf("Connected to %d nodes", len(nodes))

	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	exp, err := experiments.NewResolveExperiment()
	if err != nil {
		log.Fatalf("Error creating experiment: %v", err)
	}

	if err := exp.Start(ipfs, ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Done...")
}



