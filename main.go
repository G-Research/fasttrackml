package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/cmd"
)

func main() {
	err := cmd.RootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
