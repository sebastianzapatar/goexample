package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Proveedor struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Nombre    string    `json:"nombre" gorm:"size:150;not null"`
	Email     string    `json:"email" gorm:"size:255;not null;uniqueIndex"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Proveedor) TableName() string { return "proveedores" }

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

type proveedorInput struct {
	Nombre string `json:"nombre" binding:"required,max=150"`
	Email  string `json:"email" binding:"required,email,max=255"`
}

type productoInput struct {
	SKU         string   `json:"sku" binding:"required,max=50"`
	Nombre      string   `json:"nombre" binding:"required,max=150"`
	Precio      *float64 `json:"precio" binding:"required,gte=0"`
	Stock       *int     `json:"stock" binding:"required,gte=0"`
	ProveedorID uint     `json:"proveedor_id" binding:"required,gt=0"`
}

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN es obligatorio")
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
		log.Fatalf("no se pudo conectar a MySQL: %v", err)
	}
	if err := db.AutoMigrate(&Proveedor{}, &Producto{}); err != nil {
		log.Fatalf("no se pudieron crear las tablas: %v", err)
	}
	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "8080"
	}
	log.Fatal(router(db).Run(":" + puerto))
}

func router(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "base de datos no disponible"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"estado": "ok"})
	})
	a := r.Group("/api")
	a.GET("/proveedores", func(c *gin.Context) {
		proveedores := []Proveedor{}
		if err := db.Order("id").Find(&proveedores).Error; err != nil {
			falloDB(c, err)
			return
		}
		c.JSON(http.StatusOK, proveedores)
	})
	a.POST("/proveedores", func(c *gin.Context) {
		var in proveedorInput
		if !leerJSON(c, &in) {
			return
		}
		p := Proveedor{Nombre: strings.TrimSpace(in.Nombre), Email: strings.TrimSpace(in.Email)}
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
	a.GET("/proveedores/:id", func(c *gin.Context) {
		id, ok := idParam(c)
		if !ok {
			return
		}
		var p Proveedor
		if err := db.First(&p, id).Error; err != nil {
			falloDB(c, err)
			return
		}
		c.JSON(http.StatusOK, p)
	})
	a.PUT("/proveedores/:id", func(c *gin.Context) {
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
		var p Proveedor
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
	a.DELETE("/proveedores/:id", func(c *gin.Context) {
		id, ok := idParam(c)
		if !ok {
			return
		}
		result := db.Delete(&Proveedor{}, id)
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
	a.GET("/productos", func(c *gin.Context) {
		q := db.Preload("Proveedor").Order("productos.id")
		if raw := c.Query("proveedor_id"); raw != "" {
			id, err := strconv.ParseUint(raw, 10, 64)
			if err != nil || id == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "proveedor_id inválido"})
				return
			}
			q = q.Where("productos.proveedor_id = ?", id)
		}
		productos := []Producto{}
		if err := q.Find(&productos).Error; err != nil {
			falloDB(c, err)
			return
		}
		c.JSON(http.StatusOK, productos)
	})
	a.POST("/productos", func(c *gin.Context) {
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
	a.GET("/productos/:id", func(c *gin.Context) {
		id, ok := idParam(c)
		if !ok {
			return
		}
		var p Producto
		if err := db.Preload("Proveedor").First(&p, id).Error; err != nil {
			falloDB(c, err)
			return
		}
		c.JSON(http.StatusOK, p)
	})
	a.PUT("/productos/:id", func(c *gin.Context) {
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
		var p Producto
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
	a.DELETE("/productos/:id", func(c *gin.Context) {
		id, ok := idParam(c)
		if !ok {
			return
		}
		result := db.Delete(&Producto{}, id)
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
	// Raw SQL: JOIN, GROUP BY y agregados sobre las dos tablas.
	a.GET("/consultas/resumen-proveedores", func(c *gin.Context) {
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
	a.GET("/consultas/stock-bajo", func(c *gin.Context) {
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
	return r
}

func leerJSON(c *gin.Context, v interface{}) bool {
	if err := c.ShouldBindJSON(v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido o campos requeridos incorrectos", "detalle": err.Error()})
		return false
	}
	return true
}

func idParam(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return 0, false
	}
	return uint(id), true
}

func productoDesdeInput(c *gin.Context, db *gorm.DB, in productoInput) (Producto, bool) {
	in.SKU, in.Nombre = strings.TrimSpace(in.SKU), strings.TrimSpace(in.Nombre)
	if in.SKU == "" || in.Nombre == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sku y nombre no pueden estar vacíos"})
		return Producto{}, false
	}
	var proveedor Proveedor
	if err := db.First(&proveedor, in.ProveedorID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "proveedor_id no existe"})
		} else {
			falloDB(c, err)
		}
		return Producto{}, false
	}
	return Producto{SKU: in.SKU, Nombre: in.Nombre, Precio: *in.Precio, Stock: *in.Stock, ProveedorID: in.ProveedorID}, true
}

func falloDB(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "registro no encontrado"})
	case errors.Is(err, gorm.ErrDuplicatedKey):
		c.JSON(http.StatusConflict, gin.H{"error": "sku o email duplicado"})
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		c.JSON(http.StatusConflict, gin.H{"error": "el proveedor tiene productos asociados"})
	default:
		log.Printf("error de base de datos: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error de base de datos"})
	}
}
