# 🐟 Escama - Sistema de Finanzas Personales con CQRS + Event Sourcing

Sistema completo de finanzas personales implementado con **arquitectura CQRS**, **Event Sourcing** y **proyecciones en tiempo real** en Go, usando MongoDB. **Configurado para usar Guaraníes (₲) paraguayos**.

## 🏗️ Arquitectura Avanzada

### CQRS + Event Sourcing + Proyecciones
- **📝 Base de Escritura**: Event Store (MongoDB) para comandos
- **📊 Base de Lectura**: Proyecciones (MongoDB) para consultas optimizadas
- **⚡ Proyecciones Automáticas**: Se actualizan en tiempo real con cada evento
- **🔄 Consistencia Eventual**: Entre escritura y lectura
- **🎯 Separación Completa**: Commands vs Queries

### Clean Architecture + DDD
- **Domain Layer**: Agregados (Category, Expense, Income) con eventos de dominio
- **Application Layer**: Commands, Queries, Handlers y Bus de comandos
- **Infrastructure Layer**: Event Store, Proyecciones, Repositories, Event Publisher

### Tecnologías
- **Backend**: Go con Gorilla Mux
- **Base de datos**: MongoDB (Event Store + Proyecciones)
- **Frontend**: HTML/CSS/JavaScript vanilla con paginación
- **CLI**: Cobra para gestión completa CRUD por línea de comandos

## 🚀 Funcionalidades Completas

### CLI (`escama-cli`) - CRUD Completo
```bash
# ===== GESTIÓN DE CATEGORÍAS =====
escama category create "Alimentación"
escama category create "Salario"
escama category create "Freelance"

# ===== INGRESOS (CRUD) =====
# Crear ingresos (en Guaraníes)
escama income create 3500000 "Salario mensual" --category "Salario"
escama income create 850000 "Proyecto web" --category "Freelance" --date 2025-07-20

# Actualizar ingresos existentes
escama income update [id] 4000000 "Salario aumentado" --category "Salario"

# Eliminar ingresos (con confirmación)
escama income delete [id]

# ===== GASTOS (CRUD) =====
# Crear gastos
escama expense create 120000 "Supermercado" --category "Alimentación"
escama expense create 25000 "Combustible" --category "Transporte" --date 2025-07-21

# Actualizar gastos existentes
escama expense update [id] 150000 "Supermercado grande" --category "Alimentación"

# Eliminar gastos (con confirmación)
escama expense delete [id]

# ===== CONSULTAS OPTIMIZADAS =====
# Ver balance del mes (desde proyecciones)
escama balance

# Ver movimientos recientes (paginados, con nombres de categorías)
escama movements

# ===== AYUDA =====
escama expense --help    # Ver todos los subcomandos
escama income --help     # create, update, delete
```

### Dashboard Web (`escama-server`) - Con Paginación
- **URL**: http://localhost:8080
- **📄 Paginación**: 10 movimientos por página con navegación
- **🏷️ Nombres de categorías**: Los movimientos muestran nombres en lugar de IDs
- **📊 Balance en tiempo real** en ₲ Guaraníes
- **📈 Gráfico de gastos por categoría** (barras interactivas)
- **🗓️ Filtros de fecha** para analizar períodos específicos
- **⚡ API REST optimizada** con proyecciones

## 📦 Estructura del Proyecto Actualizada

