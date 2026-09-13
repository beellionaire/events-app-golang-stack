package controllers

import (
	"net/http"

	"example.com/events-app/config"
	"example.com/events-app/models"
	"github.com/gin-gonic/gin"
)

// Function create data ===========================================
func CreateEvent(context *gin.Context) {
	var event models.Event

	err := context.ShouldBindJSON(&event)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error" : err.Error(),
		})
		return // agar tidak melanjutkan mengeksekusi kode di bawahnya
	}

	// data dummy untuk UserID jika belum membuat fitur autentikasi
	event.UserID = 1

	// masukkan data ke dalam database
	config.DB.Create(&event)
	context.JSON(http.StatusCreated, gin.H{
		"message" : "Berhasil membuat data",
		"event" : event,
	})
}

// Function Get all data ===========================================
func GetEvents(context *gin.Context) {
	// menampung semua data
	var events []models.Event	// menampung array kosong yang nilainya diambil dari models.event

	// mengambil semua data
	config.DB.Find(&events) // mengambil semua data berdasarkan data yang berada di event kita

	context.JSON(http.StatusOK, gin.H{
		"message" : "Berhasil menampilkan semua data", // tampilkan pesan
		"event" : events, 		// tampilkan data yang sudah dibuat
	})
}

// Function show detail data =======================================
func GetEventById(context *gin.Context) {
	var event models.Event // hanya mengambil satu objek saja
	
	// mengambil parameter
	paramsId := context.Param("id") // id => adalah parameter yang ada di route | api.GET("/events/:id", ) |

	// buat kondisi jika data tidak ditemukan
	var eventData = config.DB.First(&event, paramsId). Error // mengambil satu data berdasarkan kondisi event kita | .Error (untuk menampilkan jika terjadi error)
	if eventData != nil {
		context.JSON(http.StatusNotFound, gin.H{
			"message" : "Data tidak ditemukan",
		})
		return // agar tidak melanjutkan mengeksekusi kode di bawahnya
	}

	// tampilkan detail datanya
	context.JSON(http.StatusOK, gin.H{
		"message" : "Data detail event berhasil ditampilkan",
		"event" : event,
	})

}

// Function update data ==============================================
func UpdateEvent(context *gin.Context) {
	var event models.Event

	// mengambil id yang ingin di update 
	ParamsId := context.Param("id")

	// kondisi ketika not found
	var eventData = config.DB.First(&event, ParamsId).Error
	if eventData != nil {
		context.JSON(http.StatusNotFound, gin.H{
			"error" : "Data tidak ditemukan",
		})
		return
	}

	// input
	var input models.Event

	err := context.ShouldBindJSON(&input)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error" : err.Error(),
		})
		return
	}

	// update datanya
	config.DB.Model(&event).Updates(input)
	context.JSON(http.StatusOK, gin.H{
		"message" : "Data berhasil diupdate",  // tampilkan pesan
		"event" : event,		// tampilkan data yang sudah diupdate
	})
}

// Function Delete ===============================================
func DeleteEvent(context *gin.Context) {
	var event models.Event
	ParamsId := context.Param("id")

	var eventData = config.DB.First(&event, ParamsId).Error
	if eventData != nil {
		context.JSON(http.StatusNotFound, gin.H{
			"error" : "Data tidak ditemukan",
		})
		return
	}

	// Delete data
	config.DB.Unscoped().Delete(&event)
	// kenapa unscoped() =>  karena jika tidak menggunakan unscoped maka yang terdelete hanya softdeletenya saja buka forcedelete
	context.JSON(http.StatusOK, gin.H{
		"message": "Data berhasil di hapus",
	})
}