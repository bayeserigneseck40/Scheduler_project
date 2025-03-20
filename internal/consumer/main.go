package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Structure pour stocker les ressources récupérées de l’API Config
type Resource struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	UcaId int    `json:"uca_id"`
}

// Structure pour stocker les événements après parsing
type Event struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Start       string `json:"dtstart"`
	End         string `json:"dtend"`
}

// Fonction pour récupérer les ressources depuis l'API Config
func getResourcesFromConfig(apiURL string) ([]Resource, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des ressources : %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Config a retourné un statut %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erreur de lecture des données : %v", err)
	}

	var resources []Resource
	if err := json.Unmarshal(body, &resources); err != nil {
		return nil, fmt.Errorf("erreur de décodage JSON : %v", err)
	}

	return resources, nil
}

// Fonction pour construire dynamiquement l'URL de l’EDT
func buildEDTURL(resources []Resource) (string, error) {
	var ucaIDs []string
	for _, res := range resources {
		ucaIDs = append(ucaIDs, fmt.Sprintf("%d", res.UcaId))
	}

	if len(ucaIDs) == 0 {
		return "", fmt.Errorf("aucun UcaId trouvé dans les ressources")
	}

	return fmt.Sprintf(
		"https://edt.uca.fr/jsp/custom/modules/plannings/anonymous_cal.jsp?resources=%s&projectId=2&calType=ical&nbWeeks=1&displayConfigId=128",
		strings.Join(ucaIDs, ","),
	), nil
}

// Fonction pour récupérer l’emploi du temps depuis l’EDT
func getScheduleData(edtURL string) ([]byte, error) {
	resp, err := http.Get(edtURL)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération de l'EDT : %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("EDT API a retourné un statut %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// Fonction pour nettoyer les valeurs des événements
func cleanValue(value string) string {
	value = strings.ReplaceAll(value, "\\n", "\n") // Convertir `\n` en vrai retour à la ligne
	value = strings.ReplaceAll(value, "\\,", ",")  // Convertir `\,` en virgule
	return strings.TrimSpace(value)                // Supprimer les espaces inutiles
}

// Fonction pour parser l’ICAL en JSON
func parseICalData(data []byte) ([]Event, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var events []Event
	currentEvent := Event{}
	currentKey := ""
	currentValue := ""
	inEvent := false

	for scanner.Scan() {
		line := scanner.Text()

		// Début d’un événement
		if line == "BEGIN:VEVENT" {
			inEvent = true
			currentEvent = Event{} // Réinitialisation
			continue
		}

		// Fin d’un événement
		if line == "END:VEVENT" {
			inEvent = false
			events = append(events, currentEvent)
			continue
		}

		// Si on est en dehors d'un événement, on ignore la ligne
		if !inEvent {
			continue
		}

		// Gestion des valeurs multi-lignes
		if strings.HasPrefix(line, " ") {
			currentValue += strings.TrimSpace(line)
			continue
		}

		// Si une nouvelle clé commence, on stocke l’ancienne valeur
		if currentKey != "" {
			currentValue = cleanValue(currentValue)
			switch currentKey {
			case "SUMMARY":
				currentEvent.Summary = currentValue
			case "DESCRIPTION":
				currentEvent.Description = currentValue
			case "LOCATION":
				currentEvent.Location = currentValue
			case "DTSTART":
				currentEvent.Start = currentValue
			case "DTEND":
				currentEvent.End = currentValue
			}
		}

		// Extraire la nouvelle clé et sa valeur
		splitted := strings.SplitN(line, ":", 2)
		if len(splitted) < 2 {
			continue
		}
		currentKey = splitted[0]
		currentValue = splitted[1]

		// Gérer les propriétés encodées (ex: `DTSTART;TZID=Europe/Paris:20240214T080000`)
		if strings.Contains(currentKey, ";") {
			parts := strings.SplitN(currentKey, ";", 2)
			currentKey = parts[0] // Garder uniquement "DTSTART" ou "DTEND"
		}
	}

	return events, nil
}

func main() {
	// 1️⃣ Récupérer les ressources depuis l'API Config
	apiConfigURL := "http://localhost:8080/resources" // Remplace par l'URL réelle
	resources, err := getResourcesFromConfig(apiConfigURL)
	if err != nil {
		fmt.Println("Erreur :", err)
		os.Exit(1)
	}

	// 2️⃣ Construire dynamiquement l’URL de l’EDT
	edtURL, err := buildEDTURL(resources)
	if err != nil {
		fmt.Println("Erreur :", err)
		os.Exit(1)
	}
	fmt.Println("URL EDT générée :", edtURL) // Debug

	// 3️⃣ Récupérer l’emploi du temps
	rawData, err := getScheduleData(edtURL)
	if err != nil {
		fmt.Println("Erreur :", err)
		os.Exit(1)
	}

	// 4️⃣ Parser les données ICAL en JSON
	events, err := parseICalData(rawData)
	if err != nil {
		fmt.Println("Erreur :", err)
		os.Exit(1)
	}

	// 5️⃣ Afficher le JSON final
	jsonData, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		fmt.Println("Erreur d'encodage JSON :", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonData))
}
