package controllers

import (
	"net/http"
	"strconv"

	"go-ecommerce/database"
	"go-ecommerce/models"
	"go-ecommerce/services"

	"github.com/gin-gonic/gin"
)

// CartController handles cart-related HTTP requests
type CartController struct {
	productService *services.ProductService
}

// NewCartController creates a new cart controller
func NewCartController() *CartController {
	return &CartController{
		productService: services.NewProductService(),
	}
}

// GetOrCreateCart gets or creates a cart for a user
// @Summary      Get or create a cart for a user
// @Description  Retrieve an existing cart or create a new one for the specified user
// @Tags         carts
// @Accept       json
// @Produce      json
// @Param        user_id  path      int  true  "User ID"
// @Success      200      {object}  models.Cart
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /carts/user/{user_id} [get]
func (c *CartController) GetOrCreateCart(ctx *gin.Context) {
	userID, err := strconv.ParseUint(ctx.Param("user_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var cart models.Cart
	result := database.GetDB().Where("user_id = ?", userID).First(&cart)
	if result.Error != nil {
		// Create new cart
		cart = models.Cart{UserID: uint(userID)}
		if err := database.GetDB().Create(&cart).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Load items
	database.GetDB().Preload("Product").Where("cart_id = ?", cart.ID).Find(&cart.Items)

	ctx.JSON(http.StatusOK, cart)
}

// AddItemToCart handles POST /api/carts/:cart_id/items
// @Summary      Add item to cart
// @Description  Add a product to the shopping cart
// @Tags         carts
// @Accept       json
// @Produce      json
// @Param        cart_id  path      int  true  "Cart ID"
// @Param        item     body      object  true  "Item information"  example({"product_id": 1, "quantity": 2})
// @Success      201      {object}  models.CartItem
// @Success      200      {object}  models.CartItem
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /carts/{cart_id}/items [post]
func (c *CartController) AddItemToCart(ctx *gin.Context) {
	cartID, err := strconv.ParseUint(ctx.Param("cart_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid cart ID"})
		return
	}

	var req struct {
		ProductID uint `json:"product_id" binding:"required"`
		Quantity  int  `json:"quantity" binding:"required,min=1"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if product exists and has stock
	if !c.productService.IsInStock(req.ProductID, req.Quantity) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "insufficient stock"})
		return
	}

	// Check if item already exists in cart
	var existingItem models.CartItem
	result := database.GetDB().Where("cart_id = ? AND product_id = ?", cartID, req.ProductID).First(&existingItem)

	if result.Error == nil {
		// Update quantity
		existingItem.Quantity += req.Quantity
		if err := database.GetDB().Save(&existingItem).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, existingItem)
		return
	}

	// Create new cart item
	cartItem := models.CartItem{
		CartID:    uint(cartID),
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	}

	if err := database.GetDB().Create(&cartItem).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	database.GetDB().Preload("Product").First(&cartItem, cartItem.ID)
	ctx.JSON(http.StatusCreated, cartItem)
}

// RemoveItemFromCart handles DELETE /api/carts/:cart_id/items/:item_id
// @Summary      Remove item from cart
// @Description  Remove an item from the shopping cart
// @Tags         carts
// @Accept       json
// @Produce      json
// @Param        cart_id  path      int  true  "Cart ID"
// @Param        item_id  path      int  true  "Item ID"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /carts/{cart_id}/items/{item_id} [delete]
func (c *CartController) RemoveItemFromCart(ctx *gin.Context) {
	itemID, err := strconv.ParseUint(ctx.Param("item_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}

	if err := database.GetDB().Delete(&models.CartItem{}, itemID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "item removed from cart"})
}

// GetCart handles GET /api/carts/:cart_id
// @Summary      Get cart by ID
// @Description  Retrieve a shopping cart with all its items
// @Tags         carts
// @Accept       json
// @Produce      json
// @Param        cart_id  path      int  true  "Cart ID"
// @Success      200      {object}  models.Cart
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Router       /carts/{cart_id} [get]
func (c *CartController) GetCart(ctx *gin.Context) {
	cartID, err := strconv.ParseUint(ctx.Param("cart_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid cart ID"})
		return
	}

	var cart models.Cart
	if err := database.GetDB().Preload("Items.Product").First(&cart, cartID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "cart not found"})
		return
	}

	ctx.JSON(http.StatusOK, cart)
}
