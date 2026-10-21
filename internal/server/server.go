package server

import (
	"github.com/dani-dzotsi/double-entry-ledger/internal/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server struct {
	engine *gin.Engine
	cfg    *config.Config
	
}

func New(cfg *config.Config,db *gorm.DB, registrar ...RouteRegistrar) *Server {
	engine := gin.Default()
	SetupRoutes(engine, db, registrar...)
	return &Server{
		engine: engine,
		cfg:    cfg,
	}
}

func (s *Server) Run() error {
	return s.engine.Run(":" + s.cfg.Port)
}
