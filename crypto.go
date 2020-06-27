package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"

	"github.com/boufni95/goutils"
)

func (t *TestInfo) Decrypt(text []byte) ([]byte, error) {
	encData := EncryptedData{}
	err := goutils.JsonBytesToType(text, &encData)
	if err != nil {
		return nil, err
	}
	key, err := base64.StdEncoding.DecodeString(encData.Key)
	if err != nil {
		return nil, err
	}
	aesKey, err := t.RSAPrivKey.Decrypt(rand.Reader, key, nil)
	if err != nil {
		return nil, err
	}
	hexKey, err := hex.DecodeString(string(aesKey))
	block, err := aes.NewCipher(hexKey)
	if err != nil {
		return nil, err
	}
	iv, err := hex.DecodeString(encData.IV)
	if err != nil {
		return nil, err
	}
	cbc := cipher.NewCBCDecrypter(block, iv)
	data, err := base64.StdEncoding.DecodeString(encData.Data)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, len(data))
	cbc.CryptBlocks(ciphertext, data)
	_, err = goutils.JsonBytesToInterface(ciphertext)
	if err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			ciphertext = ciphertext[:e.Offset-1]
			_, err = goutils.JsonBytesToInterface(ciphertext)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}

	}
	return ciphertext, nil

}

func (t *TestInfo) Encrypt(text []byte) (EncryptedData, error) {
	encData := EncryptedData{}
	textPad := PKCS5Padding(text, aes.BlockSize)
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return encData, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return encData, err
	}
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return encData, err
	}
	cbc := cipher.NewCBCEncrypter(block, iv)
	encryptedText := make([]byte, len(textPad))
	cbc.CryptBlocks(encryptedText, textPad)

	hexKey := hex.EncodeToString(key)
	pemData, _ := pem.Decode([]byte(t.APIPubKey))
	if pemData == nil {
		return encData, errors.New("cant decode API pub")
	}
	APIPub, err := x509.ParsePKCS1PublicKey(pemData.Bytes)
	encryptedKey, err := rsa.EncryptPKCS1v15(rand.Reader, APIPub, []byte(hexKey))
	if err != nil {
		return encData, err
	}
	base64Key := base64.StdEncoding.EncodeToString(encryptedKey)
	base64Data := base64.StdEncoding.EncodeToString(encryptedText)
	hexIv := hex.EncodeToString(iv)
	encData.Data = base64Data
	encData.IV = hexIv
	encData.Key = base64Key
	return encData, nil

}

func PKCS5Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func PKCS5UnPadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}
