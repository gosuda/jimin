package passhash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"strings"

	"golang.org/x/crypto/argon2"
)

// argon2idVersion01 = Alg Identifier: ag2id01#
func argon2idVersion01(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, 5, 7168, 1, 32)
}

func NewPassHash(password string) (passhash string) {
	var salt [16]byte
	rand.Read(salt[:])

	hash := argon2idVersion01([]byte(password), salt[:])
	passhash = "ag2id01$" + base64.RawURLEncoding.EncodeToString(salt[:]) + "$" + base64.RawURLEncoding.EncodeToString(hash)
	return
}

func VerifyPassHash(password, passhash string) (ok bool, upgrade_required bool) {
	parts := strings.Split(passhash, "$")
	if len(parts) != 3 {
		return false, false
	}

	switch parts[0] {
	case "ag2id01":
		salt, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return false, false
		}

		raw_hash, err := base64.RawURLEncoding.DecodeString(parts[2])
		if err != nil {
			return false, false
		}

		hash := argon2idVersion01([]byte(password), salt)
		if subtle.ConstantTimeCompare(hash, raw_hash) == 1 {
			return true, false
		}

		return false, false
	}

	return false, true
}
