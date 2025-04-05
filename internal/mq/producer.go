package mq

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"log"
	"scheduler-project/internal/config"
	"scheduler-project/internal/edit"
)

// PublishEvent publie un événement sur NATS
func PublishEvent(event edit.Event) error {
	// Connexion à NATS
	nc, err := nats.Connect(config.NATS_URL)
	if err != nil {
		return err
	}
	defer nc.Close()

	// Convertir l'événement en JSON
	eventData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Publier l'événement sur le sujet NATS
	err = nc.Publish("USERS.events", eventData)
	if err != nil {
		return err
	}

	log.Println("Événement publié :", string(eventData))
	return nil
}
