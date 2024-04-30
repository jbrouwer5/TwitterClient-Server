package main

import (
	"Twitter/server"
	"encoding/json"
	"os"
	"strconv"
)
func main() {
	// get command line arguments 
	args := os.Args
	
	dec := json.NewDecoder(os.Stdin)
    enc := json.NewEncoder(os.Stdout)

	var numConsumers int 
	mode := "s"

	if len(args) > 1{
		numConsumers, _ = strconv.Atoi(args[1])
		mode = "p"
	}
	
	config := server.Config{Encoder: enc, Decoder: dec, Mode: mode, ConsumersCount: numConsumers}

	server.Run(config) 
}
