package config

import (
	"log"
	"os"

	"example.com/events-app/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := os.Getenv("DATABASE_URI") // mengambil data dari env

	// menampilkan error ketika env kosong
	if dsn == "" {
		log.Fatal("Environtment variabel belum diisi")
	}

	// konfigurasi koneksi db ke postgresql
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	// mengecek sudah terkoneksi atau belum
	if err != nil {
		log.Fatal("Gagal bos, benerin dulu koneksinya", err)
	}

	// jika berhasil maka setiap ada perubahan langsung terganti
	err = database.AutoMigrate(&models.Event{})
	if err != nil {
		log.Fatal("Gagal melakukan migration database", err)
	}

	// ketika koneksi berhasil
	DB = database
	log.Println("Berhasil terkoneksi ke database")
}