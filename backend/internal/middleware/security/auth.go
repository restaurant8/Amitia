// SPDX-FileCopyrightText: 2026 彭旭
// SPDX-License-Identifier: AGPL-3.0-only
package security

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/u-ai/backend/config"
	"github.com/u-ai/backend/pkg/comment/response"
	"github.com/u-ai/backend/pkg/util"
)

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

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if len(auth) <= 7 || auth[:7] != "Bearer " {
			util.ErrorResponse(c, response.Unauthorized, "请先登录", nil)
			c.Abort()
			return
		}
		claims, err := validateToken(auth[7:])
		if err != nil {
			util.ErrorResponse(c, response.InvalidToken, "令牌无效或已过期", nil)
			c.Abort()
			return
		}
		setClaims(c, claims)
		c.Next()
	}
}

func AuthQueryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.Query("token")
		if tokenStr == "" {
			auth := c.GetHeader("Authorization")
			if len(auth) > 7 && auth[:7] == "Bearer " {
				tokenStr = auth[7:]
			}
		}
		if tokenStr == "" {
			c.AbortWithStatus(401)
			return
		}
		claims, err := validateToken(tokenStr)
		if err != nil {
			c.AbortWithStatus(401)
			return
		}
		setClaims(c, claims)
		c.Next()
	}
}
