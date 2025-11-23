package main

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"
	"website/config"
	"website/database"
	"website/webserver"
)

var debugLogging *bool

func init() {
	debugLogging = flag.Bool("debug", false, "Whether to enable debug logging")
	flag.Parse()
	config.Load("config.yaml")
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	database.Connect(*debugLogging)
	defer database.Close()

	webserver.Start("../webserver")

	fmt.Println("Press Ctrl+C to stop")
	<-ctx.Done()
	fmt.Println("\nShutting down!")
}
