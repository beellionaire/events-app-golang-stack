package main

import (
	"log"

	"example.com/events-app/config"
	"example.com/events-app/controllers"
	"example.com/events-app/middlewares"
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

		// route event
		api.GET("/events", controllers.GetEvents)
		api.GET("/events/:id", controllers.GetEventById) // membuat route detail
		
		// route auth
		api.POST("/auth/register", controllers.RegisterUser)
		api.POST("/auth/login", controllers.LoginUser)

		protected := api.Group("/")
		protected.Use(middlewares.RequiredAuth()) 
		{
			protected.GET("/events/user", controllers.GetEventByUser)
			protected.GET("/auth/me", controllers.GetCurrentUser)
			protected.POST("/events", controllers.CreateEvent)
			protected.PUT("/events/:id", controllers.UpdateEvent) // route update by id
			protected.DELETE("/events/:id", controllers.DeleteEvent) // route delete by id

			protected.POST("/booking/", controllers.CreateBookingEvent)
			protected.GET("/booking/user", controllers.GetBookingByUser)
			
			protected.DELETE("/booking/:id", controllers.DeleteBooking)
		}

	}

	server.Run(":8080") // menjalankan server

}
