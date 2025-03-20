package scheduler

import (
	"log"
	"scheduler-project/internal/edit"
	"scheduler-project/internal/mq"
	"time"
)

func StartScheduler(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("🔄 Récupération de l'emploi du temps...")
		events, err := edit.FetchEDT()
		if err != nil {
			log.Println("❌ Erreur lors de la récupération :", err)
			continue
		}

		for _, event := range events {
			err := mq.PublishEvent(event)
			if err != nil {
				log.Println("❌ Erreur d'envoi à NATS :", err)
			}
		}
	}
}
