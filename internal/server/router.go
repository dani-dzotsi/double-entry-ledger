package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RouteRegistrar interface {
	RegisterRoutes(rg *gin.RouterGroup)
}

func SetupRoutes(r *gin.Engine, db *gorm.DB, registrars ...RouteRegistrar) {
	health := NewHealthHandler(db)
	r.GET("/healthz", Healthz)
	r.GET("/readyz", health.Readyz)

	v1 := r.Group("/v1")

	for _, reg := range registrars {
		reg.RegisterRoutes(v1)
	}
}

// simple handler and service to handle the healthz and readyz endpoints

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Readyz(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "Unhealthy",
			"database": "Unavailable",
		})
		return
	}
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "Unhealthy",
			"database": "Unavailable",
		})
		return
	}

	if err := sqlDB.PingContext(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "Unhealthy",
			"database": "Unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "Healthy",
		"database": "Available",
	})
}

func Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "Healthy",
	})
}
