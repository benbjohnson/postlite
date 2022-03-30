package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/benbjohnson/postlite"
)

// Command line flags
var (
	addr = ":5432"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	flag.StringVar(&addr, "addr", addr, "postgres protocol bind address")
	flag.Parse()

	s := postlite.NewServer()
	s.Addr = addr
	if err := s.Open(); err != nil {
		return err
	}
	defer s.Close()

	log.Printf("listening on %s", s.Addr)

	// Wait on signal before shutting down.
	<-ctx.Done()
	log.Printf("SIGINT received, shutting down")

	// Perform clean shutdown.
	if err := s.Close(); err != nil {
		return err
	}
	log.Printf("postlite shutdown complete")

	return nil
}
