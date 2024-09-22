package helpers

import (
	"github.com/joho/godotenv"
	"log"
)

func LoadConfigs() {
	err := godotenv.Load("./configs/.env")
	if err != nil {
		log.Printf("Error loading .env file")
	}
}
