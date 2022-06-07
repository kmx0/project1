package crypto

import (
	"crypto/md5"
	"encoding/hex"
)

func CookieHash(ip, userAgent, login string) string {
	hash := md5.Sum([]byte(ip + userAgent + login))
	return hex.EncodeToString(hash[:])
}

// CalculateLuhn return the check number
func CalculateLuhn(number int) bool {
	checkNumber := checksum(number)
	return checkNumber == 0
}

func checksum(number int) int {
	var luhn int

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 { // even
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number = number / 10
	}
	return luhn % 10
}
