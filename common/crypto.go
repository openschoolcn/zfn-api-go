package common

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"math/big"
)

func EncryptPassword(pwd string, n string, e string) (string, error) {
	message := []byte(pwd)
	rsaN, _ := base64.StdEncoding.DecodeString(n)
	rsaE, _ := base64.StdEncoding.DecodeString(e)
	nBytes := hex.EncodeToString(rsaN)
	eBytes := hex.EncodeToString(rsaE)

	nInt := new(big.Int)
	nInt.SetString(nBytes, 16)

	eInt := new(big.Int)
	eInt.SetString(eBytes, 16)

	pubKey := &rsa.PublicKey{
		N: nInt,
		E: int(eInt.Int64()),
	}

	encryptedPwd, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, message)
	if err != nil {
		return "", err
	}

	encResult := base64.StdEncoding.EncodeToString(encryptedPwd)
	return encResult, nil
}
