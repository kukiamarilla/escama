package queries

import (
	"context"
	"sort"
	"time"

	"escama/domain/events"
	"escama/infrastructure/eventstore"
)

// Movement representa un movimiento en el flujo de caja
type Movement struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"` // "income" o "expense"
	CategoryID   string    `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Amount       float64   `json:"amount"`
	Description  *string   `json:"description"`
	Date         time.Time `json:"date"`
	CreatedAt    time.Time `json:"created_at"`
}

// Balance representa el balance de un período
type Balance struct {
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	NetBalance   float64 `json:"net_balance"`
	Period       string  `json:"period"`
}

// GetMovementsQuery consulta para obtener movimientos con filtros de fecha
type GetMovementsQuery struct {
	StartDate *time.Time
	EndDate   *time.Time
}

// GetBalanceQuery consulta para obtener balance de un período
type GetBalanceQuery struct {
	StartDate time.Time
	EndDate   time.Time
}

// CategoryExpense representa el gasto total por categoría
type CategoryExpense struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Total        float64 `json:"total"`
	Count        int     `json:"count"`
}

// GetExpensesByCategoryQuery consulta para obtener gastos agrupados por categoría
type GetExpensesByCategoryQuery struct {
	StartDate *time.Time
	EndDate   *time.Time
}

// MovementsQueryHandler maneja consultas de movimientos
type MovementsQueryHandler struct {
	eventStore        eventstore.EventStore
	categoriesHandler *CategoriesQueryHandler
}

func NewMovementsQueryHandler(eventStore eventstore.EventStore) *MovementsQueryHandler {
	return &MovementsQueryHandler{
		eventStore:        eventStore,
		categoriesHandler: NewCategoriesQueryHandler(eventStore),
	}
}

func (h *MovementsQueryHandler) GetMovements(ctx context.Context, query GetMovementsQuery) ([]Movement, error) {
	storedEvents, err := h.eventStore.GetAllEvents(ctx, query.StartDate, query.EndDate)
	if err != nil {
		return []Movement{}, err
	}

	movements := h.eventsToMovements(storedEvents)
	if movements == nil {
		return []Movement{}, nil
	}

	// Obtener categorías para mapear nombres
	categories, err := h.categoriesHandler.GetCategories(ctx, GetCategoriesQuery{})
	if err != nil {
		return []Movement{}, err
	}

	// Crear un mapa de ID a nombre de categoría
	categoryNames := make(map[string]string)
	for _, category := range categories {
		categoryNames[category.ID] = category.Name
	}

	// Agregar nombres de categorías a los movimientos
	for i := range movements {
		if name, exists := categoryNames[movements[i].CategoryID]; exists {
			movements[i].CategoryName = name
		} else {
			movements[i].CategoryName = "Sin categoría"
		}
	}

	return movements, nil
}

func (h *MovementsQueryHandler) GetBalance(ctx context.Context, query GetBalanceQuery) (Balance, error) {
	movements, err := h.GetMovements(ctx, GetMovementsQuery{
		StartDate: &query.StartDate,
		EndDate:   &query.EndDate,
	})
	if err != nil {
		return Balance{}, err
	}

	var totalIncome, totalExpense float64
	for _, movement := range movements {
		if movement.Type == "income" {
			totalIncome += movement.Amount
		} else if movement.Type == "expense" {
			totalExpense += movement.Amount
		}
	}

	return Balance{
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
		NetBalance:   totalIncome - totalExpense,
		Period:       query.StartDate.Format("2006-01-02") + " - " + query.EndDate.Format("2006-01-02"),
	}, nil
}

func (h *MovementsQueryHandler) GetExpensesByCategory(ctx context.Context, query GetExpensesByCategoryQuery) ([]CategoryExpense, error) {
	movements, err := h.GetMovements(ctx, GetMovementsQuery{
		StartDate: query.StartDate,
		EndDate:   query.EndDate,
	})
	if err != nil {
		return []CategoryExpense{}, err
	}

	// Agrupar gastos por categoría
	categoryTotals := make(map[string]*CategoryExpense)

	for _, movement := range movements {
		if movement.Type == "expense" {
			categoryID := movement.CategoryID
			categoryName := movement.CategoryName

			if categoryID == "" {
				categoryID = "Sin categoría"
				categoryName = "Sin categoría"
			}

			if existing, exists := categoryTotals[categoryID]; exists {
				existing.Total += movement.Amount
				existing.Count++
			} else {
				categoryTotals[categoryID] = &CategoryExpense{
					CategoryID:   categoryID,
					CategoryName: categoryName,
					Total:        movement.Amount,
					Count:        1,
				}
			}
		}
	}

	// Convertir mapa a slice y ordenar por total descendente
	result := make([]CategoryExpense, 0, len(categoryTotals))
	for _, expense := range categoryTotals {
		result = append(result, *expense)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Total > result[j].Total
	})

	return result, nil
}

func (h *MovementsQueryHandler) getStringFromPayload(payload map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := payload[key]; ok {
			if strVal, ok := val.(string); ok {
				return strVal
			}
		}
	}
	return ""
}

func (h *MovementsQueryHandler) getFloat64FromPayload(payload map[string]interface{}, keys ...string) float64 {
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

func (h *MovementsQueryHandler) getStringPtrFromPayload(payload map[string]interface{}, keys ...string) *string {
	for _, key := range keys {
		if val, ok := payload[key]; ok {
			if strVal, ok := val.(string); ok && strVal != "" {
				return &strVal
			}
		}
	}
	return nil
}

func (h *MovementsQueryHandler) getTimeFromPayload(payload map[string]interface{}, keys ...string) time.Time {
	for _, key := range keys {
		if val, ok := payload[key]; ok {
			// Intentar diferentes formatos de fecha
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

// Helper para convertir eventos a movements
func (h *MovementsQueryHandler) eventsToMovements(storedEvents []events.StoredEvent) []Movement {
	movements := make([]Movement, 0, len(storedEvents))

	for _, storedEvent := range storedEvents {
		switch storedEvent.EventType {
		case "ExpenseCreated":
			movement := Movement{
				Type:      "expense",
				CreatedAt: storedEvent.OccurredAt,
			}

			// Probar tanto PascalCase como snake_case
			movement.ID = h.getStringFromPayload(storedEvent.Payload, "ExpenseID", "expense_id")
			movement.CategoryID = h.getStringFromPayload(storedEvent.Payload, "CategoryID", "category_id")
			movement.Amount = h.getFloat64FromPayload(storedEvent.Payload, "Amount", "amount")
			movement.Description = h.getStringPtrFromPayload(storedEvent.Payload, "Description", "description")
			movement.Date = h.getTimeFromPayload(storedEvent.Payload, "Date", "date")

			if movement.Date.IsZero() {
				movement.Date = storedEvent.OccurredAt
			}

			movements = append(movements, movement)

		case "IncomeCreated":
			movement := Movement{
				Type:      "income",
				CreatedAt: storedEvent.OccurredAt,
			}

			// Probar tanto PascalCase como snake_case
			movement.ID = h.getStringFromPayload(storedEvent.Payload, "IncomeID", "income_id")
			movement.CategoryID = h.getStringFromPayload(storedEvent.Payload, "CategoryID", "category_id")
			movement.Amount = h.getFloat64FromPayload(storedEvent.Payload, "Amount", "amount")
			movement.Description = h.getStringPtrFromPayload(storedEvent.Payload, "Description", "description")
			movement.Date = h.getTimeFromPayload(storedEvent.Payload, "Date", "date")

			if movement.Date.IsZero() {
				movement.Date = storedEvent.OccurredAt
			}

			movements = append(movements, movement)
		}
	}

	// Ordenar por fecha descendente
	sort.Slice(movements, func(i, j int) bool {
		return movements[i].Date.After(movements[j].Date)
	})

	return movements
}
