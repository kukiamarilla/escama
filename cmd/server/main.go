package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"escama/application/queries"
	"escama/infrastructure/eventstore"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Server struct {
	queryHandler *queries.MovementsQueryHandler
}

func main() {
	// Cargar variables de entorno desde .env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Configurar infrastructure - usar MongoDB
	mongoStore, err := eventstore.NewMongoEventStore()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoStore.Close()

	queryHandler := queries.NewMovementsQueryHandler(mongoStore)

	server := &Server{
		queryHandler: queryHandler,
	}

	// Configurar rutas
	r := mux.NewRouter()

	// API endpoints
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/movements", server.getMovements).Methods("GET")
	api.HandleFunc("/balance", server.getBalance).Methods("GET")
	api.HandleFunc("/expenses-by-category", server.getExpensesByCategory).Methods("GET")

	// Servir archivos estáticos (HTML, CSS, JS)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/"))).Methods("GET")

	// CORS middleware
	r.Use(corsMiddleware)

	port := ":8080"
	fmt.Printf("🚀 Escama Dashboard server starting on http://localhost%s\n", port)
	fmt.Println("📊 Access the dashboard at http://localhost:8080")

	log.Fatal(http.ListenAndServe(port, r))
}

func (s *Server) getMovements(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Parsear parámetros de fecha opcionales
	query := queries.GetMovementsQuery{}

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			query.StartDate = &startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Ajustar end_date al final del día
			endOfDay := endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			query.EndDate = &endOfDay
		}
	}

	// Parsear parámetros de paginación
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			query.Offset = (page - 1) * 10 // Asumiendo 10 elementos por página
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			query.Limit = limit
			if pageStr := r.URL.Query().Get("page"); pageStr != "" {
				if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
					query.Offset = (page - 1) * limit
				}
			}
		}
	}

	// Si se especifican parámetros de paginación, usar el endpoint paginado
	if query.Limit > 0 || query.Offset > 0 {
		if query.Limit == 0 {
			query.Limit = 10 // Default
		}

		paginatedMovements, err := s.queryHandler.GetPaginatedMovements(ctx, query)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting paginated movements: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(paginatedMovements)
		return
	}

	// Sin paginación, usar el endpoint original
	movements, err := s.queryHandler.GetMovements(ctx, query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting movements: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movements)
}

func (s *Server) getBalance(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Parsear parámetros de fecha (requeridos para balance)
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr == "" || endDateStr == "" {
		// Usar mes actual por defecto
		now := time.Now()
		startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate := startDate.AddDate(0, 1, -1)

		balance, err := s.queryHandler.GetBalance(ctx, queries.GetBalanceQuery{
			StartDate: startDate,
			EndDate:   endDate,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting balance: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(balance)
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Error(w, "Invalid start_date format (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		http.Error(w, "Invalid end_date format (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	// Ajustar end_date al final del día
	endOfDay := endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	balance, err := s.queryHandler.GetBalance(ctx, queries.GetBalanceQuery{
		StartDate: startDate,
		EndDate:   endOfDay,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting balance: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

func (s *Server) getExpensesByCategory(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Parsear parámetros de fecha opcionales
	query := queries.GetExpensesByCategoryQuery{}

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			query.StartDate = &startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Ajustar end_date al final del día
			endOfDay := endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			query.EndDate = &endOfDay
		}
	}

	expensesByCategory, err := s.queryHandler.GetExpensesByCategory(ctx, query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting expenses by category: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expensesByCategory)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
