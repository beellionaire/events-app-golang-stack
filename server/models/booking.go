package models

import "gorm.io/gorm"

type Booking struct {
	gorm.Model
	BookingCode string `json:"bookingCode"`
	Phone string `json:"phone"`

	UserID uint `json:"userId"`
	User 	User `gorm:"foreignKey:UserID" json:"user"`

	EventID uint `json:"eventId"`
	Event Event `gorm:"foreignKey:EventID" json:"event"`
}