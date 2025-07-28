package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"escama/application"
	"escama/application/commands"
	"escama/application/queries"
	"escama/infrastructure/eventbus"
	"escama/infrastructure/eventstore"
	"escama/infrastructure/projections"
	"escama/infrastructure/repositories"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	eventStore             eventstore.EventStore
	commandBus             *application.CommandBus
	queryHandler           *queries.ProjectionQueryHandler
	categoriesQueryHandler *queries.CategoriesQueryHandler
	eventPublisher         *eventbus.InMemoryEventPublisher
	projectionStore        *projections.ProjectionStore
	categoryRepo           *repositories.CategoryRepository
	expenseRepo            *repositories.ExpenseRepository
	incomeRepo             *repositories.IncomeRepository
)

func init() {
	// Cargar variables de entorno desde .env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Configurar infrastructure - usar MongoDB
	mongoStore, err := eventstore.NewMongoEventStore()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	eventStore = mongoStore

	// Configurar cliente MongoDB para proyecciones
	connectionString := os.Getenv("MONGODB_CONNECTION_STRING")
	if connectionString == "" {
		connectionString = "mongodb://localhost:27017/escama"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB for projections: %v", err)
	}

	// Configurar proyecciones
	projectionStore = projections.NewProjectionStore(mongoClient, "escama_read")
	projectionSubscriber := eventbus.NewProjectionSubscriber(projectionStore)

	eventPublisher = eventbus.NewInMemoryEventPublisher()
	eventPublisher.SetProjectionSubscriber(projectionSubscriber)

	categoryRepo = repositories.NewCategoryRepository(eventStore)
	expenseRepo = repositories.NewExpenseRepository(eventStore)
	incomeRepo = repositories.NewIncomeRepository(eventStore)

	// Usar proyecciones para queries (más rápido)
	queryHandler = queries.NewProjectionQueryHandler(projectionStore)
	categoriesQueryHandler = queries.NewCategoriesQueryHandler(eventStore) // Mantenemos este por ahora

	// Configurar command bus
	commandBus = application.NewCommandBus()

	// Registrar handlers
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

	createIncomeHandler := &commands.CreateIncomeHandler{
		Save:    incomeRepo.Save,
		Publish: eventPublisher.Publish,
	}
	commandBus.Register(commands.CreateIncomeCommand{}, &incomeCommandAdapter{handler: createIncomeHandler})

	// Registrar handlers de actualización
	updateExpenseHandler := &commands.UpdateExpenseHandler{
		Repository: expenseRepo,
		Publish:    eventPublisher.Publish,
	}
	commandBus.Register(commands.UpdateExpenseCommand{}, &updateExpenseCommandAdapter{handler: updateExpenseHandler})

	updateIncomeHandler := &commands.UpdateIncomeHandler{
		Repository: incomeRepo,
		Publish:    eventPublisher.Publish,
	}
	commandBus.Register(commands.UpdateIncomeCommand{}, &updateIncomeCommandAdapter{handler: updateIncomeHandler})

	// Registrar handlers de eliminación
	deleteExpenseHandler := &commands.DeleteExpenseHandler{
		Repository: expenseRepo,
		Publish:    eventPublisher.Publish,
	}
	commandBus.Register(commands.DeleteExpenseCommand{}, &deleteExpenseCommandAdapter{handler: deleteExpenseHandler})

	deleteIncomeHandler := &commands.DeleteIncomeHandler{
		Repository: incomeRepo,
		Publish:    eventPublisher.Publish,
	}
	commandBus.Register(commands.DeleteIncomeCommand{}, &deleteIncomeCommandAdapter{handler: deleteIncomeHandler})
}

var rootCmd = &cobra.Command{
	Use:   "escama",
	Short: "Gestor de finanzas personales con Event Sourcing",
	Long:  "Una aplicación CLI para gestionar ingresos y gastos usando Event Sourcing con MongoDB",
}

var categoryCmd = &cobra.Command{
	Use:   "category",
	Short: "Gestión de categorías",
}

var createCategoryCmd = &cobra.Command{
	Use:   "create [nombre]",
	Short: "Crear una nueva categoría",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		categoryName := args[0]

		createCmd := commands.CreateCategoryCommand{
			Name: categoryName,
		}

		if err := commandBus.Dispatch(createCmd); err != nil {
			log.Fatalf("Error creating category: %v", err)
		}

		fmt.Printf("✅ Categoría '%s' creada exitosamente\n", categoryName)
	},
}

