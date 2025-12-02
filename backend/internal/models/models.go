package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Address   string    `json:"address" gorm:"uniqueIndex;not null"`
	PublicKey string    `json:"public_key" gorm:"type:text;not null"`
	Nonce     string    `json:"nonce" gorm:"not null"` // 用于防重放攻击
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EncryptedContent 加密内容模型
type EncryptedContent struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserAddress   string    `json:"user_address" gorm:"index;not null"`
	Title         string    `json:"title" gorm:"not null"`
	EncryptedData string    `json:"encrypted_data" gorm:"type:text;not null"` // 加密后的正文
	EncryptedKey  string    `json:"encrypted_key" gorm:"type:text;not null"`  // 使用用户公钥加密的对称密钥
	IV            string    `json:"iv" gorm:"type:text;not null"`             // 初始化向量
	Nonce         string    `json:"nonce" gorm:"not null"`                    // 用于解密时的防重放攻击
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Address   string `json:"address" binding:"required"`
	Signature string `json:"signature" binding:"required"`
	Message   string `json:"message" binding:"required"`
	Nonce     string `json:"nonce" binding:"required"`
}

// RegisterPublicKeyRequest 注册公钥请求
type RegisterPublicKeyRequest struct {
	Address   string `json:"address" binding:"required"`
	PublicKey string `json:"public_key" binding:"required"`
	Signature string `json:"signature" binding:"required"`
	Message   string `json:"message" binding:"required"`
}

// CreateContentRequest 创建内容请求
type CreateContentRequest struct {
	Title         string `json:"title" binding:"required,max=100"`
	EncryptedKey  string `json:"encrypted_key" binding:"required"`  // 使用公钥加密的对称密钥
	IV            string `json:"iv" binding:"required"`             // 初始化向量
	EncryptedData string `json:"encrypted_data" binding:"required"` // 加密后的内容
}

// DecryptContentRequest 解密内容请求
type DecryptContentRequest struct {
	ContentID uint   `json:"content_id" binding:"required"`
	Signature string `json:"signature" binding:"required"`
	Message   string `json:"message" binding:"required"`
	Nonce     string `json:"nonce" binding:"required"`
}

// API 响应结构
type LoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	Address string `json:"address"`
	Message string `json:"message,omitempty"`
}

type ContentResponse struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

type ContentDetailResponse struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"` // 解密后的内容
	CreatedAt time.Time `json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
