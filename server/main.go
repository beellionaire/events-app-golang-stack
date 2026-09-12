package main

import (
	"log"
	"net/http"

	"example.com/events-app/config"
	"example.com/events-app/models"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	// menjalankan environtment
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// mengatur config ke database
	config.ConnectDB();

	server := gin.Default();

	// Route => mengatur jalur aplikasi kita
	
	api := server.Group("/api") // membuat route group api
	{
		api.POST("/events", createEvent)
		api.GET("/events", getEvents)
	}

	server.Run(":8080") // menjalankan server

}

// Function Handlers => menjalankan bussines logicnya
func getEvents(context *gin.Context) {
	// variabel untuk menampung nilainya
	events := models.GetAllEvents()

	context.JSON(http.StatusOK, events)
}

func createEvent(context *gin.Context) {
	var event models.Event		// mengambil inputan dari modals
	err := context.ShouldBindJSON(&event)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"message" : "Couldnt parse request data",
			"error" : err.Error(),
		})
		return
	}

	// dummy data
	event.UserId = 1

	// simpan event => panggil function save yang ada di modal
	event.Save();

	context.JSON(http.StatusCreated, gin.H{
		"message" : "Created Event",
		"event" : event,
	})

}