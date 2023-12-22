package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/JamesHertz/ipfs-client/client"
	"github.com/JamesHertz/ipfs-client/experiments"
	. "github.com/JamesHertz/ipfs-client/utils"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

func eprintf(format string, args ...any){
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func loadConfigs() (*NodeConfig, error) {
	var (
		k = koanf.New(".")
		cfg = &NodeConfig{}
	)

	err := k.Load(env.Provider("", ".", func(str string) string {
		return str
	}), nil)

	if err != nil {
		return nil, fmt.Errorf("Error loading configs: %s", err)
	}

	if err := k.Unmarshal("", cfg); err != nil {
		return nil, fmt.Errorf("Error unmarshalling configs: %s", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("Error validating configs: %s", err)
	}

	// TODO: override with the cmd arguments
	return cfg, nil
}


func saveBootstrapAddress(addrInfo *peer.AddrInfo, filename string) error {
	builder := strings.Builder{}
	for _, addr := range addrInfo.Addrs {
		builder.WriteString(
			fmt.Sprintf("%s/p2p/%s\n", addr, addrInfo.ID.Pretty()),
		)
	}

	err := os.WriteFile(filename, []byte(builder.String()), 0666)
	if err != nil {
		return fmt.Errorf("Error writing address to file: %v", err)
	}

	return nil
}

func main() {

	var (
		ipfs *client.IpfsClientNode
		cfg *NodeConfig
		err error
	)

	cfg, err = loadConfigs()
	if err != nil {
		eprintf("Error loading configs: %v\n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ExpDuration)
	defer cancel()

	log.SetPrefix(fmt.Sprintf(
		"%s-ipfs-client[%s]", cfg.Mode, cfg.Role, 
	))

	log.Print("Running ipfs-client")

	if cfg.IsBootstrap() {
		ipfs, err = client.NewClient()
		if err != nil {
			log.Fatalf("Error initializing client: %v", err)
		}

		addrInfo, err := ipfs.SuitableAddresses()
		if err != nil {
			log.Fatalf("Error getting address: %v", err)
		}

		filename := fmt.Sprintf("%s/%s", cfg.BootDirectory, addrInfo.ID.Pretty())
		if err := saveBootstrapAddress(addrInfo, filename); err != nil {
			log.Fatalf("Error saving bootstrap address to %s: %v", filename, err)
		}

		log.Printf("Address saved with success to %s", filename)
	} else {
		// TODO: wait a bit + a random time c:
		bootstraps, err := cfg.LoadBootstraps()
		if err != nil {
			log.Fatalf("Error loading bootstraps: %v", err)
		}

		ipfs, err = client.NewClient(
			client.Bootstrap(bootstraps...),
		)

		log.Printf("Connected to %d nodes", len(bootstraps))
	}

	exp, err := experiments.NewResolveExperiment(cfg)
	if err != nil {
		log.Fatalf("Error creating experiment: %v", err)
	}

	// TODO: wait enough to start at same time as the others c:
	if err := exp.Start(ipfs, ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Done...")
}



