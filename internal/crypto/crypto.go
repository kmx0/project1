package crypto

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	"github.com/theplant/luhn"
)

func CookieHash(ip, userAgent, login string) string {
	ips := strings.Split(ip, ":")
	ip = ips[0]
	hash := md5.Sum([]byte(ip + userAgent + login))
	return hex.EncodeToString(hash[:])
}

// CalculateLuhn return the check number
func CalculateLuhn(number int) bool {
	//= true
	return luhn.Valid(number)
}
