package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware(origins []string) gin.HandlerFunc {
	originSet := make(map[string]bool, len(origins))
	for _, o := range origins {
		originSet[strings.TrimRight(o, "/")] = true
	}
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := ""

		if originSet[origin] {
			allowed = origin
		} else if len(originSet) == 1 {
			for o := range originSet {
				allowed = o
			}
		} else if originSet["*"] {
			allowed = "*"
		}

		if allowed != "" {
			c.Header("Access-Control-Allow-Origin", allowed)
		}
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
