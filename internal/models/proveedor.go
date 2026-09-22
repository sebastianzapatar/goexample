package models

import "time"

type Proveedor struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Nombre    string    `json:"nombre" gorm:"size:150;not null"`
	Email     string    `json:"email" gorm:"size:255;not null;uniqueIndex"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Proveedor) TableName() string { return "proveedores" }
