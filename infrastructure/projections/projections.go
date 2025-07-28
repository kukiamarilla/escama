package projections

import (
	"context"
	"fmt"
	"log"
	"time"

	"escama/domain/events"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MovementProjection representa un movimiento en la base de datos de lectura
type MovementProjection struct {
	ID           string    `bson:"_id" json:"id"`
	Type         string    `bson:"type" json:"type"`
	CategoryID   string    `bson:"category_id" json:"category_id"`
	CategoryName string    `bson:"category_name" json:"category_name"`
	Amount       float64   `bson:"amount" json:"amount"`
	Description  *string   `bson:"description" json:"description"`
	Date         time.Time `bson:"date" json:"date"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at" json:"updated_at"`
	IsDeleted    bool      `bson:"is_deleted" json:"is_deleted"`
}

// CategoryProjection representa una categoría en la base de datos de lectura
type CategoryProjection struct {
	ID        string    `bson:"_id" json:"id"`
	Name      string    `bson:"name" json:"name"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	IsDeleted bool      `bson:"is_deleted" json:"is_deleted"`
}

// ProjectionStore maneja las proyecciones en MongoDB
type ProjectionStore struct {
	client               *mongo.Client
	database             *mongo.Database
	movementsCollection  *mongo.Collection
	categoriesCollection *mongo.Collection
}

func NewProjectionStore(client *mongo.Client, databaseName string) *ProjectionStore {
	database := client.Database(databaseName)

	return &ProjectionStore{
		client:               client,
		database:             database,
		movementsCollection:  database.Collection("movements"),
		categoriesCollection: database.Collection("categories"),
	}
}

// ProcessEvent procesa un evento y actualiza las proyecciones
func (ps *ProjectionStore) ProcessEvent(ctx context.Context, event events.StoredEvent) error {
	switch event.EventType {
	case "CategoryCreated":
		return ps.handleCategoryCreated(ctx, event)
	case "ExpenseCreated":
		return ps.handleExpenseCreated(ctx, event)
	case "IncomeCreated":
		return ps.handleIncomeCreated(ctx, event)
	case "ExpenseUpdated":
		return ps.handleExpenseUpdated(ctx, event)
	case "IncomeUpdated":
		return ps.handleIncomeUpdated(ctx, event)
	case "ExpenseDeleted":
		return ps.handleExpenseDeleted(ctx, event)
	case "IncomeDeleted":
		return ps.handleIncomeDeleted(ctx, event)
	default:
		log.Printf("Unknown event type: %s", event.EventType)
		return nil
	}
}

func (ps *ProjectionStore) handleCategoryCreated(ctx context.Context, event events.StoredEvent) error {
	categoryID := ps.getStringFromPayload(event.Payload, "CategoryID", "category_id")
	name := ps.getStringFromPayload(event.Payload, "Name", "name")

	if categoryID == "" || name == "" {
		return fmt.Errorf("invalid category created event: missing required fields")
	}

	category := CategoryProjection{
		ID:        categoryID,
		Name:      name,
		CreatedAt: event.OccurredAt,
		UpdatedAt: event.OccurredAt,
		IsDeleted: false,
	}

	_, err := ps.categoriesCollection.ReplaceOne(
		ctx,
		bson.M{"_id": categoryID},
		category,
		options.Replace().SetUpsert(true),
	)

	if err != nil {
		return fmt.Errorf("failed to upsert category projection: %w", err)
	}

	log.Printf("Category projection updated: %s - %s", categoryID, name)
	return nil
}

func (ps *ProjectionStore) handleExpenseCreated(ctx context.Context, event events.StoredEvent) error {
	return ps.handleMovementCreated(ctx, event, "expense")
}

func (ps *ProjectionStore) handleIncomeCreated(ctx context.Context, event events.StoredEvent) error {
	return ps.handleMovementCreated(ctx, event, "income")
}

func (ps *ProjectionStore) handleMovementCreated(ctx context.Context, event events.StoredEvent, movementType string) error {
	var movementID, categoryID string
	if movementType == "expense" {
		movementID = ps.getStringFromPayload(event.Payload, "ExpenseID", "expense_id")
	} else {
		movementID = ps.getStringFromPayload(event.Payload, "IncomeID", "income_id")
	}

	categoryID = ps.getStringFromPayload(event.Payload, "CategoryID", "category_id")
	amount := ps.getFloat64FromPayload(event.Payload, "Amount", "amount")
	description := ps.getStringPtrFromPayload(event.Payload, "Description", "description")
	date := ps.getTimeFromPayload(event.Payload, "Date", "date")

	if movementID == "" {
		return fmt.Errorf("invalid %s created event: missing ID", movementType)
	}

	if date.IsZero() {
		date = event.OccurredAt
	}

	// Obtener nombre de la categoría
	categoryName := "Sin categoría"
	if categoryID != "" {
		var category CategoryProjection
		err := ps.categoriesCollection.FindOne(ctx, bson.M{"_id": categoryID}).Decode(&category)
		if err == nil {
			categoryName = category.Name
		}
	}

	movement := MovementProjection{
		ID:           movementID,
		Type:         movementType,
		CategoryID:   categoryID,
		CategoryName: categoryName,
		Amount:       amount,
		Description:  description,
		Date:         date,
		CreatedAt:    event.OccurredAt,
		UpdatedAt:    event.OccurredAt,
		IsDeleted:    false,
	}

	_, err := ps.movementsCollection.ReplaceOne(
		ctx,
		bson.M{"_id": movementID},
		movement,
		options.Replace().SetUpsert(true),
	)

	if err != nil {
		return fmt.Errorf("failed to upsert movement projection: %w", err)
	}

	log.Printf("%s projection updated: %s - ₲%.0f", movementType, movementID, amount)
	return nil
}

func (ps *ProjectionStore) handleExpenseUpdated(ctx context.Context, event events.StoredEvent) error {
	return ps.handleMovementUpdated(ctx, event, "expense")
}

func (ps *ProjectionStore) handleIncomeUpdated(ctx context.Context, event events.StoredEvent) error {
	return ps.handleMovementUpdated(ctx, event, "income")
}

func (ps *ProjectionStore) handleMovementUpdated(ctx context.Context, event events.StoredEvent, movementType string) error {
	var movementID string
	if movementType == "expense" {
		movementID = ps.getStringFromPayload(event.Payload, "ExpenseID", "expense_id")
	} else {
		movementID = ps.getStringFromPayload(event.Payload, "IncomeID", "income_id")
	}

	if movementID == "" {
		return fmt.Errorf("invalid %s updated event: missing ID", movementType)
	}

	// Obtener los nuevos valores del evento
	categoryID := ps.getStringFromPayload(event.Payload, "CategoryID", "category_id")
	amount := ps.getFloat64FromPayload(event.Payload, "Amount", "amount")
	description := ps.getStringPtrFromPayload(event.Payload, "Description", "description")
	date := ps.getTimeFromPayload(event.Payload, "Date", "date")

	// Obtener nombre de la categoría
	categoryName := "Sin categoría"
	if categoryID != "" {
		var category CategoryProjection
		err := ps.categoriesCollection.FindOne(ctx, bson.M{"_id": categoryID}).Decode(&category)
		if err == nil {
			categoryName = category.Name
		}
	}

	// Actualizar la proyección
	update := bson.M{
		"$set": bson.M{
			"category_id":   categoryID,
			"category_name": categoryName,
			"amount":        amount,
			"description":   description,
			"date":          date,
			"updated_at":    event.OccurredAt,
		},
	}

	_, err := ps.movementsCollection.UpdateOne(ctx, bson.M{"_id": movementID}, update)
	if err != nil {
		return fmt.Errorf("failed to update movement projection: %w", err)
	}

	log.Printf("%s projection updated: %s", movementType, movementID)
	return nil
}

func (ps *ProjectionStore) handleExpenseDeleted(ctx context.Context, event events.StoredEvent) error {
	return ps.handleMovementDeleted(ctx, event, "expense")
}

func (ps *ProjectionStore) handleIncomeDeleted(ctx context.Context, event events.StoredEvent) error {
	return ps.handleMovementDeleted(ctx, event, "income")
}

func (ps *ProjectionStore) handleMovementDeleted(ctx context.Context, event events.StoredEvent, movementType string) error {
	var movementID string
	if movementType == "expense" {
		movementID = ps.getStringFromPayload(event.Payload, "ExpenseID", "expense_id")
	} else {
		movementID = ps.getStringFromPayload(event.Payload, "IncomeID", "income_id")
	}

	if movementID == "" {
		return fmt.Errorf("invalid %s deleted event: missing ID", movementType)
	}

	// Marcar como eliminado (soft delete)
	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"updated_at": event.OccurredAt,
		},
	}

	_, err := ps.movementsCollection.UpdateOne(ctx, bson.M{"_id": movementID}, update)
	if err != nil {
		return fmt.Errorf("failed to delete movement projection: %w", err)
	}

	log.Printf("%s projection deleted: %s", movementType, movementID)
	return nil
}

// Helper functions
func (ps *ProjectionStore) getStringFromPayload(payload map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := payload[key]; ok {
			if strVal, ok := val.(string); ok {
				return strVal
			}
		}
	}
	return ""
}

func (ps *ProjectionStore) getFloat64FromPayload(payload map[string]interface{}, keys ...string) float64 {
	for _, key := range keys {
		if val, ok := payload[key]; ok {
			switch v := val.(type) {
			case float64:
				return v
			case float32:
				return float64(v)
			case int:
				return float64(v)
			case int64:
				return float64(v)
			}
		}
	}
	return 0
}

func (ps *ProjectionStore) getStringPtrFromPayload(payload map[string]interface{}, keys ...string) *string {
	for _, key := range keys {
		if val, ok := payload[key]; ok {
			if strVal, ok := val.(string); ok && strVal != "" {
				return &strVal
			}
		}
	}
	return nil
}

func (ps *ProjectionStore) getTimeFromPayload(payload map[string]interface{}, keys ...string) time.Time {
	for _, key := range keys {
		if val, ok := payload[key]; ok {
			switch v := val.(type) {
			case string:
				// RFC3339 format
				if parsedDate, err := time.Parse(time.RFC3339, v); err == nil {
					return parsedDate
				}
				// Otros formatos comunes
				formats := []string{
					"2006-01-02T15:04:05Z07:00",
					"2006-01-02T15:04:05.000Z",
					"2006-01-02T15:04:05",
					"2006-01-02 15:04:05",
					"2006-01-02",
				}
				for _, format := range formats {
					if parsedDate, err := time.Parse(format, v); err == nil {
						return parsedDate
					}
				}
			case time.Time:
				return v
			}
		}
	}
	return time.Time{}
}

// Query methods for reading projections
func (ps *ProjectionStore) GetMovements(ctx context.Context, startDate, endDate *time.Time, limit, offset int) ([]MovementProjection, int, error) {
	filter := bson.M{"is_deleted": false}

	// Agregar filtros de fecha
	if startDate != nil || endDate != nil {
		dateFilter := bson.M{}
		if startDate != nil {
			dateFilter["$gte"] = *startDate
		}
		if endDate != nil {
			dateFilter["$lte"] = *endDate
		}
		filter["date"] = dateFilter
	}

	// Contar total
	total, err := ps.movementsCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count movements: %w", err)
	}

	// Configurar opciones de búsqueda
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"date", -1}, {"created_at", -1}})
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}
	if offset > 0 {
		findOptions.SetSkip(int64(offset))
	}

	cursor, err := ps.movementsCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find movements: %w", err)
	}
	defer cursor.Close(ctx)

	var movements []MovementProjection
	if err := cursor.All(ctx, &movements); err != nil {
		return nil, 0, fmt.Errorf("failed to decode movements: %w", err)
	}

	return movements, int(total), nil
}

func (ps *ProjectionStore) GetCategories(ctx context.Context) ([]CategoryProjection, error) {
	filter := bson.M{"is_deleted": false}
	findOptions := options.Find().SetSort(bson.M{"name": 1})

	cursor, err := ps.categoriesCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to find categories: %w", err)
	}
	defer cursor.Close(ctx)

	var categories []CategoryProjection
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, fmt.Errorf("failed to decode categories: %w", err)
	}

	return categories, nil
}

func (ps *ProjectionStore) GetMovementByID(ctx context.Context, id string) (*MovementProjection, error) {
	var movement MovementProjection
	err := ps.movementsCollection.FindOne(ctx, bson.M{"_id": id, "is_deleted": false}).Decode(&movement)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find movement: %w", err)
	}

	return &movement, nil
}

func (ps *ProjectionStore) GetCategoryByID(ctx context.Context, id string) (*CategoryProjection, error) {
	var category CategoryProjection
	err := ps.categoriesCollection.FindOne(ctx, bson.M{"_id": id, "is_deleted": false}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find category: %w", err)
	}

	return &category, nil
}
