package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"productos-api/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type productoInput struct {
	SKU         string   `json:"sku" binding:"required,max=50"`
	Nombre      string   `json:"nombre" binding:"required,max=150"`
	Precio      *float64 `json:"precio" binding:"required,gte=0"`
	Stock       *int     `json:"stock" binding:"required,gte=0"`
	ProveedorID uint     `json:"proveedor_id" binding:"required,gt=0"`
}

func registrarProductos(api *gin.RouterGroup, db *gorm.DB) {
	api.GET("/productos", func(c *gin.Context) {
		q := db.Preload("Proveedor").Order("productos.id")
		if raw := c.Query("proveedor_id"); raw != "" {
			id, err := strconv.ParseUint(raw, 10, 64)
			if err != nil || id == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "proveedor_id inválido"})
				return
			}
			q = q.Where("productos.proveedor_id = ?", id)
		}
		productos := []models.Producto{}
		if err := q.Find(&productos).Error; err != nil {
			falloDB(c, err)
			return
		}
		c.JSON(http.StatusOK, productos)
	})
	api.POST("/productos", func(c *gin.Context) {
		var in productoInput
		if !leerJSON(c, &in) {
			return
		}
		p, ok := productoDesdeInput(c, db, in)
		if !ok {
			return
		}
		if err := db.Create(&p).Error; err != nil {
			falloDB(c, err)
			return
		}
		if err := db.Preload("Proveedor").First(&p, p.ID).Error; err != nil {
			falloDB(c, err)
			return
		}
		c.JSON(http.StatusCreated, p)
	})
	api.GET("/productos/:id", func(c *gin.Context) {
		id, ok := idParam(c)
		if !ok {
			return
		}
		var p models.Producto
		if err := db.Preload("Proveedor").First(&p, id).Error; err != nil {
			falloDB(c, err)
			return
		}
		c.JSON(http.StatusOK, p)
	})
	api.PUT("/productos/:id", func(c *gin.Context) {
		id, ok := idParam(c)
		if !ok {
			return
		}
		var in productoInput
		if !leerJSON(c, &in) {
			return
		}
		nuevo, ok := productoDesdeInput(c, db, in)
		if !ok {
			return
		}
		var p models.Producto
		if err := db.First(&p, id).Error; err != nil {
			falloDB(c, err)
			return
		}
		if err := db.Model(&p).Updates(map[string]interface{}{"sku": nuevo.SKU, "nombre": nuevo.Nombre, "precio": nuevo.Precio, "stock": nuevo.Stock, "proveedor_id": nuevo.ProveedorID}).Error; err != nil {
			falloDB(c, err)
			return
		}
		if err := db.Preload("Proveedor").First(&p, id).Error; err != nil {
			falloDB(c, err)
			return
		}
		c.JSON(http.StatusOK, p)
	})
	api.DELETE("/productos/:id", func(c *gin.Context) {
		id, ok := idParam(c)
		if !ok {
			return
		}
		result := db.Delete(&models.Producto{}, id)
		if result.Error != nil {
			falloDB(c, result.Error)
			return
		}
		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "registro no encontrado"})
			return
		}
		c.Status(http.StatusNoContent)
	})
}

func productoDesdeInput(c *gin.Context, db *gorm.DB, in productoInput) (models.Producto, bool) {
	in.SKU, in.Nombre = strings.TrimSpace(in.SKU), strings.TrimSpace(in.Nombre)
	if in.SKU == "" || in.Nombre == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sku y nombre no pueden estar vacíos"})
		return models.Producto{}, false
	}
	var proveedor models.Proveedor
	if err := db.First(&proveedor, in.ProveedorID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "proveedor_id no existe"})
		} else {
			falloDB(c, err)
		}
		return models.Producto{}, false
	}
	return models.Producto{SKU: in.SKU, Nombre: in.Nombre, Precio: *in.Precio, Stock: *in.Stock, ProveedorID: in.ProveedorID}, true
}
