package queries

import (
	"context"
	"time"

	"escama/infrastructure/projections"
)

// ProjectionQueryHandler maneja consultas usando la base de datos de proyecciones
type ProjectionQueryHandler struct {
	projectionStore *projections.ProjectionStore
}

func NewProjectionQueryHandler(projectionStore *projections.ProjectionStore) *ProjectionQueryHandler {
	return &ProjectionQueryHandler{
		projectionStore: projectionStore,
	}
}

// GetMovements obtiene movimientos paginados desde las proyecciones
func (h *ProjectionQueryHandler) GetMovements(ctx context.Context, startDate, endDate *time.Time, limit, offset int) ([]Movement, int, error) {
	projectionMovements, total, err := h.projectionStore.GetMovements(ctx, startDate, endDate, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Convertir proyecciones a DTOs
	movements := make([]Movement, len(projectionMovements))
	for i, pm := range projectionMovements {
		movements[i] = Movement{
			ID:           pm.ID,
			Type:         pm.Type,
			CategoryID:   pm.CategoryID,
			CategoryName: pm.CategoryName,
			Amount:       pm.Amount,
			Description:  pm.Description,
			Date:         pm.Date,
			CreatedAt:    pm.CreatedAt,
		}
	}

	return movements, total, nil
}

// GetPaginatedMovements obtiene movimientos con metadatos de paginación
func (h *ProjectionQueryHandler) GetPaginatedMovements(ctx context.Context, query GetMovementsQuery) (PaginatedMovements, error) {
	movements, total, err := h.GetMovements(ctx, query.StartDate, query.EndDate, query.Limit, query.Offset)
	if err != nil {
		return PaginatedMovements{}, err
	}

	perPage := query.Limit
	if perPage <= 0 {
		perPage = 10
	}

	page := (query.Offset / perPage) + 1

	return PaginatedMovements{
		Movements: movements,
		Total:     total,
		Page:      page,
		PerPage:   perPage,
		HasNext:   (query.Offset + perPage) < total,
		HasPrev:   query.Offset > 0,
	}, nil
}

// GetBalance calcula el balance desde las proyecciones
func (h *ProjectionQueryHandler) GetBalance(ctx context.Context, query GetBalanceQuery) (Balance, error) {
	movements, _, err := h.GetMovements(ctx, &query.StartDate, &query.EndDate, 0, 0) // Sin paginación para el balance
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

// GetExpensesByCategory obtiene gastos agrupados por categoría desde las proyecciones
func (h *ProjectionQueryHandler) GetExpensesByCategory(ctx context.Context, query GetExpensesByCategoryQuery) ([]CategoryExpense, error) {
	movements, _, err := h.GetMovements(ctx, query.StartDate, query.EndDate, 0, 0) // Sin paginación
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

	// Ordenar por total descendente (bubble sort simple)
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Total < result[j].Total {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}

// GetCategories obtiene todas las categorías desde las proyecciones
func (h *ProjectionQueryHandler) GetCategories(ctx context.Context, query GetCategoriesQuery) ([]Category, error) {
	projectionCategories, err := h.projectionStore.GetCategories(ctx)
	if err != nil {
		return []Category{}, err
	}

	// Convertir proyecciones a DTOs
	categories := make([]Category, len(projectionCategories))
	for i, pc := range projectionCategories {
		categories[i] = Category{
			ID:        pc.ID,
			Name:      pc.Name,
			CreatedAt: pc.CreatedAt,
		}
	}

	return categories, nil
}

// GetMovementByID obtiene un movimiento específico por ID
func (h *ProjectionQueryHandler) GetMovementByID(ctx context.Context, id string) (*Movement, error) {
	projectionMovement, err := h.projectionStore.GetMovementByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if projectionMovement == nil {
		return nil, nil
	}

	return &Movement{
		ID:           projectionMovement.ID,
		Type:         projectionMovement.Type,
		CategoryID:   projectionMovement.CategoryID,
		CategoryName: projectionMovement.CategoryName,
		Amount:       projectionMovement.Amount,
		Description:  projectionMovement.Description,
		Date:         projectionMovement.Date,
		CreatedAt:    projectionMovement.CreatedAt,
	}, nil
}

// GetCategoryByID obtiene una categoría específica por ID
func (h *ProjectionQueryHandler) GetCategoryByID(ctx context.Context, id string) (*Category, error) {
	projectionCategory, err := h.projectionStore.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if projectionCategory == nil {
		return nil, nil
	}

	return &Category{
		ID:        projectionCategory.ID,
		Name:      projectionCategory.Name,
		CreatedAt: projectionCategory.CreatedAt,
	}, nil
}
