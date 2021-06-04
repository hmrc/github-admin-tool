package main

import (
	"log"

	"github-admin-tool/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
