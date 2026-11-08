package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/David89Lee/pyrolysis_ty_server/common"
	"github.com/David89Lee/pyrolysis_ty_server/router/response"
)

// token 格式: base64url(payload).base64url(hmacSHA256(payload, secret))
// payload = username|expireUnix

// GenerateToken 为指定用户名签发带过期时间的访问令牌
func GenerateToken(username string) (string, error) {
	expire := time.Now().Add(time.Duration(common.AuthSetting.TokenExpireHours) * time.Hour).Unix()
	payload := fmt.Sprintf("%s|%d", username, expire)
	sig := sign(payload)
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." +
		base64.RawURLEncoding.EncodeToString(sig), nil
}

// ValidateToken 校验令牌签名与有效期，成功时返回用户名
func ValidateToken(token string) (string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", false
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", false
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", false
	}
	if !hmac.Equal(sig, sign(string(payloadBytes))) {
		return "", false
	}
	segs := strings.Split(string(payloadBytes), "|")
	if len(segs) != 2 {
		return "", false
	}
	expire, err := strconv.ParseInt(segs[1], 10, 64)
	if err != nil || time.Now().Unix() > expire {
		return "", false
	}
	return segs[0], true
}

func sign(payload string) []byte {
	mac := hmac.New(sha256.New, []byte(common.AuthSetting.TokenSecret))
	mac.Write([]byte(payload))
	return mac.Sum(nil)
}

// AuthRequired 校验请求令牌：优先取 Authorization: Bearer <token>，
// 其次取查询参数 token（EventSource 无法自定义请求头，供 SSE 使用）
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := ""
		if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
			token = strings.TrimPrefix(h, "Bearer ")
		}
		if token == "" {
			token = c.Query("token")
		}
		if _, ok := ValidateToken(token); !ok {
			appG := response.Gin{C: c}
			appG.Response(200, response.ErrorAuthCheckTokenFail, nil)
			c.Abort()
			return
		}
		c.Next()
	}
}
