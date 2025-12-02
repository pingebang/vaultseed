package handlers

import (
	"net/http"
	"vaultseed-backend/internal/database"
	"vaultseed-backend/internal/models"
	"vaultseed-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LoginHandler 处理用户登录
func LoginHandler(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// 验证签名
	if !utils.VerifyEthereumSignature(req.Message, req.Signature, req.Address) {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid signature"})
		return
	}

	db := database.GetDB()

	// 查找或创建用户
	var user models.User
	result := db.Where("address = ?", req.Address).First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		// 新用户，生成 nonce
		nonce, err := utils.GenerateNonce()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate nonce"})
			return
		}

		user = models.User{
			Address: req.Address,
			Nonce:   nonce,
		}
		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create user"})
			return
		}
	} else if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Database error"})
		return
	}

	// 更新 nonce（防重放）
	newNonce, err := utils.GenerateNonce()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate nonce"})
		return
	}

	user.Nonce = newNonce
	db.Save(&user)

	// 生成简单的 token（在实际应用中应该使用 JWT）
	token := req.Address + ":" + newNonce

	c.JSON(http.StatusOK, models.LoginResponse{
		Success: true,
		Token:   token,
		Address: req.Address,
	})
}

// RegisterPublicKeyHandler 处理公钥注册
func RegisterPublicKeyHandler(c *gin.Context) {
	var req models.RegisterPublicKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// 验证签名
	if !utils.VerifyEthereumSignature(req.Message, req.Signature, req.Address) {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid signature"})
		return
	}

	db := database.GetDB()

	// 查找用户
	var user models.User
	result := db.Where("address = ?", req.Address).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "User not found"})
		return
	}

	// 更新公钥
	user.PublicKey = req.PublicKey
	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to save public key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetNonceHandler 获取 nonce
func GetNonceHandler(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Address is required"})
		return
	}

	db := database.GetDB()

	var user models.User
	result := db.Where("address = ?", address).First(&user)

	var nonce string
	if result.Error == gorm.ErrRecordNotFound {
		// 新用户，生成 nonce
		newNonce, err := utils.GenerateNonce()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate nonce"})
			return
		}
		nonce = newNonce
	} else if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Database error"})
		return
	} else {
		nonce = user.Nonce
	}

	c.JSON(http.StatusOK, gin.H{"nonce": nonce})
}
