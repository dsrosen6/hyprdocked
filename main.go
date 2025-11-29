package main

import (
	"fmt"
	"os"
)

func main() {
	a, err := newApp()
	if err != nil {
		fmt.Printf("Error: initializing app: %v\n", err)
		os.Exit(1)
	}

	if err := a.run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
