package mq

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"scheduler-project/internal/config"
)

func SubscribeEvents() {
	nc, err := nats.Connect(config.NATS_URL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	_, err = nc.Subscribe("USERS.events", func(m *nats.Msg) {
		fmt.Println("📥 Reçu :", string(m.Data))
	})
	if err != nil {
		log.Fatal(err)
	}

	select {} // On bloque le programme pour écouter en continu
}
