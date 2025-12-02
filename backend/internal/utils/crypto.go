package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// VerifyEthereumSignature 验证以太坊签名
func VerifyEthereumSignature(message, signature, expectedAddress string) bool {
	// 清理消息
	cleanedMessage := strings.TrimSpace(message)
	if len(cleanedMessage) >= 2 && cleanedMessage[0] == '"' && cleanedMessage[len(cleanedMessage)-1] == '"' {
		cleanedMessage = cleanedMessage[1 : len(cleanedMessage)-1]
	}
	cleanedMessage = strings.TrimSpace(cleanedMessage)

	// 确保签名有 0x 前缀
	if !strings.HasPrefix(signature, "0x") {
		signature = "0x" + signature
	}

	// 解码签名
	sigBytes, err := hexutil.Decode(signature)
	if err != nil {
		return false
	}

	if len(sigBytes) != 65 {
		return false
	}

	// 处理 V 值
	adjustedSigBytes := make([]byte, 65)
	copy(adjustedSigBytes, sigBytes)

	v := sigBytes[64]
	if v == 27 || v == 28 {
		adjustedSigBytes[64] = v - 27
	} else if v == 0 || v == 1 {
		// 已经是正确格式
	} else {
		// 尝试调整
		if v >= 35 {
			adjustedSigBytes[64] = v - 35 - 2 // 假设链ID为1
		} else if v >= 27 {
			adjustedSigBytes[64] = v - 27
		} else {
			adjustedSigBytes[64] = 0
		}
	}

	// 使用 Ethereum 标准消息哈希方法
	msgBytes := []byte(cleanedMessage)
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msgBytes), cleanedMessage)
	hash := crypto.Keccak256Hash([]byte(prefix))

	// 从签名恢复公钥
	pubKey, err := crypto.SigToPub(hash.Bytes(), adjustedSigBytes)
	if err != nil {
		return false
	}

	// 从公钥生成地址
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	// 使用字符串比较，确保大小写不敏感
	return strings.ToLower(recoveredAddr.Hex()) == strings.ToLower(expectedAddress)
}

// GenerateNonce 生成随机 nonce
func GenerateNonce() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateMessageForSigning 生成用于签名的消息
func GenerateMessageForSigning(address, nonce string) string {
	return fmt.Sprintf("Sign this message to authenticate with VaultSeed. Address: %s, Nonce: %s", address, nonce)
}

// GenerateDecryptMessage 生成用于解密的签名消息
func GenerateDecryptMessage(contentID uint, nonce string) string {
	return fmt.Sprintf("Sign this message to decrypt content. Content ID: %d, Nonce: %s", contentID, nonce)
}
