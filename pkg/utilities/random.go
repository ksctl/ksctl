package utilities

import (
	"crypto/rand"
	"math/big"
)

func GenRandomString(length int) (string, error) {
	const letters string = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", ksctlErrors.ErrUnknown.Wrap(err)
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}
