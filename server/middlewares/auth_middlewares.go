package middlewares

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RequiredAuth() gin.HandlerFunc {
	
	// cek apakah token berhasil diambil
	return func(context *gin.Context) {
		tokenString := context.GetHeader("Authorization") // mengambil data header yang dikirimkan
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error" : "Tidak dapat mengambil token",
			})
			return
		}

		// ambil token
		tokenString = strings.TrimPrefix(tokenString, "Bearer ") // menghapus awalan (prefix) tertentu dari awal sebuah string jika string tersebut memilikinya.
		token,_ := jwt.Parse(tokenString, func (token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		// ambil sub (data user yang kita gunakan login) dan cek apakah token valid
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			context.Set("userID", int(claims["sub"].(float64)))
			context.Next()
		} else {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H {
				"error" : "Token tidak valid",
			})
			return
		}
	}
}

