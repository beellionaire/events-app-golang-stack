package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	gorm.Model	
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Image				string `json:"image"`
	ImageID 		string `json:"imageId"`  // untuk menghapus gambar berdasarkan id nya
	Location    string `json:"location" binding:"required"`
	UserID      uint	 `json:"userid"`
	
	/* ========================================================
	- relasi ke tabel User || foreignkey berdasarkan kolom UserID kita [CASE SENSITIVE] 
	- Di dalam GORM, nilai yang dimasukkan ke dalam foreignKey harus sama persis (case-sensitive) dengan Nama Field Struct-nya, BUKAN nama kolom di database atau nama di tag JSON.
	- json:"-" => karena inputannya tidak ada 
	- json:"user" => menampilkan list user yang membuat event
	*/
	User 				User 	 `gorm:"foreignKey:UserID" json:"user"` 
	// ========================================================

	Datetime		time.Time `json:"datetime" binding:"required"`

// =====================================================================
// RELASI ONE-TO-MANY (Satu Event memiliki Banyak Pendaftar/Booking)
// =====================================================================

Booking []Booking `gorm:"foreignKey:EventID" json:"listBooking"`

/* PENJELASAN PER KATA:
   1. Booking
      Nama field atau properti penampung di dalam struct Event.
   2. []Booking
      Tipe datanya berupa Slice/Array (ditandai dengan kurung siku []).
      Artinya: Field ini disiapkan untuk menampung BANYAK data dari struct Booking sekaligus.
   3. gorm:"foreignKey:EventID"
      Instruksi untuk GORM: "Hai GORM, untuk mengisi data array di atas, tolong pergi ke 
      tabel 'bookings'. Lalu ambil semua data yang nilai kolom EventID-nya sama dengan 
      ID Event ini."
   4. json:"listBooking"
      Instruksi untuk Response API: Saat data ini dicetak menjadi JSON (ke Postman/Frontend), 
      jangan gunakan nama "Booking", tapi ubah namanya menjadi "listBooking" agar 
      Frontend tahu bahwa isinya adalah daftar (list) panjang.
*/
}
