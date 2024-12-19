package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/Esbaevnurdos/routes"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Set up MongoDB client options and connect
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Ping MongoDB to ensure connection is established
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")

	// Set up the router and routes
	router := mux.NewRouter()
	routes.RegisterStudentRoutes(router, client)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on port %s", port)
	http.ListenAndServe(":"+port, router)
}
