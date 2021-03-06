package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	_ "embed"

	"github.com/quintodown/quintodownbot/internal/app"
)

func main() {
	testBot := flag.Bool("test", false, "Should execute test bot")
	flag.Parse()

	if err := app.InitializeConfiguration(*testBot, envFile, envTestFile); err != nil {
		log.Fatal(err)
	}

	botApp, cleanup, err := app.ProvideApp()
	if err != nil {
		log.Fatal(err)
	}

	if err := botApp.Start(context.Background()); err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		defer close(c)
		<-c
		botApp.Stop()
		cleanup()
	}()

	botApp.Run()
}
