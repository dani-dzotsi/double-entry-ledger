package server

import (
	"github.com/gin-gonic/gin"
)

type RouteRegistrar interface{
	RegisterRoutes(rg *gin.RouterGroup)
}

func SetupRoutes(r *gin.Engine, registrars ...RouteRegistrar){
	v1 := r.Group("/api/v1")

	for _, reg := range registrars{
		reg.RegisterRoutes(v1)
	}
}