var expenseCmd = &cobra.Command{
	Use:   "expense",
	Short: "Gestión de gastos",
}

var createExpenseCmd = &cobra.Command{
	Use:   "create [monto] [descripcion] [--category nombre-categoria]",
	Short: "Registrar un nuevo gasto",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		amountStr := args[0]

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			log.Fatalf("Monto inválido: %v", err)
		}

		var description *string
		if len(args) > 1 {
			desc := args[1]
			description = &desc
		}

		// Obtener categoría desde flag o selector interactivo
		categoryFlag, _ := cmd.Flags().GetString("category")
		var categoryID string

		if categoryFlag != "" {
			// Intentar buscar por nombre primero
			if foundID, err := findCategoryByName(categoryFlag); err == nil {
				categoryID = foundID
			} else {
				log.Fatalf("Error: %v", err)
			}
		} else {
			selectedCategory, err := selectCategory()
			if err != nil {
				log.Fatalf("Error al seleccionar categoría: %v", err)
			}
			categoryID = selectedCategory
		}

		// Obtener fecha desde flag o usar fecha actual
		dateStr, _ := cmd.Flags().GetString("date")
		var movementDate time.Time

		if dateStr != "" {
			parsedDate, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				log.Fatalf("Fecha inválida. Use formato YYYY-MM-DD: %v", err)
			}
			movementDate = parsedDate
		} else {
			movementDate = time.Now()
		}

		createCmd := commands.CreateExpenseCommand{
			CategoryID:  categoryID,
			Amount:      amount,
			Description: description,
			Date:        movementDate,
		}

		if err := commandBus.Dispatch(createCmd); err != nil {
			log.Fatalf("Error creating expense: %v", err)
		}

		dateDisplay := movementDate.Format("2006-01-02")
		fmt.Printf("💸 Gasto de ₲%.0f registrado exitosamente para el %s\n", amount, dateDisplay)
	},
}

var incomeCmd = &cobra.Command{
	Use:   "income",
	Short: "Gestión de ingresos",
}

var createIncomeCmd = &cobra.Command{
	Use:   "create [monto] [descripcion] [--category nombre-categoria]",
	Short: "Registrar un nuevo ingreso",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		amountStr := args[0]

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			log.Fatalf("Monto inválido: %v", err)
		}

		var description *string
		if len(args) > 1 {
			desc := args[1]
			description = &desc
		}

		// Obtener categoría desde flag o selector interactivo
		categoryFlag, _ := cmd.Flags().GetString("category")
		var categoryID string

		if categoryFlag != "" {
			// Intentar buscar por nombre primero
			if foundID, err := findCategoryByName(categoryFlag); err == nil {
				categoryID = foundID
			} else {
				log.Fatalf("Error: %v", err)
			}
		} else {
			selectedCategory, err := selectCategory()
			if err != nil {
				log.Fatalf("Error al seleccionar categoría: %v", err)
			}
			categoryID = selectedCategory
		}

		// Obtener fecha desde flag o usar fecha actual
		dateStr, _ := cmd.Flags().GetString("date")
		var movementDate time.Time

		if dateStr != "" {
			parsedDate, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				log.Fatalf("Fecha inválida. Use formato YYYY-MM-DD: %v", err)
			}
			movementDate = parsedDate
		} else {
			movementDate = time.Now()
		}

		createCmd := commands.CreateIncomeCommand{
			CategoryID:  categoryID,
			Amount:      amount,
			Description: description,
			Date:        movementDate,
		}

		if err := commandBus.Dispatch(createCmd); err != nil {
			log.Fatalf("Error creating income: %v", err)
		}

		dateDisplay := movementDate.Format("2006-01-02")
		fmt.Printf("💰 Ingreso de ₲%.0f registrado exitosamente para el %s\n", amount, dateDisplay)
	},
}

