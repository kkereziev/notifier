package main

import (
	"log"
	"time"

	"github.com/kkereziev/notifier/v2"
)

func main() {
	c, err := notifier.NewClient(
		&notifier.Config{
			ServerHost:              "0.0.0.0",
			ServerPort:              8000,
			ServerConnectionTimeout: 5,
			Retries:                 3,
			Delay:                   time.Second * 5,
		},
		notifier.DefaultOptions()...,
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := c.SendSlackNotification("Hello!"); err != nil {
		log.Fatal(err)
	}
}
