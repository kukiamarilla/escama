# Escama - Gestor de Finanzas Personales con Event Sourcing

Sistema completo de finanzas personales implementado con **Event Sourcing** en Go, usando MongoDB para persistencia. **Configurado para usar Guaraníes (₲) paraguayos**.

## 🏗️ Arquitectura

### Clean Architecture + Event Sourcing
- **Domain Layer**: Agregados (Category, Expense, Income) con eventos de dominio
- **Application Layer**: Commands, Queries y Bus de comandos
- **Infrastructure Layer**: Event Store (MongoDB), Repositories, Event Publisher

### Tecnologías
- **Backend**: Go con Gorilla Mux
- **Base de datos**: MongoDB con Event Store
- **Frontend**: HTML/CSS/JavaScript vanilla
- **CLI**: Cobra para gestión por línea de comandos

## 🚀 Funcionalidades

### CLI (`escama-cli`)
```bash
# Gestión de categorías
./escama-cli category create "Alimentación"
./escama-cli category create "Trabajo"

# Registro de ingresos (en Guaraníes)
./escama-cli income create "trabajo-id" 3500000 "Salario enero"
./escama-cli income create "freelance" 850000 "Proyecto web"

# Registro de ingresos con fecha específica
./escama-cli income create "trabajo" 500000 "Proyecto del 20" -t 2025-07-20
./escama-cli income create "freelance" 300000 "Consultoría" --date 2025-07-15

# Registro de gastos (en Guaraníes)
./escama-cli expense create "alimentacion-id" 120000 "Supermercado"
./escama-cli expense create "transporte" 25000 "Combustible"

# Registro de gastos con fecha específica
./escama-cli expense create "alimentacion" 45000 "Almuerzo del lunes" -t 2025-07-21
./escama-cli expense create "transporte" 8000 "Taxi" --date 2025-07-23

# Ver balance del mes
./escama-cli balance

# Ver movimientos recientes
./escama-cli movements
```

### Dashboard Web (`escama-server`)
- **URL**: http://localhost:8080
- **Listado de movimientos** con filtros de fecha
- **Balance del período** (ingresos vs gastos) en ₲ Guaraníes
- **Gráfico de gastos por categoría** (dona interactiva con Chart.js)
- **Controles de fecha** para analizar períodos específicos
- **API REST** en `/api/movements`, `/api/balance` y `/api/expenses-by-category`

## 📦 Estructura del Proyecto

```
escama/
├── domain/                     # Capa de dominio
│   ├── category.go            # Agregado Category
│   ├── expense.go             # Agregado Expense  
│   ├── income.go              # Agregado Income
│   └── events/                # Eventos de dominio
│       ├── base.go           # Interfaces base
│       ├── category_created.go
│       ├── expense_created.go
│       └── income_created.go
├── application/               # Capa de aplicación
│   ├── bus.go                # Command/Query Bus
│   ├── commands/             # Command handlers
│   │   ├── create_category.go
│   │   ├── create_expense.go
│   │   └── create_income.go
│   └── queries/              # Query handlers
│       └── movements.go
├── infrastructure/           # Capa de infraestructura
│   ├── eventstore/          # Event Store
│   │   ├── eventstore.go    # Interface + InMemory
│   │   └── mongodb.go       # Implementación MongoDB
│   ├── repositories/        # Repositories
│   │   ├── category.go
│   │   ├── expense.go
│   │   └── income.go
│   └── eventbus/            # Event Publisher
│       └── publisher.go
├── cmd/                     # Aplicaciones
│   ├── cli/main.go         # CLI application
│   └── server/main.go      # Web server
├── web/                    # Frontend
│   └── index.html         # Dashboard SPA
├── go.mod                 # Dependencias Go
└── .env                   # Variables de entorno
```

## 🛠️ Configuración y Uso

### Requisitos
- Go 1.24+
- MongoDB (local o cloud)

### Variables de Entorno (.env)
```bash
MONGODB_CONNECTION_STRING=mongodb://localhost:27017/escama
```

### Compilación
```bash
# CLI
go build -o escama-cli ./cmd/cli

# Servidor web
go build -o escama-server ./cmd/server
```

### Ejecución
```bash
# Iniciar servidor web
./escama-server

# Usar CLI
./escama-cli --help
```

## 🎯 Patrones Implementados

### Event Sourcing
- ✅ **Eventos como fuente de verdad**
- ✅ **Agregados que generan eventos**
- ✅ **Event Store persistente** (MongoDB)
- ✅ **Reconstrucción de estado desde eventos**
- ✅ **Command/Query Separation**

### Clean Architecture  
- ✅ **Separación de capas**
- ✅ **Inversión de dependencias**
- ✅ **Domain-driven design**
- ✅ **Repository pattern**

