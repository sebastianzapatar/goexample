package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Router(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "base de datos no disponible"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"estado": "ok"})
	})

	api := r.Group("/api")
	registrarProveedores(api, db)
	registrarProductos(api, db)
	registrarConsultas(api, db)
	return r
}
