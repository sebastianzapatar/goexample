package main

import (
	"log"
	"os"

	"productos-api/internal/database"
	"productos-api/internal/httpapi"
)

func main() {
	db, err := database.Connect(os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatal(err)
	}

	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "8080"
	}
	log.Fatal(httpapi.Router(db).Run(":" + puerto))
}
