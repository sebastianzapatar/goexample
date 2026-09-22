package models

import "time"

type Producto struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	SKU         string    `json:"sku" gorm:"size:50;not null;uniqueIndex"`
	Nombre      string    `json:"nombre" gorm:"size:150;not null"`
	Precio      float64   `json:"precio" gorm:"type:decimal(12,2);not null"`
	Stock       int       `json:"stock" gorm:"not null"`
	ProveedorID uint      `json:"proveedor_id" gorm:"not null;index"`
	Proveedor   Proveedor `json:"proveedor" gorm:"foreignKey:ProveedorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
