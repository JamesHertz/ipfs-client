package main

import (
	"math/rand"
	"time"
	"log"

	client "github.com/JamesHertz/ipfs-client/client"
)


func main(){
	rand.Seed( time.Now().Unix() )

	log.Println("Running ipfs-client")

	ipfs := client.NewClient()
	if err := ipfs.BootstrapNode() ; err != nil {
		log.Fatalf("Error bootstrap the client: %v", err)
	}

	
	log.Println("Not bootstraped with success. Waiting 10 minutes...")

	time.Sleep(time.Minute * 10)
}