// Comando para actualizar gastos
var updateExpenseCmd = &cobra.Command{
	Use:   "update [id] [monto] [descripcion] [--category nombre-categoria]",
	Short: "Actualizar un gasto existente",
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		expenseID := args[0]
		amountStr := args[1]
		description := args[2]

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			log.Fatalf("Monto inválido: %v", err)
		}

		// Obtener categoría desde flag
		categoryFlag, _ := cmd.Flags().GetString("category")
		var categoryID string

		if categoryFlag != "" {
			// Buscar por nombre
			if foundID, err := findCategoryByName(categoryFlag); err == nil {
				categoryID = foundID
			} else {
				log.Fatalf("Error: %v", err)
			}
		} else {
			selectedCategory, err := selectCategory()
			if err != nil {
				log.Fatalf("Error al seleccionar categoría: %v", err)
			}
			categoryID = selectedCategory
		}

		// Obtener fecha desde flag o usar fecha actual
		dateStr, _ := cmd.Flags().GetString("date")
		var movementDate time.Time

		if dateStr != "" {
			parsedDate, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				log.Fatalf("Fecha inválida. Use formato YYYY-MM-DD: %v", err)
			}
			movementDate = parsedDate
		} else {
			movementDate = time.Now()
		}

		updateCmd := commands.UpdateExpenseCommand{
			ID:          expenseID,
			CategoryID:  categoryID,
			Amount:      amount,
			Description: &description,
			Date:        movementDate,
		}

		if err := commandBus.Dispatch(updateCmd); err != nil {
			log.Fatalf("Error updating expense: %v", err)
		}

		dateDisplay := movementDate.Format("2006-01-02")
		fmt.Printf("💸 Gasto actualizado: ₲%.0f para el %s\n", amount, dateDisplay)
	},
}

// Comando para actualizar ingresos
var updateIncomeCmd = &cobra.Command{
	Use:   "update [id] [monto] [descripcion] [--category nombre-categoria]",
	Short: "Actualizar un ingreso existente",
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		incomeID := args[0]
		amountStr := args[1]
		description := args[2]

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			log.Fatalf("Monto inválido: %v", err)
		}

		// Obtener categoría desde flag
		categoryFlag, _ := cmd.Flags().GetString("category")
		var categoryID string

		if categoryFlag != "" {
			// Buscar por nombre
			if foundID, err := findCategoryByName(categoryFlag); err == nil {
				categoryID = foundID
			} else {
				log.Fatalf("Error: %v", err)
			}
		} else {
			selectedCategory, err := selectCategory()
			if err != nil {
				log.Fatalf("Error al seleccionar categoría: %v", err)
			}
			categoryID = selectedCategory
		}

		// Obtener fecha desde flag o usar fecha actual
		dateStr, _ := cmd.Flags().GetString("date")
		var movementDate time.Time

		if dateStr != "" {
			parsedDate, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				log.Fatalf("Fecha inválida. Use formato YYYY-MM-DD: %v", err)
			}
			movementDate = parsedDate
		} else {
			movementDate = time.Now()
		}

		updateCmd := commands.UpdateIncomeCommand{
			ID:          incomeID,
			CategoryID:  categoryID,
			Amount:      amount,
			Description: &description,
			Date:        movementDate,
		}

		if err := commandBus.Dispatch(updateCmd); err != nil {
			log.Fatalf("Error updating income: %v", err)
		}

		dateDisplay := movementDate.Format("2006-01-02")
		fmt.Printf("💰 Ingreso actualizado: ₲%.0f para el %s\n", amount, dateDisplay)
	},
}

// Comando para eliminar gastos
var deleteExpenseCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Eliminar un gasto existente",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		expenseID := args[0]

		// Confirmar eliminación
		fmt.Printf("⚠️  ¿Estás seguro de que deseas eliminar el gasto %s? (y/N): ", expenseID)
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error al leer input: %v", err)
		}

		input = strings.TrimSpace(strings.ToLower(input))
		if input != "y" && input != "yes" && input != "sí" && input != "si" {
			fmt.Println("❌ Operación cancelada")
			return
		}

		deleteCmd := commands.DeleteExpenseCommand{
			ID: expenseID,
		}

		if err := commandBus.Dispatch(deleteCmd); err != nil {
			log.Fatalf("Error deleting expense: %v", err)
		}

		fmt.Printf("💸 Gasto %s eliminado exitosamente\n", expenseID)
	},
}