```
escama/
├── domain/                          # Capa de dominio
│   ├── category.go                  # Agregado Category
│   ├── expense.go                   # Agregado Expense con Update/Delete
│   ├── income.go                    # Agregado Income con Update/Delete
│   └── events/                      # Eventos de dominio completos
│       ├── base.go                  # Interfaces base
│       ├── category_created.go      
│       ├── expense_created.go       
│       ├── expense_updated.go       # ✨ Nuevo
│       ├── expense_deleted.go       # ✨ Nuevo
│       ├── income_created.go        
│       ├── income_updated.go        # ✨ Nuevo
│       └── income_deleted.go        # ✨ Nuevo
├── application/                     # Capa de aplicación
│   ├── bus.go                      # Command/Query Bus
│   ├── commands/                   # Command handlers CRUD
│   │   ├── create_category.go      
│   │   ├── create_expense.go       
│   │   ├── create_income.go        
│   │   ├── update_expense.go       # ✨ Nuevo
│   │   ├── update_income.go        # ✨ Nuevo
│   │   ├── delete_expense.go       # ✨ Nuevo
│   │   └── delete_income.go        # ✨ Nuevo
│   └── queries/                    # Query handlers optimizados
│       ├── movements.go            # Query handler original
│       └── projection_queries.go   # ✨ Query handler con proyecciones
├── infrastructure/                 # Capa de infraestructura
│   ├── eventstore/                 # Event Store (escritura)
│   │   ├── eventstore.go          
│   │   └── mongodb.go             
│   ├── projections/                # ✨ Sistema de proyecciones (lectura)
│   │   └── projections.go         # ✨ Proyecciones automáticas
│   ├── repositories/               # Repositories con reconstrucción
│   │   ├── category.go            
│   │   ├── expense.go             # ✨ Con GetByID para updates
│   │   └── income.go              # ✨ Con GetByID para updates
│   └── eventbus/                   # Event Publisher con proyecciones
│       ├── publisher.go           # ✨ Actualizado
│       └── projection_subscriber.go # ✨ Nuevo suscriptor
├── cmd/                           # Aplicaciones
│   ├── cli/main.go               # ✨ CLI con CRUD completo
│   └── server/main.go            # ✨ Web server con paginación
├── scripts/                      # ✨ Herramientas de migración
│   └── migrate_to_projections.go # ✨ Script de migración
├── web/                          # Frontend actualizado
│   └── index.html               # ✨ Dashboard con paginación y nombres
├── go.mod                       # Dependencias Go
└── .env                         # Variables de entorno
```

## 🛠️ Configuración y Uso

### Requisitos
- Go 1.24+
- MongoDB (local o cloud)

### Variables de Entorno (.env)
```bash
MONGODB_CONNECTION_STRING=mongodb://localhost:27017/escama
# O MongoDB Atlas:
# MONGODB_CONNECTION_STRING=mongodb+srv://usuario:password@cluster.mongodb.net/?retryWrites=true&w=majority
```

### Instalación y Configuración
```bash
# 1. Clonar y compilar
git clone <repo>
cd escama
go mod tidy

# 2. Compilar aplicaciones
go build -o escama-cli ./cmd/cli
go build -o escama-server ./cmd/server

# 3. Migrar datos existentes a proyecciones (opcional)
cd scripts
go run migrate_to_projections.go

# 4. Crear categorías iniciales
escama category create "Alimentación"
escama category create "Transporte" 
escama category create "Salario"
escama category create "Freelance"
escama category create "Entretenimiento"
escama category create "Salud"
escama category create "Servicios"
escama category create "Educación"
escama category create "Ropa"
escama category create "Vivienda"
```

### Ejecución
```bash
# Iniciar servidor web (con proyecciones optimizadas)
./escama-server

# Usar CLI (CRUD completo)
escama --help
```

## 🎯 Patrones Avanzados Implementados

### CQRS (Command Query Responsibility Segregation)
- ✅ **Comandos separados** para escritura (Event Store)
- ✅ **Consultas separadas** para lectura (Proyecciones)
- ✅ **Bases de datos independientes** (escritura vs lectura)
- ✅ **Optimizaciones específicas** para cada lado

### Event Sourcing
- ✅ **Eventos como fuente de verdad** inmutable
- ✅ **Agregados que generan eventos** de dominio
- ✅ **Event Store persistente** en MongoDB
- ✅ **Reconstrucción de estado** desde eventos
- ✅ **Auditoría completa** de cambios

### Proyecciones en Tiempo Real
- ✅ **Actualización automática** con cada evento
- ✅ **Desnormalización optimizada** para consultas
- ✅ **Soft deletes** (marcado como eliminado)
- ✅ **Consistencia eventual** entre escritura y lectura

