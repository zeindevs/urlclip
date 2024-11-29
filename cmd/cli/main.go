package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/zeindevs/urlclip"
)

func main() {
	active := true
	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Interrupt, syscall.SIGTERM)

	urlclip.Run(quitCh, &active)
}
