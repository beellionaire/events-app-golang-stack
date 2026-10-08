package controllers

import (
	"context"
	"net/http"
	"os"
	"time"

	"example.com/events-app/config"
	"example.com/events-app/models"
	"github.com/gin-gonic/gin"
	"github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
)

// =====================================================================
// INIT IMAGEKIT
// Menginisialisasi koneksi ke server ImageKit menggunakan Private Key.
// =====================================================================
func initImageKit() *imagekit.Client {
	client := imagekit.NewClient(
		option.WithPrivateKey(os.Getenv("IMAGEKIT_PRIVATE_KEY")),
	)
	return &client
}

// =====================================================================
// CREATE EVENT
// Membuat event baru lengkap dengan upload gambar.
// =====================================================================
func CreateEvent(c *gin.Context) {
	// 1. Ambil ID user yang sedang login dari Middleware Auth
	userID, _ := c.Get("userID")

	// 2. Ambil file gambar dari form frontend
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mohon upload gambar"})
		return
	}
	defer file.Close() // Pastikan file ditutup setelah fungsi selesai

	// 3. Upload file ke ImageKit
	filename := header.Filename
	ik := initImageKit()
	uploadRes, errUpload := ik.Files.Upload(context.Background(), imagekit.FileUploadParams{
		File:     file,
		FileName: filename,
	})

	if errUpload != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gambar gagal diupload"})
		return
	}

	// 4. Ubah format string tanggal menjadi tipe data time.Time
	parsedTime, _ := time.Parse(time.RFC3339, c.PostForm("datetime"))

	// 5. Rangkai data event baru ke dalam struct
	event := models.Event{
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		Location:    c.PostForm("location"),
		Datetime:    parsedTime,
		Image:       uploadRes.URL,    // URL gambar untuk ditampilkan
		ImageID:     uploadRes.FileID, // ID gambar untuk keperluan hapus/update nanti
		UserID:      uint(userID.(int)),
	}

	// 6. Simpan ke database
	config.DB.Create(&event)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Berhasil membuat data",
		"event":   event,
	})
}

// =====================================================================
// GET ALL EVENTS
// Menampilkan semua data event yang ada di database.
// =====================================================================
func GetEvents(context *gin.Context) {
	var events []models.Event // Siapkan keranjang (slice) kosong

	// Tarik semua data dari tabel events dan masukkan ke keranjang
	config.DB.Find(&events)

	context.JSON(http.StatusOK, gin.H{
		"message": "Berhasil menampilkan semua data",
		"event":   events,
	})
}

// =====================================================================
// GET EVENT BY ID
// Menampilkan satu data event spesifik berdasarkan ID.
// =====================================================================
func GetEventById(context *gin.Context) {
	var event models.Event 

	// 1. Ambil ID dari parameter URL (contoh: /events/5)
	paramsId := context.Param("id")

	// 2. Cari data berdasarkan ID tersebut
	var eventData = config.DB.First(&event, paramsId).Error
	if eventData != nil {
		context.JSON(http.StatusNotFound, gin.H{"message": "Data tidak ditemukan"})
		return
	}

	// 3. Tampilkan datanya
	context.JSON(http.StatusOK, gin.H{
		"message": "Data detail event berhasil ditampilkan",
		"event":   event,
	})
}

// =====================================================================
// UPDATE EVENT
// Mengubah data event (Teks dan/atau Gambar).
// =====================================================================
func UpdateEvent(c *gin.Context) {
	// 1. Ambil ID user yang login & cari event di database
	userID, _ := c.Get("userID")
	var event models.Event
	ParamsId := c.Param("id")

	var eventData = config.DB.First(&event, ParamsId).Error
	if eventData != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data tidak ditemukan"})
		return
	}

	// 2. OTORISASI: Tolak jika yang login BUKAN pemilik event
	userIdInt := userID.(int)
	if event.UserID != uint(userIdInt) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Tidak bisa mengupdate event pengguna lain"})
		return
	}

	// 3. CEK DAN UPDATE GAMBAR (Opsional: Hanya jika user mengupload gambar baru)
	file, header, err := c.Request.FormFile("image")
	if err == nil {
		defer file.Close()
		ik := initImageKit()
		
		// Upload gambar baru
		uploadRes, errUpload := ik.Files.Upload(context.Background(), imagekit.FileUploadParams{
			File:     file,
			FileName: header.Filename,
		})

		if errUpload == nil {
			// Hapus gambar lama dari ImageKit untuk menghemat kapasitas storage
			if event.ImageID != "" {
				ik.Files.Delete(context.Background(), event.ImageID)
			}
			// Timpa data gambar di struct dengan yang baru
			event.Image = uploadRes.URL
			event.ImageID = uploadRes.FileID
		}
	}

	// 4. CEK DAN UPDATE TEKS 
	// (PENTING: Ditaruh DI LUAR blok gambar agar teks tetap bisa diupdate walau tanpa gambar)
	if name := c.PostForm("name"); name != "" {
		event.Name = name
	}
	if description := c.PostForm("description"); description != "" {
		event.Description = description
	}
	if location := c.PostForm("location"); location != "" {
		event.Location = location
	}
	if dateTimeStr := c.PostForm("datetime"); dateTimeStr != "" {
		parseTime, errParse := time.Parse(time.RFC3339, dateTimeStr)
		if errParse == nil {
			event.Datetime = parseTime
		}
	}

	// 5. Simpan semua perubahan ke database
	config.DB.Save(&event)
	c.JSON(http.StatusOK, gin.H{
		"message": "Data berhasil diupdate",
		"event":   event,
	})
}

// =====================================================================
// DELETE EVENT
// Menghapus data dari database dan menghapus filenya dari ImageKit.
// =====================================================================
func DeleteEvent(c *gin.Context) {
	// 1. Ambil data user login dan cari event di database
	userID, _ := c.Get("userID")
	var event models.Event
	ParamsId := c.Param("id")

	var eventData = config.DB.First(&event, ParamsId).Error
	if eventData != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data tidak ditemukan"})
		return
	}

	// 2. OTORISASI: Pastikan pemilik yang menghapus
	userIdInt := userID.(int)
	if event.UserID != uint(userIdInt) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Tidak dapat menghapus event milik user lain"})
		return
	}

	// 3. HAPUS GAMBAR FISIK: Bersihkan storage ImageKit
	if event.ImageID != "" {
		ik := initImageKit()
		ik.Files.Delete(context.Background(), event.ImageID)
	}

	// 4. HAPUS DATA DATABASE: 
	// Menggunakan Unscoped() agar data benar-benar hilang dari database (Hard Delete),
	// bukan sekadar disembunyikan (Soft Delete).
	config.DB.Unscoped().Delete(&event)
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Data berhasil di hapus",
	})
}