// Comando para eliminar ingresos
var deleteIncomeCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Eliminar un ingreso existente",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		incomeID := args[0]

		// Confirmar eliminación
		fmt.Printf("⚠️  ¿Estás seguro de que deseas eliminar el ingreso %s? (y/N): ", incomeID)
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error al leer input: %v", err)
		}

		input = strings.TrimSpace(strings.ToLower(input))
		if input != "y" && input != "yes" && input != "sí" && input != "si" {
			fmt.Println("❌ Operación cancelada")
			return
		}

		deleteCmd := commands.DeleteIncomeCommand{
			ID: incomeID,
		}

		if err := commandBus.Dispatch(deleteCmd); err != nil {
			log.Fatalf("Error deleting income: %v", err)
		}

		fmt.Printf("💰 Ingreso %s eliminado exitosamente\n", incomeID)
	},
}

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Ver balance actual",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Balance del mes actual
		now := time.Now()
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endOfMonth := startOfMonth.AddDate(0, 1, -1)

		balance, err := queryHandler.GetBalance(ctx, queries.GetBalanceQuery{
			StartDate: startOfMonth,
			EndDate:   endOfMonth,
		})
		if err != nil {
			log.Fatalf("Error getting balance: %v", err)
		}

		fmt.Printf("\n📊 Balance del mes (%s)\n", balance.Period)
		fmt.Printf("════════════════════════════════════\n")
		fmt.Printf("💰 Total Ingresos:  ₲%.0f\n", balance.TotalIncome)
		fmt.Printf("💸 Total Gastos:    ₲%.0f\n", balance.TotalExpense)
		fmt.Printf("📈 Balance Neto:    ₲%.0f\n", balance.NetBalance)

		if balance.NetBalance > 0 {
			fmt.Printf("✅ ¡Felicitaciones! Tienes un balance positivo\n")
		} else if balance.NetBalance < 0 {
			fmt.Printf("⚠️  Cuidado, tienes un balance negativo\n")
		} else {
			fmt.Printf("⚖️  Balance equilibrado\n")
		}
	},
}

var movementsCmd = &cobra.Command{
	Use:   "movements",
	Short: "Ver movimientos recientes",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		paginatedResult, err := queryHandler.GetPaginatedMovements(ctx, queries.GetMovementsQuery{})
		if err != nil {
			log.Fatalf("Error getting movements: %v", err)
		}

		movements := paginatedResult.Movements

		if len(movements) == 0 {
			fmt.Println("📝 No hay movimientos registrados")
			return
		}

		fmt.Printf("\n📋 Movimientos recientes (%d)\n", len(movements))
		fmt.Printf("════════════════════════════════════════════════════════════\n")

		for _, movement := range movements {
			typeIcon := "💸"
			if movement.Type == "income" {
				typeIcon = "💰"
			}

			desc := "Sin descripción"
			if movement.Description != nil {
				desc = *movement.Description
			}

			fmt.Printf("%s %s - ₲%.0f - %s - %s\n",
				typeIcon,
				movement.Date.Format("2006-01-02"),
				movement.Amount,
				desc,
				movement.CategoryID)
		}
	},
}

// findCategoryByName busca una categoría por su nombre y devuelve su ID
func findCategoryByName(categoryName string) (string, error) {
	ctx := context.Background()
	categories, err := categoriesQueryHandler.GetCategories(ctx, queries.GetCategoriesQuery{})
	if err != nil {
		return "", fmt.Errorf("error al obtener categorías: %w", err)
	}

	// Buscar por coincidencia exacta (sin importar mayúsculas/minúsculas)
	for _, category := range categories {
		if strings.EqualFold(category.Name, categoryName) {
			return category.ID, nil
		}
	}

	// Si no hay coincidencia exacta, mostrar categorías disponibles
	fmt.Printf("❌ Categoría '%s' no encontrada.\n", categoryName)
	fmt.Println("\n📋 Categorías disponibles:")
	for _, category := range categories {
		fmt.Printf("  • %s\n", category.Name)
	}

	return "", fmt.Errorf("categoría '%s' no encontrada", categoryName)
}

// selectCategory muestra un selector interactivo de categorías existentes
func selectCategory() (string, error) {
	ctx := context.Background()
	categories, err := categoriesQueryHandler.GetCategories(ctx, queries.GetCategoriesQuery{})
	if err != nil {
		return "", fmt.Errorf("error al obtener categorías: %w", err)
	}

	if len(categories) == 0 {
		fmt.Println("❌ No hay categorías disponibles.")
		fmt.Println("💡 Crea una categoría primero con: escama-cli category create [nombre]")
		return "", fmt.Errorf("no hay categorías disponibles")
	}

	fmt.Println("\n📋 Categorías disponibles:")
	for i, category := range categories {
		fmt.Printf("  %d. %s\n", i+1, category.Name)
	}

	fmt.Print("\n🎯 Selecciona una categoría (número): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error al leer input: %w", err)
	}

	input = strings.TrimSpace(input)
	selection, err := strconv.Atoi(input)
	if err != nil || selection < 1 || selection > len(categories) {
		return "", fmt.Errorf("selección inválida. Debe ser un número entre 1 y %d", len(categories))
	}

	selectedCategory := categories[selection-1]
	fmt.Printf("✅ Categoría seleccionada: %s\n", selectedCategory.Name)
	return selectedCategory.ID, nil
}

