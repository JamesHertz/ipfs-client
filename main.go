package main

import (
	// "context"
	"fmt"
	// "log"
	"os"

	. "github.com/JamesHertz/ipfs-client/utils"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"

)

func eprintf(format string, args ...any){
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func loadConfigs() *NodeConfig {
	var (
		k = koanf.New(".")
		cfg = &NodeConfig{}
	)

	err := k.Load(env.Provider("", ".", func(str string) string {
		return str
	}), nil)

	if err != nil {
		eprintf("Error loading configs: %s\n", err)
	}

	if err := k.Unmarshal("", cfg); err != nil {
		eprintf("Error unmarshalling configs: %s\n", err)
	}


	if err := cfg.Validate(); err != nil {
		eprintf("Error validating configs: %s\n", err)
	}

	// TODO: override with the cmd arguments
	return cfg
}

func main() {
	loadConfigs()

	// get infos :)
	// nodeMode  := parseMode()
	// duration  := experimentDuration()
	// nodes     := boostrapNodes()

	// ctx, cancel := context.WithTimeout(context.Background(), duration)

	// defer cancel()

	// log.SetPrefix(mode + "-ipfs-client ")

	// log.Print("Running ipfs-client")

	// ipfs, err := client.NewClient(
	// 	client.Mode(nodeMode),
	// 	client.Bootstrap(nodes...),
	// )

	// log.Printf("Connected to %d nodes", len(nodes))

	// if err != nil {
	// 	log.Fatalf("Error creating client: %v", err)
	// }

	// exp, err := experiments.NewResolveExperiment()
	// if err != nil {
	// 	log.Fatalf("Error creating experiment: %v", err)
	// }

	// if err := exp.Start(ipfs, ctx); err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println("Done...")
}



