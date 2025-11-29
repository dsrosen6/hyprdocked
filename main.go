package main

import "github.com/dsrosen6/hyprlaptop/internal/hypr"

func main() {
	panic(run())
}

func run() error {
	conn, err := hypr.NewConn()
	if err != nil {
		return err
	}

	if err := conn.Listen(); err != nil {
		return err
	}

	return nil
}
