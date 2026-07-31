// Copyright (c) 2021-present, Scribli


package encryption

import (
	"crypto/rand"

	"golang.org/x/crypto/scrypt"
)

func KDF(password, salt string) (key []byte, err error) {
	key, err = scrypt.Key([]byte(password), []byte(salt), 32768, 8, 1, 32)
	if nil != err {
		return
	}
	return
}

func getRandomData(size int) (ret []byte, err error) {
	ret = make([]byte, size)
	_, err = rand.Read(ret)
	return
}
