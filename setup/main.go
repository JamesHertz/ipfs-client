package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"

	. "github.com/JamesHertz/ipfs-client"
	client "github.com/JamesHertz/ipfs-client/client"
)

func boostrapNodes() []string{
	boot_file := os.Getenv("EXP_BOOT_FILE")

	if boot_file == "" {
		Eprintf("variable EXP_BOOT_FILE not set...\n")
	}

	data, err := os.ReadFile(boot_file)
	if err != nil {
		Eprintf("Error reading EXP_BOOT_FILE (%s): %s\n", boot_file, err)
	}

	var (
		nodes [][]string
		chosen  [] string
	)

	if err := json.Unmarshal(data, &nodes); err != nil {
		Eprintf("Error unmarshalling EXP_BOOT_FILE (%s): %s\n", boot_file, err)
	}

	for _, node := range nodes {
		chosen = append(chosen, node[ rand.Intn(len(node)) ])
	}

	if len(chosen) == 0 {
		Eprintf("Not nodes found in EXP_BOOT_FILE (%s)\n", boot_file)
	}

	return chosen
}

func main(){
	log.SetPrefix("node-setup")
	nodes     := boostrapNodes()

	_, err := client.NewClient(
		client.Bootstrap(nodes...),
	)

	if err != nil {
		log.Fatalf("Error bootstraping node: %s", err)
	}

	log.Printf("Connected to %d nodes", len(nodes))
}