// Adapters (similar a main.go)
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

type incomeCommandAdapter struct {
	handler *commands.CreateIncomeHandler
}

func (a *incomeCommandAdapter) Handle(cmd application.Command) error {
	incomeCmd, ok := cmd.(commands.CreateIncomeCommand)
	if !ok {
		return fmt.Errorf("invalid command type for income handler")
	}
	return a.handler.Handle(context.Background(), incomeCmd)
}

// Adaptadores para comandos de actualización
type updateExpenseCommandAdapter struct {
	handler *commands.UpdateExpenseHandler
}

func (a *updateExpenseCommandAdapter) Handle(cmd application.Command) error {
	updateCmd, ok := cmd.(commands.UpdateExpenseCommand)
	if !ok {
		return fmt.Errorf("invalid command type for update expense handler")
	}
	return a.handler.Handle(context.Background(), updateCmd)
}

type updateIncomeCommandAdapter struct {
	handler *commands.UpdateIncomeHandler
}

func (a *updateIncomeCommandAdapter) Handle(cmd application.Command) error {
	updateCmd, ok := cmd.(commands.UpdateIncomeCommand)
	if !ok {
		return fmt.Errorf("invalid command type for update income handler")
	}
	return a.handler.Handle(context.Background(), updateCmd)
}

// Adaptadores para comandos de eliminación
type deleteExpenseCommandAdapter struct {
	handler *commands.DeleteExpenseHandler
}

func (a *deleteExpenseCommandAdapter) Handle(cmd application.Command) error {
	deleteCmd, ok := cmd.(commands.DeleteExpenseCommand)
	if !ok {
		return fmt.Errorf("invalid command type for delete expense handler")
	}
	return a.handler.Handle(context.Background(), deleteCmd)
}

type deleteIncomeCommandAdapter struct {
	handler *commands.DeleteIncomeHandler
}

func (a *deleteIncomeCommandAdapter) Handle(cmd application.Command) error {
	deleteCmd, ok := cmd.(commands.DeleteIncomeCommand)
	if !ok {
		return fmt.Errorf("invalid command type for delete income handler")
	}
	return a.handler.Handle(context.Background(), deleteCmd)
}

func main() {
	// Agregar flags de fecha a los comandos
	createExpenseCmd.Flags().StringP("date", "t", "", "Fecha del gasto (formato: YYYY-MM-DD). Si no se especifica, usa la fecha actual")
	createIncomeCmd.Flags().StringP("date", "t", "", "Fecha del ingreso (formato: YYYY-MM-DD). Si no se especifica, usa la fecha actual")
	createExpenseCmd.Flags().StringP("category", "c", "", "Nombre de la categoría para el gasto (si no se especifica, se pedirá interactivamente)")
	createIncomeCmd.Flags().StringP("category", "c", "", "Nombre de la categoría para el ingreso (si no se especifica, se pedirá interactivamente)")

	// Agregar flags a comandos de actualización
	updateExpenseCmd.Flags().StringP("date", "t", "", "Fecha del gasto (formato: YYYY-MM-DD). Si no se especifica, usa la fecha actual")
	updateIncomeCmd.Flags().StringP("date", "t", "", "Fecha del ingreso (formato: YYYY-MM-DD). Si no se especifica, usa la fecha actual")
	updateExpenseCmd.Flags().StringP("category", "c", "", "Nombre de la categoría para el gasto (si no se especifica, se pedirá interactivamente)")
	updateIncomeCmd.Flags().StringP("category", "c", "", "Nombre de la categoría para el ingreso (si no se especifica, se pedirá interactivamente)")

	// Agregar subcomandos
	categoryCmd.AddCommand(createCategoryCmd)
	expenseCmd.AddCommand(createExpenseCmd)
	expenseCmd.AddCommand(updateExpenseCmd)
	expenseCmd.AddCommand(deleteExpenseCmd)
	incomeCmd.AddCommand(createIncomeCmd)
	incomeCmd.AddCommand(updateIncomeCmd)
	incomeCmd.AddCommand(deleteIncomeCmd)

	rootCmd.AddCommand(categoryCmd)
	rootCmd.AddCommand(expenseCmd)
	rootCmd.AddCommand(incomeCmd)
	rootCmd.AddCommand(balanceCmd)
	rootCmd.AddCommand(movementsCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Cerrar conexión MongoDB al terminar
	if mongoStore, ok := eventStore.(*eventstore.MongoEventStore); ok {
		mongoStore.Close()
	}
}
