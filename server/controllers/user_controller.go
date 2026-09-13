package controllers

import (
	"net/http"
	"os"
	"time"

	"example.com/events-app/config"
	"example.com/events-app/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthInputRegister struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=5"`
}

type AuthInputLogin struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// FUNCTION REGISTER =====================================================
func RegisterUser(context *gin.Context) {

	var input AuthInputRegister

	// validation register
	err := context.ShouldBindJSON(&input)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error" : err.Error(),
		})
		return
	}

	// hashing password (enkripsi)
	hashedPassword, errHash := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if errHash != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error" : "Gagal mengenkripsi password",
		})
		return
	}

	// simpan ke database
	user := models.User {
		Name: input.Name,
		Email: input.Email,
		Password: string(hashedPassword),
	}

	// create data
	userCreated := config.DB.Create(&user).Error
	if userCreated != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error" : "Email sudah terdaftar",
		})
		return
	}

	// response success
	context.JSON(http.StatusCreated, gin.H{
		"message" : "Data berhasil dibuat",
		"user" : gin.H{
			"id" : user.ID,
			"name" : user.Name,
			"email" : user.Email,
			"events" : user.Events,
		},
	})		

}

// FUNCTION LOGIN =====================================================
func LoginUser(context *gin.Context) {
	var input AuthInputLogin

	err := context.ShouldBindJSON(&input)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error" : err.Error(),
		})
		return
	}

	// cek apakah data user sudah ada di database atau belum
	var user models.User
	userData := config.DB.Where("email = ?", input.Email).First(&user).Error
	if userData != nil {
		context.JSON(http.StatusUnauthorized, gin.H{
			"error" : "Email belum terdaftar",
		})
		return
	}

	// cek apakah password sudah sesuai dengan emailnya
	errMatchPassword := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if errMatchPassword != nil {
		context.JSON(http.StatusUnauthorized, gin.H{
			"error" : "Password tidak sesuai",
		})
		return
	}

	// buat token apakah user sedang login atau tidak
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub" : user.ID,
		"exp" : time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	// masukkan jwt_secret ke dalam token
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error" : "Tidak dapat mengambil token",
		})
		return
	}

	// buat response
	context.JSON(http.StatusOK, gin.H{
		"message" : "Login berhasil",
		"token" : tokenString,
		"user" : gin.H{
			"id" : user.ID,
			"name" : user.Name,
			"email" : user.Email,
			"events" :user.Events,
		},
	})
}
