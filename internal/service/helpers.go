package service

import "crypto/rand"

func GenCode(size int) (string, error) {
	digits := "0123456789"
	code := make([]byte, size)

	_, err := rand.Read(code)
	if err != nil {
		return "", err
	}

	for i := range size {
		code[i] = digits[int(code[i])%len(digits)]
	}

	return string(code), nil
}
