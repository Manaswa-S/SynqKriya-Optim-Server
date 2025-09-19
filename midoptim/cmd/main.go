package main

import (
	"fmt"
	"midoptim/cmd/db"
	"midoptim/cmd/server"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {

	fmt.Println("Starting Server...")

	flowChan := make(chan os.Signal, 1)
	signal.Notify(flowChan, syscall.SIGINT, syscall.SIGTERM)

	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
		return
	}

	dataStore, err := db.NewDataStore()
	if err != nil {
		fmt.Println(err)
		return
	}

	// err = server.InitHTTPServer(dataStore)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	stoppedOptim := false
	optim, err := server.InitOptimServer(dataStore)
	defer func() {
		if !stoppedOptim {
			optim.StopOptim()
		}
	}()

	if err != nil {
		fmt.Println(err)
		return
	}

	<-flowChan

	fmt.Println("\nshutting down optim, please wait for all processes to close gracefully")
	err = optim.StopOptim()
	if err != nil {
		fmt.Println(err)
		return
	}
	stoppedOptim = true

	if err := db.Close(dataStore); err != nil {
		fmt.Println(err)
	}

}
