// Scribli - Refactor your thinking
// Copyright (c) 2020-present, b3log.org
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/hkdf"
)

var encryptionMagic = [4]byte{'S', 'E', 'N', 'C'}

const (
	EncryptionSpec byte = 1

	encryptionAlgorithmAES256GCM byte = 1
	encryptionEnvelopeHeaderSize      = len(encryptionMagic) + 3 // magic + spec + algorithm + nonce length
)

type Argon2Params struct {
	Memory      uint32 `json:"memory"`
	Iterations  uint32 `json:"iterations"`
	Parallelism uint8  `json:"parallelism"`
	KeyLength   uint32 `json:"keyLength"`
}

func DefaultArgon2Params() Argon2Params {
	return Argon2Params{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 4,
		KeyLength:   32,
	}
}

func ValidateArgon2Params(p Argon2Params) (Argon2Params, error) {
	if p.KeyLength == 0 {
		return DefaultArgon2Params(), nil
	}
	if p.KeyLength != 32 {
		return p, errors.New("Argon2id KeyLength must be 32")
	}
	if p.Memory < 64*1024 {
		return p, errors.New("Argon2id Memory too low (minimum 64 MB)")
	}
	if p.Memory > 256*1024 {
		return p, errors.New("Argon2id Memory too high (maximum 256 MB)")
	}
	if p.Iterations < 3 {
		return p, errors.New("Argon2id Iterations too low (minimum 3)")
	}
	if p.Iterations > 10 {
		return p, errors.New("Argon2id Iterations too high (maximum 10)")
	}
	if p.Parallelism == 0 || p.Parallelism > 16 {
		return p, errors.New("Argon2id Parallelism must be between 1 and 16")
	}
	return p, nil
}

func DeriveKey(password string, salt []byte, p Argon2Params) []byte {
	return argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)
}

func Encrypt(key, plaintext []byte) ([]byte, error) {
	return encryptGCM(key, plaintext, nil, "Encrypt")
}

func Decrypt(key, ciphertext []byte) ([]byte, error) {
	return decryptGCM(key, ciphertext, nil, "Decrypt")
}

func EncryptionNonce(ciphertext []byte) ([]byte, error) {
	if hasEncryptionMagic(ciphertext) {
		if len(ciphertext) < encryptionEnvelopeHeaderSize {
			return nil, errors.New("encrypted envelope too short")
		}
		if ciphertext[len(encryptionMagic)] != EncryptionSpec {
			return nil, errors.New("unsupported encrypted envelope spec")
		}
		if ciphertext[len(encryptionMagic)+1] != encryptionAlgorithmAES256GCM {
			return nil, errors.New("unsupported encrypted envelope algorithm")
		}
		nonceLength := int(ciphertext[len(encryptionMagic)+2])
		if nonceLength == 0 || len(ciphertext) < encryptionEnvelopeHeaderSize+nonceLength {
			return nil, errors.New("invalid encrypted envelope nonce length")
		}
		return append([]byte(nil), ciphertext[encryptionEnvelopeHeaderSize:encryptionEnvelopeHeaderSize+nonceLength]...), nil
	}
	if len(ciphertext) < 12 {
		return nil, errors.New("ciphertext too short to extract nonce")
	}
	return append([]byte(nil), ciphertext[:12]...), nil
}

func DeriveSubKey(dek []byte, purpose string) []byte {

	r := hkdf.New(sha256.New, dek, nil, []byte(purpose))
	out := make([]byte, 32) // AES-256
	if _, err := io.ReadFull(r, out); err != nil {

		panic("hkdf derive failed: " + err.Error())
	}
	return out
}

func EncryptWithAAD(key, plaintext, aad []byte) ([]byte, error) {
	return encryptGCM(key, plaintext, aad, "EncryptWithAAD")
}

func DecryptWithAAD(key, ciphertext, aad []byte) ([]byte, error) {
	return decryptGCM(key, ciphertext, aad, "DecryptWithAAD")
}

func encryptGCM(key, plaintext, aad []byte, operation string) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New(operation + " requires a 32-byte (AES-256) key")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	nonce := make([]byte, nonceSize)
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	envelope := make([]byte, encryptionEnvelopeHeaderSize, encryptionEnvelopeHeaderSize+nonceSize+len(plaintext)+gcm.Overhead())
	copy(envelope, encryptionMagic[:])
	envelope[len(encryptionMagic)] = EncryptionSpec
	envelope[len(encryptionMagic)+1] = encryptionAlgorithmAES256GCM
	envelope[len(encryptionMagic)+2] = byte(nonceSize)
	envelope = append(envelope, nonce...)
	return gcm.Seal(envelope, nonce, plaintext, envelopeAAD(envelope[:encryptionEnvelopeHeaderSize], aad)), nil
}

func decryptGCM(key, ciphertext, aad []byte, operation string) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New(operation + " requires a 32-byte (AES-256) key")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if hasEncryptionMagic(ciphertext) {
		if len(ciphertext) < encryptionEnvelopeHeaderSize {
			return nil, errors.New("encrypted envelope too short")
		}
		if ciphertext[len(encryptionMagic)] != EncryptionSpec {
			return nil, errors.New("unsupported encrypted envelope spec")
		}
		if ciphertext[len(encryptionMagic)+1] != encryptionAlgorithmAES256GCM {
			return nil, errors.New("unsupported encrypted envelope algorithm")
		}
		if int(ciphertext[len(encryptionMagic)+2]) != nonceSize {
			return nil, errors.New("invalid encrypted envelope nonce length")
		}
		if len(ciphertext) < encryptionEnvelopeHeaderSize+nonceSize+gcm.Overhead() {
			return nil, errors.New("encrypted envelope too short")
		}
		nonce := ciphertext[encryptionEnvelopeHeaderSize : encryptionEnvelopeHeaderSize+nonceSize]
		ct := ciphertext[encryptionEnvelopeHeaderSize+nonceSize:]
		return gcm.Open(nil, nonce, ct, envelopeAAD(ciphertext[:encryptionEnvelopeHeaderSize], aad))
	}
	if len(ciphertext) < nonceSize+gcm.Overhead() {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ct, aad)
}

func hasEncryptionMagic(ciphertext []byte) bool {
	return len(ciphertext) >= len(encryptionMagic) && bytes.Equal(ciphertext[:len(encryptionMagic)], encryptionMagic[:])
}

func IsCiphertext(data []byte) bool {
	return hasEncryptionMagic(data)
}

func envelopeAAD(header, aad []byte) []byte {
	ret := make([]byte, 0, len(header)+len(aad))
	ret = append(ret, header...)
	return append(ret, aad...)
}

func GenerateSalt() ([]byte, error) {
	return randomBytes(16)
}

func GenerateDEK() ([]byte, error) {
	return randomBytes(32)
}

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}
