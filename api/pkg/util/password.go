package util

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

// passwordKey 是加密所用的密钥，请根据实际需求修改
var passwordKey = "123456"

func ToSetPasswordKey(key string) {
	passwordKey = key
}

// EncodePassword 对原始密码进行加密处理
// 它生成一个6字节的随机盐，使用 passwordKey + 盐 作为密钥，通过 HMAC-SHA256 对密码进行哈希
// 返回格式为: base64编码的HMAC哈希 + "." + 16进制编码的盐
func EncodePassword(password string) string {
	// 1. 生成6字节的随机盐
	salt := make([]byte, 6)
	_, err := rand.Read(salt)
	if err != nil {
		return ""
	}
	// 2. 将 passwordKey 与盐拼接作为HMAC的密钥
	secretKey := passwordKey + string(salt)

	// 3. 创建HMAC-SHA256哈希器
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(password))

	// 4. 计算哈希值并进行Base64编码
	hmacResult := mac.Sum(nil)
	encodedHmac := base64.StdEncoding.EncodeToString(hmacResult)

	// 5. 将盐转换为16进制字符串便于存储
	saltHex := fmt.Sprintf("%x", salt)

	// 6. 返回组合后的字符串：哈希值.盐
	return encodedHmac + "." + saltHex
}

// CheckPassword 验证密码是否与加密字符串匹配
// 它从加密字符串中提取盐值，使用相同的密钥和算法对输入密码进行哈希
// 最后比较新生成的加密字符串与原始加密字符串是否一致
func CheckPassword(password, encodedPassword string) bool {
	// 1. 分割加密字符串，获取原始哈希值和盐
	parts := strings.Split(encodedPassword, ".")
	if len(parts) != 2 {
		return false
	}

	originalHmac := parts[0] // Base64编码的原始HMAC
	saltHex := parts[1]      // 16进制编码的盐

	// 2. 将16进制盐转换回字节
	salt, err := hexStringToBytes(saltHex)
	if err != nil {
		return false
	}

	// 3. 使用相同的密钥和盐生成新的HMAC
	secretKey := passwordKey + string(salt)
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(password))
	hmacResult := mac.Sum(nil)

	// 4. 将新生成的HMAC进行Base64编码
	newEncodedHmac := base64.StdEncoding.EncodeToString(hmacResult)

	// 5. 比较新生成的HMAC与原始HMAC是否相同
	return newEncodedHmac == originalHmac
}

// hexStringToBytes 将16进制字符串转换为字节切片
func hexStringToBytes(hexStr string) ([]byte, error) {
	if len(hexStr)%2 != 0 {
		return nil, errors.New("无效的16进制字符串长度")
	}

	bytes := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		var b byte
		_, err := fmt.Sscanf(hexStr[i:i+2], "%02x", &b)
		if err != nil {
			return nil, err
		}
		bytes[i/2] = b
	}
	return bytes, nil
}