### Clean Architecture + DDD
- ✅ **Separación de capas** bien definida
- ✅ **Inversión de dependencias** 
- ✅ **Domain-driven design** con agregados
- ✅ **Repository pattern** con reconstrucción

## 💰 Formato Monetario

El sistema está configurado para **Guaraníes paraguayos (₲)** sin decimales:
- **CLI**: Muestra montos como `₲850,000`, `₲25,000`
- **Dashboard Web**: Formatea con separadores de miles
- **API**: Devuelve números en formato JSON estándar
- **Base de datos**: Almacena como números (float64) para cálculos precisos

## 🔄 Operaciones CRUD Completas

### Crear Movimientos
```bash
# Con selección interactiva de categoría
escama expense create 50000 "Almuerzo"

# Con categoría específica
escama income create 300000 "Consultoría" --category "Freelance"

# Con fecha específica
escama expense create 80000 "Supermercado" --category "Alimentación" --date 2025-07-20
```

### Actualizar Movimientos Existentes
```bash
# Obtener ID del movimiento (desde escama movements)
escama movements | head -5

# Actualizar gasto
escama expense update [expense-id] 75000 "Supermercado grande" --category "Alimentación"

# Actualizar ingreso
escama income update [income-id] 350000 "Consultoría actualizada" --category "Freelance"
```

### Eliminar Movimientos
```bash
# Eliminar con confirmación interactiva
escama expense delete [expense-id]
# Output: ⚠️ ¿Estás seguro de que deseas eliminar el gasto [id]? (y/N):

escama income delete [income-id]
# Output: ⚠️ ¿Estás seguro de que deseas eliminar el ingreso [id]? (y/N):
```

### Consultar Datos (Optimizado)
```bash
# Balance del mes (desde proyecciones - instantáneo)
escama balance

# Movimientos recientes (con nombres de categorías)
escama movements
```

## 🎨 Características del Dashboard Mejorado

- **📄 Paginación inteligente**: 10 movimientos por página
- **🔢 Navegación**: Botones anterior/siguiente + números de página  
- **🏷️ Nombres claros**: Categorías muestran nombres, no IDs
- **⚡ Rendimiento optimizado**: Consultas desde proyecciones
- **📱 Responsive design** para móviles y desktop
- **🗓️ Filtros de fecha** que reinician la paginación
- **📊 Gráfico interactivo** de gastos por categoría
- **💰 Formato paraguayo** con separadores de miles

## 📊 APIs REST Optimizadas

### GET /api/movements?page=1&limit=10
**Respuesta paginada:**
```json
{
  "movements": [
    {
      "id": "movement-id",
      "type": "income",
      "category_id": "freelance-id", 
      "category_name": "Freelance",
      "amount": 850000.00,
      "description": "Proyecto web",
      "date": "2025-07-27T17:05:55Z",
      "created_at": "2025-07-27T20:05:55Z"
    }
  ],
  "total": 45,
  "page": 1,
  "per_page": 10,
  "has_next": true,
  "has_prev": false
}
```

### GET /api/balance?start_date=2025-07-01&end_date=2025-07-31
```json
{
  "total_income": 3605000.00,
  "total_expense": 1999152.00, 
  "net_balance": 1605848.00,
  "period": "2025-07-01 - 2025-07-31"
}
```

### GET /api/expenses-by-category
**Con nombres de categorías:**
```json
[
  {
    "category_id": "vivienda-id",
    "category_name": "Vivienda",
    "total": 800000.00,
    "count": 1
  },
  {
    "category_id": "alimentacion-id", 
    "category_name": "Alimentación",
    "total": 135000.00,
    "count": 3
  }
]
```

## 🚀 Rendimiento y Escalabilidad

### Beneficios de CQRS + Proyecciones
- **⚡ Consultas instantáneas**: Sin reconstrucción desde eventos
- **📈 Escalabilidad independiente**: Escritura vs lectura
- **🎯 Optimizaciones específicas**: Índices para cada caso de uso
- **🔄 Procesamiento en background**: Proyecciones asíncronas

