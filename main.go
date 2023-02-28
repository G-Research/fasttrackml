package main

import (
	"github.com/G-Resarch/fasttrack/cmd"

	log "github.com/sirupsen/logrus"
)

func main() {
	err := cmd.RootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
