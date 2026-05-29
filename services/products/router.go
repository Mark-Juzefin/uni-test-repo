package products

import (
	"uni-test-repo/services/products/internal/product/productcontroller"

	"github.com/gin-gonic/gin"
)

type Router struct {
	product *productcontroller.HTTPHandler
}

func NewRouter(product *productcontroller.HTTPHandler) *Router {
	return &Router{product: product}
}

func (r *Router) SetUp(engine *gin.Engine) {
	engine.POST("/products", r.product.Create)
	engine.GET("/products", r.product.List)
	engine.DELETE("/products/:id", r.product.Delete)
}
