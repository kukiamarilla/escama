package queries

import (
	"context"
	"sort"
	"time"

	"escama/domain/events"
	"escama/infrastructure/eventstore"
)

// Category representa una categoría en el sistema
type Category struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// GetCategoriesQuery consulta para obtener todas las categorías
type GetCategoriesQuery struct{}

// CategoriesQueryHandler maneja consultas de categorías
type CategoriesQueryHandler struct {
	eventStore eventstore.EventStore
}

func NewCategoriesQueryHandler(eventStore eventstore.EventStore) *CategoriesQueryHandler {
	return &CategoriesQueryHandler{eventStore: eventStore}
}

func (h *CategoriesQueryHandler) GetCategories(ctx context.Context, query GetCategoriesQuery) ([]Category, error) {
	storedEvents, err := h.eventStore.GetAllEvents(ctx, nil, nil)
	if err != nil {
		return []Category{}, err
	}

	categories := h.eventsToCategories(storedEvents)
	if categories == nil {
		return []Category{}, nil
	}
	return categories, nil
}

// Helper para convertir eventos a categorías
func (h *CategoriesQueryHandler) eventsToCategories(storedEvents []events.StoredEvent) []Category {
	categoryMap := make(map[string]Category)

	for _, storedEvent := range storedEvents {
		if storedEvent.EventType == "CategoryCreated" {
			categoryID := h.getStringFromPayload(storedEvent.Payload, "CategoryID", "category_id")
			categoryName := h.getStringFromPayload(storedEvent.Payload, "Name", "name")

			if categoryID != "" && categoryName != "" {
				categoryMap[categoryID] = Category{
					ID:        categoryID,
					Name:      categoryName,
					CreatedAt: storedEvent.OccurredAt,
				}
			}
		}
	}

	// Convertir mapa a slice y ordenar por nombre
	categories := make([]Category, 0, len(categoryMap))
	for _, category := range categoryMap {
		categories = append(categories, category)
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	return categories
}

func (h *CategoriesQueryHandler) getStringFromPayload(payload map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := payload[key]; ok {
			if strVal, ok := val.(string); ok {
				return strVal
			}
		}
	}
	return ""
}
