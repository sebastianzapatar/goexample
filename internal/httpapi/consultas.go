package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func registrarConsultas(api *gin.RouterGroup, db *gorm.DB) {
	// Raw SQL: JOIN, GROUP BY y agregados sobre las dos tablas.
	api.GET("/consultas/resumen-proveedores", func(c *gin.Context) {
		type fila struct {
			ProveedorID uint    `json:"proveedor_id"`
			Proveedor   string  `json:"proveedor"`
			Productos   int64   `json:"productos"`
			Unidades    int64   `json:"unidades"`
			ValorStock  float64 `json:"valor_stock"`
		}
		filas := []fila{}
		err := db.Raw(`SELECT pr.id AS proveedor_id, pr.nombre AS proveedor,
			COUNT(p.id) AS productos, COALESCE(SUM(p.stock), 0) AS unidades,
			COALESCE(SUM(p.precio * p.stock), 0) AS valor_stock
			FROM proveedores pr LEFT JOIN productos p ON p.proveedor_id = pr.id
			GROUP BY pr.id, pr.nombre ORDER BY pr.id`).Scan(&filas).Error
		if err != nil {
			falloDB(c, err)
			return
		}
		c.JSON(http.StatusOK, filas)
	})
	api.GET("/consultas/stock-bajo", func(c *gin.Context) {
		limite := 5
		if raw := c.Query("limite"); raw != "" {
			var err error
			limite, err = strconv.Atoi(raw)
			if err != nil || limite < 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "limite inválido"})
				return
			}
		}
		type fila struct {
			ID        uint    `json:"id"`
			SKU       string  `json:"sku"`
			Nombre    string  `json:"nombre"`
			Stock     int     `json:"stock"`
			Precio    float64 `json:"precio"`
			Proveedor string  `json:"proveedor"`
		}
		filas := []fila{}
		err := db.Raw(`SELECT p.id, p.sku, p.nombre, p.stock, p.precio,
			pr.nombre AS proveedor FROM productos p
			JOIN proveedores pr ON pr.id = p.proveedor_id
			WHERE p.stock <= ? ORDER BY p.stock, p.id`, limite).Scan(&filas).Error
		if err != nil {
			falloDB(c, err)
			return
		}
		c.JSON(http.StatusOK, filas)
	})
}
