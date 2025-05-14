package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"        
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"user-tx-backend/graph"
	"user-tx-backend/handler"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using existing env variables")
	}
	uri := os.Getenv("NEO4J_URI")
	user := os.Getenv("NEO4J_USER")
	pass := os.Getenv("NEO4J_PASS")
	seed := os.Getenv("SEED_DATA")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	drv, err := graph.NewDriver(uri, user, pass)
	if err != nil {
		log.Fatalf("DataBase connection failed: %v", err)
	}
	defer drv.Close()

	// seed sample data
	if seed == "true" {
		if err := graph.SeedData(drv); err != nil {
			log.Fatalf("Data seeding failed: %v", err)
		}
		log.Println("Sample data seeded")
		time.Sleep(500 * time.Millisecond)
	}

	router := mux.NewRouter()
	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
	)

	// routes
	h := handler.NewHandler(drv)
	router.HandleFunc("/api/users", h.CreateUser).Methods("POST")
	router.HandleFunc("/api/users", h.GetAllUsers).Methods("GET")
	router.HandleFunc("/api/transactions", h.CreateTransaction).Methods("POST")
	router.HandleFunc("/api/transactions", h.GetAllTransactions).Methods("GET")
	router.HandleFunc("/api/relationships/user/{id}", h.GetUserRelationships).Methods("GET")
	router.HandleFunc("/api/relationships/transaction/{id}", h.GetTransactionRelationships).Methods("GET")
    router.HandleFunc("/api/analytics/shortest-path/users/{from}/{to}", h.GetUserShortestPath).Methods("GET")
    router.HandleFunc("/api/export/json", h.ExportGraphJSON).Methods("GET")
    router.HandleFunc("/api/export/csv",  h.ExportGraphCSV).Methods("GET")
	router.HandleFunc("/api/analytics/transaction-clusters", h.GetTransactionClusters).Methods("GET")
    
   
	addr := ":" + port
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, cors(router)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
