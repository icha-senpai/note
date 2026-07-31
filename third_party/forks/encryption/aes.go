// Copyright (c) 2021-present, Scribli


package encryption

import (
	"crypto/aes"
	"crypto/cipher"
)

func AesEncrypt(data, key []byte) (ret []byte, err error) {
	block, err := aes.NewCipher(key)
	if nil != err {
		return
	}

	aesgcm, err := cipher.NewGCM(block)
	if nil != err {
		return
	}

	nonce, err := getRandomData(12)
	if nil != err {
		return
	}

	data = aesgcm.Seal(nil, nonce, data, nil)
	ret = append(nonce, data...)
	return
}

func AesDecrypt(cryptData, key []byte) (ret []byte, err error) {
	block, err := aes.NewCipher(key)
	if nil != err {
		return
	}

	aesgcm, err := cipher.NewGCM(block)
	if nil != err {
		return
	}

	nonce := cryptData[:12]
	ret = cryptData[12:]
	ret, err = aesgcm.Open(nil, nonce, ret, nil)
	return
}
