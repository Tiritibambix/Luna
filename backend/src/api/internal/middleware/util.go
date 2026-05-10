package middleware

import (
	"luna-backend/types"

	"github.com/gin-gonic/gin"
)

func Prepend(first gin.HandlerFunc, rest []gin.HandlerFunc) []gin.HandlerFunc {
	return append([]gin.HandlerFunc{first}, rest...)
}

func RequirePermissionAndBody[T any](perm types.Permission, handler func(c *gin.Context, obj *T)) []gin.HandlerFunc {
	return Prepend(RequirePermissions(perm), WithBody(handler))
}
