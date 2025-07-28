package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"escama/application"
	"escama/application/commands"
	"escama/infrastructure/eventbus"
	"escama/infrastructure/eventstore"
	"escama/infrastructure/repositories"
)

// Adapters para hacer bridge entre handlers específicos y CommandHandler interface
type categoryCommandAdapter struct {
	handler *commands.CreateCategoryHandler
}

func (a *categoryCommandAdapter) Handle(cmd application.Command) error {
	categoryCmd, ok := cmd.(commands.CreateCategoryCommand)
	if !ok {
		return fmt.Errorf("invalid command type for category handler")
	}
	return a.handler.Handle(context.Background(), categoryCmd)
}

type expenseCommandAdapter struct {
	handler *commands.CreateExpenseHandler
}

func (a *expenseCommandAdapter) Handle(cmd application.Command) error {
	expenseCmd, ok := cmd.(commands.CreateExpenseCommand)
	if !ok {
		return fmt.Errorf("invalid command type for expense handler")
	}
	return a.handler.Handle(context.Background(), expenseCmd)
}

func main() {
	fmt.Println("🚀 Event Sourcing Demo - Escama")
	fmt.Println("===============================")

	// Configurar infrastructure
	eventStore := eventstore.NewInMemoryEventStore()
	eventPublisher := eventbus.NewInMemoryEventPublisher()

	categoryRepo := repositories.NewCategoryRepository(eventStore)
	expenseRepo := repositories.NewExpenseRepository(eventStore)

	// Configurar application layer
	commandBus := application.NewCommandBus()

	// Registrar handlers con adapters
	createCategoryHandler := &commands.CreateCategoryHandler{
		Save:    categoryRepo.Save,
		Publish: eventPublisher.Publish,
	}
	commandBus.Register(commands.CreateCategoryCommand{}, &categoryCommandAdapter{handler: createCategoryHandler})

	createExpenseHandler := &commands.CreateExpenseHandler{
		Save:    expenseRepo.Save,
		Publish: eventPublisher.Publish,
	}
	commandBus.Register(commands.CreateExpenseCommand{}, &expenseCommandAdapter{handler: createExpenseHandler})

	// Demostrar Event Sourcing en acción
	fmt.Println("\n📝 Creating categories...")

	createCategoryCmd := commands.CreateCategoryCommand{
		Name: "Alimentación",
	}
	if err := commandBus.Dispatch(createCategoryCmd); err != nil {
		log.Fatalf("Error creating category: %v", err)
	}

	createCategoryCmd2 := commands.CreateCategoryCommand{
		Name: "Transporte",
	}
	if err := commandBus.Dispatch(createCategoryCmd2); err != nil {
		log.Fatalf("Error creating category: %v", err)
	}

	fmt.Println("\n💰 Creating expenses...")

	createExpenseCmd := commands.CreateExpenseCommand{
		CategoryID:  "some-category-id", // En una app real obtendrías el ID de la category creada
		Amount:      25.50,
		Description: stringPtr("Almuerzo en restaurante"),
		Date:        time.Now(),
	}
	if err := commandBus.Dispatch(createExpenseCmd); err != nil {
		log.Fatalf("Error creating expense: %v", err)
	}

	fmt.Println("\n✅ Event Sourcing demo completed!")
	fmt.Println("\nTu arquitectura Event Sourcing está funcionando:")
	fmt.Println("✓ Domain events generados por agregados")
	fmt.Println("✓ Command handlers procesando comandos")
	fmt.Println("✓ Events persistidos en Event Store")
	fmt.Println("✓ Events publicados a través de Event Publisher")
}

func stringPtr(s string) *string {
	return &s
}
