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
	UserID      int		 `json:"userId"`
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

