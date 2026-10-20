package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

//  middleware for application


const APIKeyHeader = "X-API-Key"

func APIKeyAuth(valid_keys ...string) gin.HandlerFunc{
	// hash the key
	hashed := make([][32]byte, len(valid_keys))
	for i,v := range valid_keys{
		hashed[i] = sha256.Sum256([]byte(v))
	}

	return func(c *gin.Context){
		key :=c.GetHeader(APIKeyHeader)
		if key == ""{
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing API key",
			})
			return 
		}

		provided := sha256.Sum256([]byte(key))
		valid := false

		for _, h :=range hashed{
			if subtle.ConstantTimeCompare(provided[:], h[:]) == 1{
				valid = true
			}
		}

		if !valid{
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API Key",
			})
			return
		}

		c.Next()
	}
}