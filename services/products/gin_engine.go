package products

import "github.com/gin-gonic/gin"

func NewGinEngine() *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())
	return engine
}
