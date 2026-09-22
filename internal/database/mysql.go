package database

import (
	"fmt"
	"log"
	"time"

	"productos-api/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Connect(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DB_DSN es obligatorio")
	}
	var db *gorm.DB
	var err error
	for intento := 0; intento < 30; intento++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{TranslateError: true})
		if err == nil {
			break
		}
		log.Printf("esperando MySQL: %v", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar a MySQL: %w", err)
	}
	if err := db.AutoMigrate(&models.Proveedor{}, &models.Producto{}); err != nil {
		return nil, fmt.Errorf("no se pudieron crear las tablas: %w", err)
	}
	return db, nil
}
