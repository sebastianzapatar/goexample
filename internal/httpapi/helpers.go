package httpapi

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
