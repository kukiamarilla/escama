package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"escama/infrastructure/eventstore"
	"escama/infrastructure/projections"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	fmt.Println("🔄 Iniciando migración de datos a proyecciones...")

	// Cargar variables de entorno desde .env
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Configurar Event Store
	eventStore, err := eventstore.NewMongoEventStore()
	if err != nil {
		log.Fatalf("❌ Error conectando al Event Store: %v", err)
	}

	// Configurar cliente MongoDB para proyecciones
	connectionString := os.Getenv("MONGODB_CONNECTION_STRING")
	if connectionString == "" {
		connectionString = "mongodb://localhost:27017/escama"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatalf("❌ Error conectando a MongoDB para proyecciones: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	// Configurar proyecciones
	projectionStore := projections.NewProjectionStore(mongoClient, "escama_read")

	fmt.Println("✅ Conexiones establecidas")

	// Limpiar proyecciones existentes (opcional)
	fmt.Print("🧹 ¿Deseas limpiar las proyecciones existentes antes de migrar? (y/N): ")
	var input string
	fmt.Scanln(&input)

	if input == "y" || input == "Y" || input == "yes" || input == "YES" {
		fmt.Println("🗑️  Limpiando proyecciones existentes...")
		if err := clearProjections(ctx, mongoClient); err != nil {
			log.Fatalf("❌ Error limpiando proyecciones: %v", err)
		}
		fmt.Println("✅ Proyecciones limpiadas")
	}

	// Obtener todos los eventos del Event Store
	fmt.Println("📖 Obteniendo eventos del Event Store...")
	storedEvents, err := eventStore.GetAllEvents(ctx, nil, nil)
	if err != nil {
		log.Fatalf("❌ Error obteniendo eventos: %v", err)
	}

	fmt.Printf("📊 Se encontraron %d eventos para procesar\n", len(storedEvents))

	// Procesar cada evento para actualizar las proyecciones
	processed := 0
	errors := 0

	for i, storedEvent := range storedEvents {
		if err := projectionStore.ProcessEvent(ctx, storedEvent); err != nil {
			log.Printf("⚠️  Error procesando evento %d (%s): %v", i+1, storedEvent.EventType, err)
			errors++
		} else {
			processed++
		}

		// Mostrar progreso cada 10 eventos
		if (i+1)%10 == 0 || i+1 == len(storedEvents) {
			fmt.Printf("📈 Progreso: %d/%d eventos procesados\n", i+1, len(storedEvents))
		}
	}

	fmt.Println("🎉 Migración completada!")
	fmt.Printf("✅ Eventos procesados exitosamente: %d\n", processed)
	if errors > 0 {
		fmt.Printf("⚠️  Eventos con errores: %d\n", errors)
	}

	// Mostrar estadísticas finales
	fmt.Println("\n📊 Verificando proyecciones creadas...")
	if err := showProjectionStats(ctx, projectionStore); err != nil {
		log.Printf("⚠️  Error obteniendo estadísticas: %v", err)
	}

	fmt.Println("✨ ¡Migración finalizada! Las proyecciones están listas para usar.")
}

func clearProjections(ctx context.Context, mongoClient *mongo.Client) error {
	// Limpiar colecciones de proyecciones
	database := mongoClient.Database("escama_read")

	if err := database.Collection("movements").Drop(ctx); err != nil {
		return fmt.Errorf("error dropping movements collection: %w", err)
	}

	if err := database.Collection("categories").Drop(ctx); err != nil {
		return fmt.Errorf("error dropping categories collection: %w", err)
	}

	return nil
}

func showProjectionStats(ctx context.Context, projectionStore *projections.ProjectionStore) error {
	// Obtener estadísticas de movimientos
	movements, total, err := projectionStore.GetMovements(ctx, nil, nil, 0, 0)
	if err != nil {
		return fmt.Errorf("error getting movements: %w", err)
	}

	incomeCount := 0
	expenseCount := 0
	var totalIncome, totalExpense float64

	for _, movement := range movements {
		if movement.Type == "income" {
			incomeCount++
			totalIncome += movement.Amount
		} else if movement.Type == "expense" {
			expenseCount++
			totalExpense += movement.Amount
		}
	}

	// Obtener estadísticas de categorías
	categories, err := projectionStore.GetCategories(ctx)
	if err != nil {
		return fmt.Errorf("error getting categories: %w", err)
	}

	fmt.Printf("📋 Movimientos totales: %d\n", total)
	fmt.Printf("💰 Ingresos: %d (₲%.0f)\n", incomeCount, totalIncome)
	fmt.Printf("💸 Gastos: %d (₲%.0f)\n", expenseCount, totalExpense)
	fmt.Printf("📈 Balance neto: ₲%.0f\n", totalIncome-totalExpense)
	fmt.Printf("🏷️  Categorías: %d\n", len(categories))

	return nil
}
