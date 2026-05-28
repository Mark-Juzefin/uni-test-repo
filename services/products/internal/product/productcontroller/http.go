package productcontroller

import (
	"errors"
	"net/http"

	"uni-test-repo/services/products/internal/product"

	"github.com/gin-gonic/gin"
)

type HTTPHandler struct {
	service *product.ProductService
}

func NewHTTPHandler(s *product.ProductService) *HTTPHandler {
	return &HTTPHandler{service: s}
}

func (h *HTTPHandler) Create(c *gin.Context) {
	var req product.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	res, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, product.ErrAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, product.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, res)
}
