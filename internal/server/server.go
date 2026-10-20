package server

import (
	"github.com/dani-dzotsi/double-entry-ledger/internal/config"
	"github.com/gin-gonic/gin"
)


type Server struct{
	engine *gin.Engine
	cfg *config.Config
}

func New(cfg *config.Config, registrar ...RouteRegistrar) *Server{
	engine := gin.Default()
	SetupRoutes(engine, registrar...)
	return &Server{
		engine: engine,
		cfg: cfg,
	}
}


func (s *Server) Run() error{
	return s.engine.Run(":" + s.cfg.Port)
}