package main

import (
	"log"
	"os"
	"time"
)

const (
	envWebURL = "WEB_URL"
)

func main() {
	retryTime := time.Second
	input := make(chan job, 64)
	defer close(input)
	output := make(chan Model, 64)
	defer close(output)

	// start run loop
	go runLoop(input, output)
	go runLoop(input, output)

	for {
		j, err := dialWS(os.Getenv(envWebURL))
		if err != nil {
			log.Println("ws:", err)
			retryTime += time.Second
			time.Sleep(retryTime)
			continue
		}
	loop:
		for {
			select {
			case <-j.disconnet:
				log.Println("ws disconneted")
				break loop

			case s := <-j.submit:
				input <- s

			case o := <-output:
				j.update <- o
			}
		}
	}
}