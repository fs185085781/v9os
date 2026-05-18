package util

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
)

func UUID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}
func UnixSeconds() int64 { //时间戳秒
	return time.Now().Unix()
}
func UnixMilliseconds() int64 { //时间戳毫秒
	return time.Now().UnixMilli()
}
func MD5Lower(text string) string {
	hash := md5.New()
	hash.Write([]byte(text))
	hashBytes := hash.Sum(nil)
	return hex.EncodeToString(hashBytes)
}