### CQRS (Command Query Responsibility Segregation)
- ✅ **Commands para escritura**
- ✅ **Queries para lectura**
- ✅ **Bus de comandos**
- ✅ **Handlers especializados**

## 💰 Formato Monetario

El sistema está configurado para **Guaraníes paraguayos (₲)** sin decimales:
- **CLI**: Muestra montos como `₲850000`, `₲25000`
- **Dashboard Web**: Formatea con separadores de miles `₲850.000`, `₲25.000`
- **API**: Devuelve números en formato JSON estándar
- **Base de datos**: Almacena como números (float64) para cálculos precisos

## 🎨 Características del Dashboard

- **Responsive design** para móviles y desktop
- **Filtros de fecha** flexibles que actualizan todos los datos
- **Balance en tiempo real** con códigos de color
- **Gráfico interactivo** de gastos por categoría (Chart.js)
- **Lista de movimientos** ordenada cronológicamente
- **Formato paraguayo** con separadores de miles
- **API REST** documentada

## 🔧 Extensibilidad

El sistema está diseñado para ser fácilmente extensible:

- **Nuevos tipos de movimientos**: Agregar nuevos agregados y eventos
- **Nuevos Event Stores**: Implementar interface EventStore (PostgreSQL, etc.)
- **Nuevas vistas**: Agregar queries personalizadas
- **Integraciones**: Event Publisher puede notificar sistemas externos

## 📊 APIs Disponibles

### GET /api/balance
```json
{
  "total_income": 856300.00,
  "total_expense": 290166.25, 
  "net_balance": 566133.75,
  "period": "2025-07-01 - 2025-07-31"
}
```

### GET /api/movements
```json
[
  {
    "id": "movement-id",
    "type": "income",
    "category_id": "freelance", 
    "amount": 850000.00,
    "description": "Proyecto web",
    "date": "2025-07-27T17:05:55Z",
    "created_at": "2025-07-27T20:05:55Z"
  }
]
```

### GET /api/expenses-by-category
```json
[
  {
    "category_id": "salud",
    "total": 150000.00,
    "count": 1
  },
  {
    "category_id": "alimentacion",
    "total": 80000.00,
    "count": 1
  },
  {
    "category_id": "entretenimiento", 
    "total": 35000.00,
    "count": 1
  }
]
```

## 💡 Ejemplos de Uso

### Salario mensual típico
```bash
./escama-cli income create "trabajo" 3500000 "Salario desarrollador"
# Output: 💰 Ingreso de ₲3500000 registrado exitosamente para el 2025-07-27
```

### Gastos cotidianos
```bash
./escama-cli expense create "alimentacion" 80000 "Almuerzo semanal"
./escama-cli expense create "transporte" 15000 "Colectivo diario"
# Output: 💸 Gasto de ₲80000 registrado exitosamente para el 2025-07-27
```

### Registros con fecha específica
```bash
# Registrar gasto de ayer
./escama-cli expense create "alimentacion" 65000 "Cena familiar" -t 2025-07-26

# Registrar ingreso del mes pasado (formato largo)
./escama-cli income create "freelance" 800000 "Proyecto junio" --date 2025-06-30

# Registrar múltiples transacciones históricas
./escama-cli expense create "transporte" 12000 "Taxi del lunes" -t 2025-07-21
./escama-cli expense create "farmacia" 35000 "Medicamentos" -t 2025-07-22
```

### Balance del mes
```bash
./escama-cli balance
# Output:
# 📊 Balance del mes (2025-07-01 - 2025-07-31)
# ════════════════════════════════════
# 💰 Total Ingresos:  ₲856300
# 💸 Total Gastos:    ₲25166
# 📈 Balance Neto:    ₲831134
# ✅ ¡Felicitaciones! Tienes un balance positivo
```

## 📅 Gestión de Fechas

### Opciones de Fecha
Los comandos `income create` y `expense create` soportan fecha opcional:

- **Flag corto**: `-t YYYY-MM-DD`
- **Flag largo**: `--date YYYY-MM-DD`
- **Por defecto**: Si no se especifica, usa la fecha actual

### Casos de Uso
- **Registro histórico**: Agregar gastos/ingresos de días/meses anteriores
- **Planificación**: Registrar transacciones futuras programadas
- **Corrección**: Registrar movimientos en la fecha correcta
- **Migración**: Importar datos históricos de otros sistemas

### Validación
- Formato requerido: `YYYY-MM-DD` (ISO 8601)
- Fechas inválidas muestran error explicativo
- Compatible con filtros de fecha del dashboard web

---

**¡Tu sistema de Event Sourcing está listo para usar en Paraguay!** 🇵🇾

- CLI para registro rápido de movimientos en Guaraníes
- Dashboard web para análisis y visualización
- MongoDB para persistencia robusta de eventos
- Arquitectura escalable y mantenible 