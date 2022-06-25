package main

import (
	"crypto/md5"
	"fmt"
)

func tokHash(token string) string {
	h := md5.Sum([]byte(token + "faster-token-salt")) // #nosec
	return fmt.Sprintf("%x", h[:5])
}

func sessionFileName(token string) string {
	return fmt.Sprintf("bot.%s.session.json", tokHash(token))
}
