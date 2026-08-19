package controllers

import (
	"net/http"
	"strconv"

	"go-ecommerce/models"
	"go-ecommerce/services"

	"github.com/gin-gonic/gin"
)

// ProductController handles product-related HTTP requests
type ProductController struct {
	service *services.ProductService
}

// NewProductController creates a new product controller
func NewProductController() *ProductController {
	return &ProductController{
		service: services.NewProductService(),
	}
}

// CreateProduct handles POST /api/products
// @Summary      Create a new product
// @Description  Create a new product in the catalog
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        product  body      models.Product  true  "Product information"
// @Success      201      {object}  models.Product
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /products [post]
func (c *ProductController) CreateProduct(ctx *gin.Context) {
	var product models.Product
	if err := ctx.ShouldBindJSON(&product); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.CreateProduct(&product); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format to ensure proper JSON serialization
	response := map[string]interface{}{
		"id":          product.ID,
		"created_at":  product.CreatedAt,
		"updated_at":  product.UpdatedAt,
		"name":        product.Name,
		"description": product.Description,
		"price":       product.Price,
		"stock":       product.Stock,
		"category":    product.Category,
	}
	if !product.DeletedAt.Valid {
		response["deleted_at"] = nil
	}

	ctx.JSON(http.StatusCreated, response)
}

// GetProduct handles GET /api/products/:id
// @Summary      Get a product by ID
// @Description  Get detailed information about a specific product
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Success      200  {object}  models.Product
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /products/{id} [get]
func (c *ProductController) GetProduct(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	product, err := c.service.GetProduct(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	// Convert to response format to ensure proper JSON serialization
	response := map[string]interface{}{
		"id":          product.ID,
		"created_at":  product.CreatedAt,
		"updated_at":  product.UpdatedAt,
		"name":        product.Name,
		"description": product.Description,
		"price":       product.Price,
		"stock":       product.Stock,
		"category":    product.Category,
	}
	if !product.DeletedAt.Valid {
		response["deleted_at"] = nil
	}

	ctx.JSON(http.StatusOK, response)
}

// GetAllProducts handles GET /api/products
// @Summary      Get all products
// @Description  Retrieve a list of all products in the catalog
// @Tags         products
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.Product
// @Failure      500  {object}  map[string]string
// @Router       /products [get]
func (c *ProductController) GetAllProducts(ctx *gin.Context) {
	products, err := c.service.GetAllProducts()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format to ensure proper JSON serialization
	response := make([]map[string]interface{}, len(products))
	for i, p := range products {
		response[i] = map[string]interface{}{
			"id":          p.ID,
			"created_at":  p.CreatedAt,
			"updated_at":  p.UpdatedAt,
			"name":        p.Name,
			"description": p.Description,
			"price":       p.Price,
			"stock":       p.Stock,
			"category":    p.Category,
		}
		if !p.DeletedAt.Valid {
			response[i]["deleted_at"] = nil
		}
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateProduct handles PUT /api/products/:id
// @Summary      Update a product
// @Description  Update product information
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id       path      int             true  "Product ID"
// @Param        product  body      models.Product  true  "Updated product information"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /products/{id} [put]
func (c *ProductController) UpdateProduct(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	var product models.Product
	if err := ctx.ShouldBindJSON(&product); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.UpdateProduct(uint(id), &product); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "product updated successfully"})
}

// DeleteProduct handles DELETE /api/products/:id
// @Summary      Delete a product
// @Description  Delete a product from the catalog
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /products/{id} [delete]
func (c *ProductController) DeleteProduct(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	if err := c.service.DeleteProduct(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "product deleted successfully"})
}
