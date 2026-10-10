package controllers

import (
	"fmt"
	"net/http"
	"time"

	"example.com/events-app/config"
	"example.com/events-app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BookingInput adalah cetakan untuk menangkap data dari body (JSON) Postman/Frontend
type BookingInput struct {
	Phone   string `json:"phone" binding:"required"`
	EventID int    `json:"eventId" binding:"required"`
}

func CreateBookingEvent(c *gin.Context) {
	// 1. AMBIL ID USER DARI MIDDLEWARE
	userID, _ := c.Get("userID")

	var input BookingInput
	var booking models.Booking

	// 2. KONVERSI TIPE DATA USER ID
	// Middleware menyimpan sebagai interface{} -> diubah ke int -> diubah ke uint
	userIDint := userID.(int)
	userIDuint := uint(userIDint)

	// 3. VALIDASI INPUT DARI USER
	// ShouldBindJSON akan mengecek apakah JSON yang dikirim sesuai dengan struct BookingInput.
	// Jika 'phone' atau 'eventId' kosong, maka akan error karena ada tag binding:"required".
	errValidation := c.ShouldBindJSON(&input)
	if errValidation != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errValidation.Error(),
		})
		return
	}

	// 4. CEK DUPLIKASI PEMESANAN (MENCEGAH DOUBLE BOOKING)
	// Cari di database: "Apakah ada data booking dengan User ID ini DAN Event ID ini?"
	// PERBAIKAN: Ubah nama variabel penampung error menjadi 'errCekBooking'
	errCekBooking := config.DB.Where("user_id = ? AND event_id = ?", userIDuint, input.EventID).First(&booking).Error

	// Jika hasilnya 'nil' (artinya TIDAK ADA error saat mencari / datanya KETEMU),
	// berarti user tersebut sudah pernah mendaftar. Tolak permintaannya!
	if errCekBooking == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Anda sudah memesan event ini",
		})
		return
	}

	// 5. GENERATE KODE BOOKING UNIK
	// Menggabungkan teks "RF-", Tanggal Hari Ini, ID Event, dan ID User.
	// Contoh Hasil: RF-20261010E5U12 (Dipesan tanggal 10 Okt 2026, untuk Event ID 5, oleh User ID 12)
	CodeBooking := fmt.Sprintf(
		"RF-%sE%dU%d",
		time.Now().Format("20060102"),
		input.EventID,
		userIDuint,
	)

	// 6. SIAPKAN DATA UNTUK DISIMPAN
	BookingData := models.Booking{
		Phone:       input.Phone,
		// PERBAIKAN: Wajib dikonversi menjadi uint agar tidak bentrok dengan models.Booking
		EventID:     uint(input.EventID), 
		BookingCode: CodeBooking,
		UserID:      userIDuint,
	}

	// 7. SIMPAN KE DATABASE
	errCreateBooking := config.DB.Create(&BookingData).Error
	if errCreateBooking != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mendaftar event",
		})
		return
	}

	// 8. TAMPILKAN PESAN SUKSES
	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mendaftar event",
	})
}

// function untuk get event berdasarkan user
func GetBookingByUser(c *gin.Context) {
	var booking []models.Booking
	userID, _ := c.Get("userID")

	errBookingData := config.DB.Preload("Event").Preload("Event.User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id","name","email")
	}).Where("user_id = ?", userID).Find(&booking).Error

	if errBookingData != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"erroro" : "Event tidak ditemukan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"booking" : booking,
	})
}


// function delete booking
func DeleteBooking(c *gin.Context) {
	// 1. Ambil data user yang sedang login dari 'kantong' Middleware
	userID, _ := c.Get("userID")
	var booking models.Booking

	// 2. Ambil parameter ID dari URL Postman (contoh URL: /bookings/5 -> paramsId = "5")
	paramsId := c.Param("id")
	
	// 3. CARI DATA DI DATABASE
	// Perintah First() akan mencari 1 baris pertama di tabel bookings yang ID-nya sama dengan paramsId.
	// Jika datanya tidak ada, GORM akan menghasilkan error (gorm.ErrRecordNotFound)
	bookingData := config.DB.First(&booking, paramsId).Error

	// Jika ada error (artinya datanya tidak ketemu)
	if bookingData != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Booking tidak ditemukan", // Teks pesan diperbaiki agar lebih akurat
		})
		return // Hentikan fungsi di sini agar tidak lanjut ke bawah
	}

	// 4. OTORISASI (PENCEGAHAN HACKING)
	// Cek apakah UserID pembuat booking SAMA DENGAN UserID orang yang sedang login.
	// Harus dikonversi menjadi uint(userID.(int)) agar sepadan dengan booking.UserID
	if booking.UserID != uint(userID.(int)) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Tidak dapat menghapus booking user lain",
		})
		return
	}

	// 5. HAPUS DATA (HARD DELETE)
	// Unscoped() digunakan agar data benar-benar musnah dari database.
	// Jika tanpa Unscoped(), GORM hanya akan mengisi kolom 'DeletedAt' (Soft Delete).
	config.DB.Unscoped().Delete(&booking)
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Booking event berhasil dihapus",
	})
	
	// 'return' di baris paling akhir fungsi sebenarnya tidak diperlukan di Golang.
	// Anda bisa menghapusnya.
}