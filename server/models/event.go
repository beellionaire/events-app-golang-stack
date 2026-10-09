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
}
