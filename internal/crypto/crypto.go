package crypto

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
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
func CalculateLuhn(number string) bool {
	//= true
	numberInt, _ := strconv.Atoi(number)
	return luhn.Valid(numberInt)
}
