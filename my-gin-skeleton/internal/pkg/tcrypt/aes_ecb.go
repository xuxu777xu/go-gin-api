package tcrypt

import (
	"bytes"
	"crypto/aes"
	"encoding/base64" // 导入 base64 包
	"errors"
	"fmt"
)

// pkcs7Padding 使用PKCS#7规范对数据进行填充
func pkcs7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// pkcs7UnPadding 去除PKCS#7填充
func pkcs7UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	if length == 0 {
		return nil, errors.New("pkcs7UnPadding: empty data")
	}
	// 获取填充的数字（最后一个字节）
	unpadding := int(origData[length-1])
	if unpadding == 0 || unpadding > length || unpadding > aes.BlockSize { // 检查 unpadding > 0
		// 填充数不合法
		return nil, fmt.Errorf("pkcs7UnPadding: invalid padding size %d", unpadding)
	}
	// 校验填充字节是否都一致
	paddingStartIndex := length - unpadding
	if paddingStartIndex < 0 { // 添加检查防止索引越界
	    return nil, errors.New("pkcs7UnPadding: calculated padding start index is negative")
	}
	for i := paddingStartIndex; i < length; i++ {
		if origData[i] != byte(unpadding) {
			return nil, errors.New("pkcs7UnPadding: invalid padding bytes")
		}
	}
	return origData[:paddingStartIndex], nil
}

// AesEcbEncrypt 实现AES ECB模式加密
// plaintextStr: 待加密的明文字符串
// 返回 base64 编码的密文字符串 和 error
func AesEcbEncrypt(plaintextStr string) (string, error) {
	// 使用原始 UTF-8 字符串的前 16 字节作为密钥，以匹配 CryptoJS 的行为
	// !!! 警告：切勿在生产代码中硬编码密钥 !!!
	key := []byte("34e6adf9-979f-4f")[:16] // 使用前 16 字节
	plaintext := []byte(plaintextStr)

	block, err := aes.NewCipher(key)
	if err != nil {
		// 避免包装密钥本身的错误信息暴露
		return "", errors.New("AesEcbEncrypt: failed to create cipher block")
	}

	blockSize := block.BlockSize()
	// 填充明文
	plaintext = pkcs7Padding(plaintext, blockSize)

	ciphertext := make([]byte, len(plaintext))
	// 分组加密
	for bs, be := 0, blockSize; bs < len(plaintext); bs, be = bs+blockSize, be+blockSize {
		block.Encrypt(ciphertext[bs:be], plaintext[bs:be])
	}

	// 返回 Base64 编码的字符串
	base64Encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return base64Encoded, nil
}

// AesEcbDecrypt 实现AES ECB模式解密
// ciphertextBase64: base64 编码的待解密密文字符串
// 返回 解密后的原始字符串 和 error
func AesEcbDecrypt(ciphertextBase64 string) (string, error) {
	// 使用原始 UTF-8 字符串的前 16 字节作为密钥，以匹配 CryptoJS 的行为
    // !!! 警告：切勿在生产代码中硬编码密钥 !!!
	key := []byte("34e6adf9-979f-4f")[:16] // 使用前 16 字节

	// Base64 解码
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", fmt.Errorf("AesEcbDecrypt: base64 decode error: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
        // 避免包装密钥本身的错误信息暴露
		return "", errors.New("AesEcbDecrypt: failed to create cipher block")
	}

	blockSize := block.BlockSize()
	if len(ciphertext)%blockSize != 0 {
		// ECB 模式解密时，密文长度必须是块大小的整数倍
		return "", errors.New("AesEcbDecrypt: decoded ciphertext is not a multiple of the block size")
	}

	plaintext := make([]byte, len(ciphertext))
	// 分组解密
	for bs, be := 0, blockSize; bs < len(ciphertext); bs, be = bs+blockSize, be+blockSize {
		block.Decrypt(plaintext[bs:be], ciphertext[bs:be])
	}

	// 去除填充
	plaintext, err = pkcs7UnPadding(plaintext)
	if err != nil {
		return "", fmt.Errorf("AesEcbDecrypt: unpadding error: %w", err)
	}

	// 返回原始字符串
	return string(plaintext), nil
}