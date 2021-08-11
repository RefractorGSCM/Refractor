/*
 * This file is part of Refractor.
 *
 * Refractor is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package aeshelper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

func Encrypt(data []byte, key string) ([]byte, error) {
	// Generate a new AES cipher using our key
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	// Use GCM (Galois/Counter Mode) for symetric encryption
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	// Create the nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt the data
	return gcm.Seal(nonce, nonce, data, nil), nil
}

func Decrypt(data []byte, key string) ([]byte, error) {
	// Generate a new AES cipher using our key
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	// Use GCM (Galois/Counter Mode) for symetric encryption
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	// Check nonce
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("data length < nonce")
	}

	// Decrypt data
	nonce, data := data[:nonceSize], data[nonceSize:]
	decrypted, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}
