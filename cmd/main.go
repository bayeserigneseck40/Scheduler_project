package main

import (
	"log"
	"scheduler-project/internal/config"
	"scheduler-project/internal/mq"
	"scheduler-project/internal/scheduler"
	"time"
)

func main() {
	log.Println("🚀 Démarrage du Scheduler")

	// Démarrage du scheduler
	go scheduler.StartScheduler(time.Duration(config.SCHEDULE_INTERVAL) * time.Second)

	// Démarrage du consumer pour écouter les événements
	mq.SubscribeEvents()
}
