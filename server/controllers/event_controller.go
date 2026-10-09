package controllers

import (
	"context"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"example.com/events-app/config"
	"example.com/events-app/models"
	"github.com/gin-gonic/gin"
	"github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
	"gorm.io/gorm"
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
func GetEvents(c *gin.Context) {
	var events []models.Event // Siapkan keranjang (slice) kosong

	// 1. inisisasi dasar query di gorm => mengatur nilai pencarian berdasarkan query
	query := config.DB.Model(&models.Event{})

	// 2. tangkap fungsi filter by query
	search := c.Query("search")
	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%" + search + "%", "%" + search + "%")
	}

	// PAGINATION
	// 3. hitung jumlah data sebelum di limit
	var totalRows int64
	query.Count(&totalRows)

	// 4. 
	pageStr := c.DefaultQuery("page", "1") // nilai default jika tidak memasukkan data page
  limitStr := c.DefaultQuery("limit", "3")	// limit berapa data yang mau kita tampilkan per pagenya

  // konversi pageStr dan limitStr menjadi int
	page, errPage := strconv.Atoi(pageStr)
	if errPage != nil || page < 1 {
		page = 1
	}
	limit, errLimit := strconv.Atoi(limitStr)
	if errLimit != nil || limit < 1 {
		limit = 3	
	}

	// 5. hitung offset
	offset := (page - 1) * limit

	// 6. hitung data per halaman 
	totalPages := int(math.Ceil(float64(totalRows) / float64(limit)))

	// 7. eksekusi fungsinya
	// preload untuk menampilkan siapa user yang membuatnya (User dari tabel relasi) => data yang diambil dari User yang membuat event hanya id, name, dan emailnya
	if err := query.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "email")
	}).Limit(limit).Offset(offset).Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : "Gagal mengambil data event",
		})
		return
	}

	// tampilkan response jika berhasil
	c.JSON(http.StatusOK, gin.H{
		"message" : "Berhasil mengambil data event",
		"event" : events,
		"meta" : gin.H{
			"page" : page,
			"limit" : limit,
			"totalRows" : totalRows,
			"totalPages" : totalPages,
		},
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
	var eventData = config.DB.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "email")
	}).First(&event, paramsId).Error
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
// GET EVENT BY USER
// Mengambil data event milik user yang sedang login saat ini
// =====================================================================
func GetEventByUser(c *gin.Context) {
	var events []models.Event

	// 1. AMBIL ID USER DARI MIDDLEWARE (PERBAIKAN TYPO)
	// Pastikan huruf kecil 'u' pada "userID", sesuai dengan yang diset di Middleware.
	// Kita gunakan variabel 'exists' untuk memastikan user benar-benar sudah login.
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}

	// 2. QUERY DATABASE DENGAN FILTER WHERE
	// Preload: Tarik data pembuat event, tapi batasi hanya id, name, dan email.
	// Where: Filter tabel event, ambil HANYA baris yang kolom user_id-nya sama dengan userID yang sedang login.
	// Find: Masukkan semua hasilnya ke dalam keranjang 'events'.
	errEvent := config.DB.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "email")
	}).Where("user_id = ?", userID).Find(&events).Error

	// 3. TANGKAP JIKA ADA ERROR SAAT QUERY KE DATABASE
	// Catatan: Jika user belum punya event sama sekali, .Find() TIDAK MENGHASILKAN ERROR,
	// melainkan hanya mengembalikan array kosong []. Error di sini benar-benar error sistem/database.
	if errEvent != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Terjadi kesalahan saat mengambil data event",
		})
		return
	}

	// 4. TAMPILKAN HASILNYA
	c.JSON(http.StatusOK, gin.H{
		"message": "Data berhasil ditemukan",
		"events":  events,
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