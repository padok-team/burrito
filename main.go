/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"
	"os"

	"github.com/padok-team/burrito/cmd"
	"github.com/padok-team/burrito/internal/burrito"
)

func main() {
	app, err := burrito.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %s\n", err)
		os.Exit(1)
	}

	if err := cmd.New(app).Execute(); err != nil {
		os.Exit(1)
	}
}
