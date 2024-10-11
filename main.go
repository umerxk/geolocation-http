package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var locationCollection *mongo.Collection

type Location struct {
	UserID    string  `json:"userId"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func connectDB() {
	mongo_uri := "mongodb+srv://umersheraxk:helloworld@coordinates.3jpe9.mongodb.net/?retryWrites=true&w=majority&appName=Coordinates"

	clientOptions := options.Client().ApplyURI(mongo_uri)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")
	locationCollection = client.Database("user_locations").Collection("locations")
}

func createLocation(w http.ResponseWriter, r *http.Request) {
	var loc Location

	// Parse request body into the Location struct
	err := json.NewDecoder(r.Body).Decode(&loc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert data into MongoDB
	_, err = locationCollection.InsertOne(context.TODO(), bson.D{
		{Key: "userId", Value: loc.UserID},
		{Key: "latitude", Value: loc.Latitude},
		{Key: "longitude", Value: loc.Longitude},
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Location saved successfully!"})
}

func getAllLocations(w http.ResponseWriter, r *http.Request) {
	var locations []Location
	cursor, err := locationCollection.Find(context.TODO(), bson.D{})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var loc Location
		cursor.Decode(&loc)
		locations = append(locations, loc)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(locations)
}

func serverRunning(w http.ResponseWriter, r *http.Request) {
	fmt.Print("Server is up")
	fmt.Fprintln(w, "server is up")
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Endpoint hit: %s %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func main() {
	router := mux.NewRouter()

	// Connect to MongoDB
	connectDB()
	fmt.Print("xyz")
	// Define the POST route
	router.HandleFunc("/make", createLocation).Methods("POST")
	router.HandleFunc("/", serverRunning).Methods("GET")
	router.HandleFunc("/locations", getAllLocations).Methods("GET")

	router.Use(loggingMiddleware)

	// // Start the server
	// log.Fatal(http.ListenAndServe(":8000", router))
	// Enable CORS for all origins
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}), // You can specify the origin instead of "*"
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	// corsHandler := handlers.CORS(handlers.AllowedOrigins([]string{"*"}))

	// Start the server with CORS enabled
	log.Fatal(http.ListenAndServe(":8000", corsHandler(router)))
}
