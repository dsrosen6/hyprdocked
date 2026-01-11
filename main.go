package main

import (
	"fmt"

	"github.com/dsrosen6/hyprdocked/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		fmt.Println("Error:", err)
	}
}
