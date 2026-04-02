package tool

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func StrToDate(s string) (*time.Time, error) {
	layout := "2006-01"
	switch strings.Count(s, "-") {
	case 2:
		layout = "2006-01-02"
	case 1:
		layout = "2006-01"
	case 0:
		if len(s) == 4 {
			layout = "2006"
		}
	}

	t, err := time.Parse(layout, s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func GetConfigPath(app string) string {
	execPath, _ := os.Executable()
	basePath := filepath.Dir(execPath)
	return filepath.Join(basePath, "etc", fmt.Sprintf("%s-api.yaml", app))
}

func MustParseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func MustParseInt(s string) int {
	f, _ := strconv.Atoi(s)
	return f
}

func MustParseUint(s string) uint {
	u64, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return uint(u64)
}

// GenerateUUID 创建UUID
func GenerateUUID() string {
	uid, _ := uuid.NewRandom()
	var buf [32]byte // 32 位固定长度
	hex.Encode(buf[:], uid[:])
	return string(buf[:])
}

// GetIPFromCtx 从 context 获取 IP
func GetIPFromCtx(ctx context.Context) string {
	if ip, ok := ctx.Value("ip").(string); ok {
		return ip
	}
	return ""
}

// GetTodayExpireAt 获取当天剩余秒数
func GetTodayExpireAt() int {
	now := time.Now()
	// 今天 23:59:59
	expireAt := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	return int(expireAt.Sub(now).Seconds())
}

// Ternary 三元运算符
func Ternary[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

func GetUid(ctx context.Context, opt ...bool) string {
	isPanic := true
	if len(opt) > 0 {
		isPanic = opt[0]
	}
	uid, ok := ctx.Value("Uid").(string)
	if (!ok || uid == "") && isPanic {
		panic(any("not logged in, please log in first")) // 或者返回错误
	}
	return uid
}

func GetContextStringValue(ctx context.Context, value string, opt ...bool) string {
	isPanic := true
	if len(opt) > 0 {
		isPanic = opt[0]
	}
	val, ok := ctx.Value(value).(string)
	if (!ok || val == "") && isPanic {
		panic(any("context value not set or empty")) // 或者返回错误
	}
	return val
}

func GetContextIntValue(ctx context.Context, key string, opt ...bool) uint64 {
	isPanic := true
	if len(opt) > 0 {
		isPanic = opt[0]
	}

	v := ctx.Value(key)
	if v == nil {
		if isPanic {
			panic("context value not set")
		}
		return 0
	}

	switch val := v.(type) {
	case uint64:
		return val
	case int:
		return uint64(val)
	case int64:
		return uint64(val)
	case string:
		id, _ := strconv.ParseUint(val, 10, 64)
		return id
	default:
		if isPanic {
			panic("invalid context value type")
		}
		return 0
	}
}

func FullURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	// 如果前面有反向代理，Host 就在 Header 里
	host := r.Host
	if host == "" {
		host = r.Header.Get("X-Forwarded-Host")
	}
	return scheme + "://" + host + r.RequestURI
}

func Rand(n int) string {
	rand.Seed(time.Now().UnixNano())

	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func Md5(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