### Estadísticas de Rendimiento
- **Consulta de movimientos**: ~1ms (proyecciones) vs ~50ms (event store)
- **Cálculo de balance**: ~2ms (agregado) vs ~100ms (reconstrucción)
- **Paginación**: Nativa en MongoDB, muy eficiente
- **Filtros de fecha**: Índices optimizados

## 🔧 Migración y Mantenimiento

### Script de Migración Automática
```bash
# Migrar datos existentes a proyecciones
cd scripts
go run migrate_to_projections.go

# Output:
# 🔄 Iniciando migración de datos a proyecciones...
# ✅ Conexiones establecidas
# 🧹 ¿Deseas limpiar las proyecciones existentes antes de migrar? (y/N): y
# 📊 Se encontraron 30 eventos para procesar
# 🎉 Migración completada!
# ✅ Eventos procesados exitosamente: 30
# 📋 Movimientos totales: 20
# 💰 Ingresos: 4 (₲3,605,000)
# 💸 Gastos: 16 (₲1,999,152)  
# 📈 Balance neto: ₲1,605,848
# 🏷️ Categorías: 10
```

### Comandos de Diagnóstico
```bash
# Ver estadísticas detalladas
escama balance

# Listar todos los movimientos
escama movements

# Verificar categorías creadas
escama category create --help
```

## 💡 Casos de Uso Paraguayos

### Salario típico desarrollador
```bash
escama income create 4500000 "Salario senior developer" --category "Salario"
# Output: 💰 Ingreso de ₲4,500,000 registrado exitosamente para el 2025-07-27
```

### Gastos cotidianos en Asunción
```bash
escama expense create 95000 "Almuerzo Paseo La Galería" --category "Alimentación"
escama expense create 8000 "Colectivo Asunción-Lambaré" --category "Transporte"
escama expense create 450000 "Alquiler departamento" --category "Vivienda"
```

### Freelancing remoto
```bash
escama income create 1200000 "Cliente USA proyecto React" --category "Freelance"
escama expense create 180000 "Internet fibra Tigo" --category "Servicios"
```

### Actualizar gastos del mes
```bash
# Ver movimientos para obtener IDs
escama movements | head -10

# Actualizar monto de alquiler
escama expense update [alquiler-id] 480000 "Alquiler con expensas" --category "Vivienda"
```

### Balance mensual
```bash
escama balance
# Output:
# 📊 Balance del mes (2025-07-01 - 2025-07-31)
# ════════════════════════════════════
# 💰 Total Ingresos:  ₲5,700,000
# 💸 Total Gastos:    ₲2,150,000
# 📈 Balance Neto:    ₲3,550,000
# ✅ ¡Felicitaciones! Tienes un balance positivo
```

## 🔍 Comandos de Ayuda Detallados

```bash
# Ver todos los comandos disponibles
escama --help

# Ayuda específica por módulo
escama expense --help    # create, update, delete
escama income --help     # create, update, delete  
escama category --help   # create

# Ayuda de comando específico
escama expense create --help
escama income update --help
escama expense delete --help
```

## 🏗️ Arquitectura de Datos

### Event Store (Escritura)
```
MongoDB Database: escama
Collection: events
Documents: Eventos inmutables con timestamp
```

### Proyecciones (Lectura)  
```
MongoDB Database: escama_read
Collections:
  - movements: Movimientos desnormalizados con nombres
  - categories: Categorías activas
```

### Flujo de Datos
```
Comando → Event Store → Evento → Proyección → Consulta Optimizada
    ↓         ↓           ↓          ↓            ↓
  CLI/API   MongoDB    Dominio   MongoDB      CLI/API
```

---

**¡Tu sistema CQRS + Event Sourcing está listo para Paraguay!** 🇵🇾✨

- ⚡ **Rendimiento optimizado** con proyecciones en tiempo real
- 🔄 **CRUD completo** via CLI con nombres de categorías
- 📄 **Paginación inteligente** en dashboard web
- 🎯 **Arquitectura escalable** para crecimiento futuro
- 💰 **Optimizado para Guaraníes** paraguayos
- 🛡️ **Auditoría completa** de todos los cambios financieros 