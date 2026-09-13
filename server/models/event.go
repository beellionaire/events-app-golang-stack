package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	gorm.Model	
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Location    string `json:"location" binding:"required"`
	UserID      uint	 `json:"userid"`
	
	/* ========================================================
	- relasi ke tabel User || foreignkey berdasarkan kolom UserID kita [CASE SENSITIVE] 
	- Di dalam GORM, nilai yang dimasukkan ke dalam foreignKey harus sama persis (case-sensitive) dengan Nama Field Struct-nya, BUKAN nama kolom di database atau nama di tag JSON.
	- json:"-" => karena inputannya tidak ada 
	*/
	User 				User 	 `gorm:"foreignKey:UserID" json:"-"` 
	// ========================================================

	Datetime		time.Time `json:"datetime" binding:"required"`
}

/*

contoh function 

var events []Event = []Event{}

// fungsi untuk menyimpan event
// Save() disebut sebagai Method (atau lebih spesifiknya, Method dengan Value Receiver) karena menempel pada struct
func (e Event) Save() {
	events = append(events, e)
}

// fungsi menampilkan semua event
func GetAllEvents() []Event {
	return events
}

*/

