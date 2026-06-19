package main

import (
	"encoding/json"
	"log"
	"os"

	"mercadia.dev/pos/services/central-backend/internal/api"
)

func main() {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(api.OpenAPI()); err != nil {
		log.Fatalf("export Central Backend OpenAPI: %v", err)
	}
}
