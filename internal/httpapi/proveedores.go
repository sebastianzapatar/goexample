package httpapi

import (
	"net/http"
	"strings"

	"productos-api/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type proveedorInput struct {
	Nombre string `json:"nombre" binding:"required,max=150"`
	Email  string `json:"email" binding:"required,email,max=255"`
}

func registrarProveedores(api *gin.RouterGroup, db *gorm.DB) {
	api.GET("/proveedores", func(c *gin.Context) {
		proveedores := []models.Proveedor{}
		if err := db.Order("id").Find(&proveedores).Error; err != nil {
			falloDB(c, err)
			return
		}
		c.JSON(http.StatusOK, proveedores)
	})
	api.POST("/proveedores", func(c *gin.Context) {
		var in proveedorInput
		if !leerJSON(c, &in) {
			return
		}
		p := models.Proveedor{Nombre: strings.TrimSpace(in.Nombre), Email: strings.TrimSpace(in.Email)}
		if p.Nombre == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nombre no puede estar vacío"})
			return
		}
		if err := db.Create(&p).Error; err != nil {
			falloDB(c, err)
			return
		}
		c.JSON(http.StatusCreated, p)
	})
	api.GET("/proveedores/:id", func(c *gin.Context) {
		id, ok := idParam(c)
		if !ok {
			return
		}
		var p models.Proveedor
		if err := db.First(&p, id).Error; err != nil {
			falloDB(c, err)
			return
		}
		c.JSON(http.StatusOK, p)
	})
	api.PUT("/proveedores/:id", func(c *gin.Context) {
		id, ok := idParam(c)
		if !ok {
			return
		}
		var in proveedorInput
		if !leerJSON(c, &in) {
			return
		}
		in.Nombre, in.Email = strings.TrimSpace(in.Nombre), strings.TrimSpace(in.Email)
		if in.Nombre == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nombre no puede estar vacío"})
			return
		}
		var p models.Proveedor
		if err := db.First(&p, id).Error; err != nil {
			falloDB(c, err)
			return
		}
		if err := db.Model(&p).Updates(map[string]interface{}{"nombre": in.Nombre, "email": in.Email}).Error; err != nil {
			falloDB(c, err)
			return
		}
		c.JSON(http.StatusOK, p)
	})
	api.DELETE("/proveedores/:id", func(c *gin.Context) {
		id, ok := idParam(c)
		if !ok {
			return
		}
		result := db.Delete(&models.Proveedor{}, id)
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
