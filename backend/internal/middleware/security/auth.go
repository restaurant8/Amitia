// SPDX-FileCopyrightText: 2026 彭旭
// SPDX-License-Identifier: AGPL-3.0-only
package security

import (
	"fmt"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/u-ai/backend/config"
	"github.com/u-ai/backend/pkg/comment/response"
	"github.com/u-ai/backend/pkg/util"
	"gorm.io/gorm"
)

var adminExistsFlag int32

func adminExists(db *gorm.DB) bool {
	if atomic.LoadInt32(&adminExistsFlag) == 1 {
		return true
	}
	if db == nil {
		return false
	}
	var n int64
	db.Table("auth_users").Where("role = ?", "admin").Count(&n)
	if n > 0 {
		atomic.StoreInt32(&adminExistsFlag, 1)
		return true
	}
	return false
}

func validateToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(config.AppCfg.JWT.Secret), nil
	})
	if err != nil || token == nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}

func setClaims(c *gin.Context, claims jwt.MapClaims) {
	if v, ok := claims["userId"]; ok {
		c.Set("userId", v)
	}
	if v, ok := claims["username"]; ok {
		c.Set("username", v)
	}
	if v, ok := claims["role"]; ok {
		c.Set("role", v)
	}
}

func extractBearer(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}

func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if tokenStr := extractBearer(c); tokenStr != "" {
			if claims, err := validateToken(tokenStr); err == nil {
				setClaims(c, claims)
				c.Next()
				return
			}
		}
		if !adminExists(db) {
			c.Next()
			return
		}
		util.ErrorResponse(c, response.Unauthorized, "请先登录", nil)
		c.Abort()
	}
}

func AuthQueryMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.Query("token")
		if tokenStr == "" {
			tokenStr = extractBearer(c)
		}
		if tokenStr != "" {
			if claims, err := validateToken(tokenStr); err == nil {
				setClaims(c, claims)
				c.Next()
				return
			}
		}
		if !adminExists(db) {
			c.Next()
			return
		}
		c.AbortWithStatus(401)
	}
}
