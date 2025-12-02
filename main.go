package main

import (
	"fmt"
	"os"

	"github.com/dsrosen6/hyprlaptop/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
