package handlers

import (
	"net/http"
	"vaultseed-backend/internal/database"
	"vaultseed-backend/internal/models"
	"vaultseed-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateContentHandler 创建加密内容
func CreateContentHandler(c *gin.Context) {
	var req models.CreateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// 从 header 获取用户地址
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Missing authorization header"})
		return
	}

	// 简化处理：假设 token 是 address:nonce 格式
	var userAddress string
	if len(authHeader) > 0 {
		// 实际应用中应该解析 token
		userAddress = authHeader
		if idx := len(userAddress); idx > 42 {
			userAddress = userAddress[:42]
		}
	}

	db := database.GetDB()

	// 验证用户存在
	var user models.User
	if err := db.Where("address = ?", userAddress).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not found"})
		return
	}

	// 生成 nonce
	nonce, err := utils.GenerateNonce()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate nonce"})
		return
	}

	// 创建加密内容记录
	content := models.EncryptedContent{
		UserAddress:   userAddress,
		Title:         req.Title,
		EncryptedData: req.EncryptedData,
		EncryptedKey:  req.EncryptedKey,
		IV:            req.IV,
		Nonce:         nonce,
	}

	if err := db.Create(&content).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to save content"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"id":      content.ID,
	})
}

// ListContentHandler 获取用户的内容列表
func ListContentHandler(c *gin.Context) {
	// 从 header 获取用户地址
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Missing authorization header"})
		return
	}

	var userAddress string
	if len(authHeader) > 0 {
		userAddress = authHeader
		if idx := len(userAddress); idx > 42 {
			userAddress = userAddress[:42]
		}
	}

	db := database.GetDB()

	// 查询用户的内容
	var contents []models.EncryptedContent
	if err := db.Where("user_address = ?", userAddress).Order("created_at DESC").Find(&contents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch content"})
		return
	}

	// 构建响应
	response := make([]models.ContentResponse, len(contents))
	for i, content := range contents {
		response[i] = models.ContentResponse{
			ID:        content.ID,
			Title:     content.Title,
			CreatedAt: content.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"contents": response,
	})
}

// DecryptContentHandler 解密内容
func DecryptContentHandler(c *gin.Context) {
	var req models.DecryptContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// 从 header 获取用户地址
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Missing authorization header"})
		return
	}

	var userAddress string
	if len(authHeader) > 0 {
		userAddress = authHeader
		if idx := len(userAddress); idx > 42 {
			userAddress = userAddress[:42]
		}
	}

	// 验证签名
	expectedMessage := utils.GenerateDecryptMessage(req.ContentID, req.Nonce)
	if !utils.VerifyEthereumSignature(expectedMessage, req.Signature, userAddress) {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid signature"})
		return
	}

	db := database.GetDB()

	// 获取内容
	var content models.EncryptedContent
	if err := db.Where("id = ? AND user_address = ?", req.ContentID, userAddress).First(&content).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Content not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch content"})
		}
		return
	}

	// 验证 nonce（防重放）
	if content.Nonce != req.Nonce {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid nonce"})
		return
	}

	// 生成新的 nonce 并更新
	newNonce, err := utils.GenerateNonce()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate nonce"})
		return
	}

	content.Nonce = newNonce
	db.Save(&content)

	// 返回加密数据（实际解密应该在前端进行）
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"content": models.ContentDetailResponse{
			ID:        content.ID,
			Title:     content.Title,
			Content:   "[ENCRYPTED - DECRYPT ON CLIENT]", // 前端需要解密
			CreatedAt: content.CreatedAt,
		},
		"encrypted_data": content.EncryptedData,
		"encrypted_key":  content.EncryptedKey,
		"iv":             content.IV,
	})
}

// GetContentDetailHandler 获取内容详情（包含 nonce）
func GetContentDetailHandler(c *gin.Context) {
	contentID := c.Param("id")
	if contentID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Content ID is required"})
		return
	}

	// 从 header 获取用户地址
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Missing authorization header"})
		return
	}

	var userAddress string
	if len(authHeader) > 0 {
		userAddress = authHeader
		if idx := len(userAddress); idx > 42 {
			userAddress = userAddress[:42]
		}
	}

	db := database.GetDB()

	// 获取内容
	var content models.EncryptedContent
	if err := db.Where("id = ? AND user_address = ?", contentID, userAddress).First(&content).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Content not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch content"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"content": gin.H{
			"id":         content.ID,
			"title":      content.Title,
			"created_at": content.CreatedAt,
			"nonce":      content.Nonce, // 返回 nonce 用于解密
		},
	})
}
