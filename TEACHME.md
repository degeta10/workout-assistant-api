# Workout Assistant API - TEACHME

Practical guide for understanding this Go project and creating new modules. Written for someone coming from Laravel or PHP.

## 1. Quick Mental Model: Laravel -> This Go Project

| Laravel                  | This Go Project                             |
| ------------------------ | ------------------------------------------- |
| `public/index.php`       | `cmd/api/main.go`                           |
| `routes/api.php`         | Route registration inside `setupRouter()`   |
| Controller               | `internal/<feature>/handler.go`             |
| Service / Action         | `internal/<feature>/service.go`             |
| Repository / Query layer | `internal/<feature>/repository.go`          |
| FormRequest              | Binding tags on request structs             |
| Eloquent Model           | Feature structs in `models.go`              |
| `config/*.php` + `.env`  | `internal/config/config.go` + `.env`        |
| App container            | Manual dependency injection in `main.go`    |
| Middleware               | `internal/pkg/middleware` + auth middleware |

## 2. The Real Architecture Used Here

Each feature lives in its own folder under `internal/`.

Examples:

- `internal/auth/`
- `internal/health/`

Inside each feature you keep four files:

- `models.go`
- `repository.go`
- `service.go`
- `handler.go`

### What each file is responsible for

`models.go`

- Defines request/response structs
- Defines domain structs
- Defines feature interfaces
- Defines feature-specific errors

`repository.go`

- Talks to PostgreSQL using `database/sql`
- Runs SQL queries
- Converts DB rows into Go structs
- Must not know about Gin or HTTP

`service.go`

- Holds business rules
- Calls repository methods
- Returns plain Go values and errors
- Must not import `github.com/gin-gonic/gin`

`handler.go`

- The HTTP layer
- Reads JSON from requests
- Validates using binding tags
- Calls service methods
- Converts errors to HTTP responses
- Owns Swagger comments

## 3. What `cmd/api/main.go` Does

`cmd/api/main.go` is the composition root.

That means it does not contain business logic. Its job is to wire the app together.

Current responsibilities:

- Load config
- Initialize database
- Create Gin router
- Attach middleware
- Create `repo -> service -> handler` for each module
- Register public routes
- Register protected routes
- Run locally or inside AWS Lambda

The dependency flow is always inward:

```text
handler -> service -> repository -> database
```

Never the reverse.

## 4. Request Flow In This Project

Example: `GET /v1/me`

1. Request enters Gin router.
2. Request ID middleware adds a `request_id`.
3. Request logger middleware logs the request.
4. Auth middleware validates the Bearer token.
5. Auth middleware parses the JWT `sub` claim.
6. Auth middleware stores `userID` in Gin context and request context.
7. Handler calls service using `c.Request.Context()`.
8. Service reads values from context if needed.
9. Service calls repository.
10. Repository runs SQL.
11. Service returns result or error.
12. Handler sends success or pushes error to global error middleware.

This is the design rule:

- HTTP concerns stay in handler and middleware.
- Business logic stays in service.
- SQL stays in repository.

## 5. Shared Packages You Should Know

### `internal/pkg/responses/responses.go`

- Standard API response envelope
- Helpers like `OK`, `Created`, `Unauthorized`, `ValidationError`

Response shape:

```json
{
  "success": true,
  "message": "...",
  "data": {},
  "errors": {}
}
```

### `internal/pkg/apperrors/errors.go`

- Stores typed application errors
- Lets handlers express HTTP status intent cleanly

### `internal/pkg/middleware/error_handler.go`

- Reads `c.Error(...)`
- Converts app errors and validation errors to consistent JSON

### `internal/pkg/middleware/request_id.go`

- Creates or forwards a request ID

### `internal/pkg/middleware/request_logger.go`

- Logs method, path, status, latency, and `request_id`

## 6. Current Auth Module Explained

### `internal/auth/models.go`

- Defines `User`, `RegisterRequest`, `LoginRequest`, `LoginResponse`, `UserSummary`
- Defines `Repository` and `Service` interfaces
- Defines errors like:
  - `ErrEmailAlreadyExists`
  - `ErrInvalidCredentials`
  - `ErrUserNotFound`
  - `ErrUserIDMissing`

### `internal/auth/repository.go`

- `CreateUser` inserts into `users`
- `GetUserByEmail` fetches login data
- `GetUserByID` fetches the current authenticated user

### `internal/auth/service.go`

- `Register` hashes password with bcrypt and creates user
- `Login` verifies password and signs JWT
- `Me` reads `userID` from request context and fetches user

### `internal/auth/handler.go`

- `RegisterPublicRoutes` adds `/register` and `/login`
- `RegisterProtectedRoutes` adds `/me`
- Handler validates input and maps service errors to app errors

### `internal/auth/middleware.go`

- Checks `Authorization` header
- Validates `Bearer <token>` format
- Parses JWT
- Extracts `sub`
- Parses `sub` as UUID
- Stores `userID` in request context for downstream layers

## 7. Tutorial: Create a New Module

This is the part to follow every time.

We will create a new module called `workouts`.

Goal:

- `POST /v1/workouts` creates a workout
- `GET /v1/workouts` lists the authenticated user’s workouts

Assumption:

- These routes are protected, so `userID` comes from JWT middleware.

## 8. Step 1: Create the Folder

Create:

- `internal/workouts/models.go`
- `internal/workouts/repository.go`
- `internal/workouts/service.go`
- `internal/workouts/handler.go`

## 9. Step 2: Write `models.go`

```go
package workouts

import (
   "context"
   "errors"
   "time"

   "github.com/google/uuid"
)

var ErrWorkoutNameRequired = errors.New("workout name is required")

type Workout struct {
   ID        uuid.UUID `json:"id"`
   UserID    uuid.UUID `json:"user_id"`
   Name      string    `json:"name"`
   Notes     string    `json:"notes"`
   CreatedAt time.Time `json:"created_at"`
}

type CreateWorkoutRequest struct {
   Name  string `json:"name" binding:"required"`
   Notes string `json:"notes"`
}

type Repository interface {
   CreateWorkout(ctx context.Context, workout Workout) (uuid.UUID, error)
   ListWorkoutsByUserID(ctx context.Context, userID uuid.UUID) ([]Workout, error)
}

type Service interface {
   Create(ctx context.Context, name, notes string) (uuid.UUID, error)
   List(ctx context.Context) ([]Workout, error)
}
```

### Why this file exists

- It defines the language of the module.
- Every other file depends on these types.

### Why the interfaces are here

- `handler.go` depends on `Service`
- `service.go` depends on `Repository`
- `models.go` becomes the contract layer for the feature

### Why `ctx context.Context` is required

- Request cancellation
- Deadlines and timeouts
- Request-scoped values like `userID`

## 10. Step 3: Write `repository.go`

```go
package workouts

import (
   "context"
   "database/sql"
   "fmt"

   "github.com/google/uuid"
)

type PostgresRepository struct {
   db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
   return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateWorkout(ctx context.Context, workout Workout) (uuid.UUID, error) {
   query := `
      insert into workouts (user_id, name, notes)
      values ($1, $2, $3)
      returning id
   `

   var id uuid.UUID
   if err := r.db.QueryRowContext(ctx, query, workout.UserID, workout.Name, workout.Notes).Scan(&id); err != nil {
      return uuid.Nil, fmt.Errorf("insert workout: %w", err)
   }

   return id, nil
}

func (r *PostgresRepository) ListWorkoutsByUserID(ctx context.Context, userID uuid.UUID) ([]Workout, error) {
   query := `
      select id, user_id, name, notes, created_at
      from workouts
      where user_id = $1
      order by created_at desc
   `

   rows, err := r.db.QueryContext(ctx, query, userID)
   if err != nil {
      return nil, fmt.Errorf("list workouts: %w", err)
   }
   defer rows.Close()

   var workouts []Workout
   for rows.Next() {
      var workout Workout
      if err := rows.Scan(
         &workout.ID,
         &workout.UserID,
         &workout.Name,
         &workout.Notes,
         &workout.CreatedAt,
      ); err != nil {
         return nil, fmt.Errorf("scan workout: %w", err)
      }
      workouts = append(workouts, workout)
   }

   if err := rows.Err(); err != nil {
      return nil, fmt.Errorf("iterate workouts: %w", err)
   }

   return workouts, nil
}
```

### Detailed explanation

`type PostgresRepository struct { db *sql.DB }`

- This struct stores the database connection pool.

`func NewRepository(db *sql.DB) Repository`

- Constructor function
- Returns the interface, not the concrete type
- Keeps the rest of the app dependent on the contract

`QueryRowContext` and `QueryContext`

- Always use the context-aware versions in this codebase
- They respect request cancellation and deadlines

`fmt.Errorf("insert workout: %w", err)`

- Wraps the error with useful context
- This is the project standard

`rows.Err()`

- Must be checked after iterating rows
- Some database errors appear only after iteration

## 11. Step 4: Write `service.go`

```go
package workouts

import (
   "context"
   "fmt"
   "strings"

   "github.com/google/uuid"
)

type WorkoutService struct {
   repo Repository
}

func NewService(repo Repository) Service {
   return &WorkoutService{repo: repo}
}

func (s *WorkoutService) Create(ctx context.Context, name, notes string) (uuid.UUID, error) {
   userIDValue := ctx.Value("userID")
   if userIDValue == nil {
      return uuid.Nil, ErrUserIDMissing
   }

   userID, ok := userIDValue.(uuid.UUID)
   if !ok {
      return uuid.Nil, fmt.Errorf("invalid user id type in context")
   }

   cleanedName := strings.TrimSpace(name)
   if cleanedName == "" {
      return uuid.Nil, ErrWorkoutNameRequired
   }

   id, err := s.repo.CreateWorkout(ctx, Workout{
      UserID: userID,
      Name:   cleanedName,
      Notes:  strings.TrimSpace(notes),
   })
   if err != nil {
      return uuid.Nil, fmt.Errorf("create workout: %w", err)
   }

   return id, nil
}

func (s *WorkoutService) List(ctx context.Context) ([]Workout, error) {
   userIDValue := ctx.Value("userID")
   if userIDValue == nil {
      return nil, ErrUserIDMissing
   }

   userID, ok := userIDValue.(uuid.UUID)
   if !ok {
      return nil, fmt.Errorf("invalid user id type in context")
   }

   workouts, err := s.repo.ListWorkoutsByUserID(ctx, userID)
   if err != nil {
      return nil, fmt.Errorf("list workouts by user id: %w", err)
   }

   return workouts, nil
}
```

### Detailed explanation

Why service exists:

- It protects business rules from HTTP and SQL details.

Why `strings.TrimSpace` belongs here:

- Cleaning and validating business input is business logic.
- Handlers only bind transport input.

Why `userID` is read here:

- The handler should not know how the business rule uses identity.
- The service decides that workouts belong to the authenticated user.

Why this file must not import Gin:

- If service imports Gin, it becomes tied to HTTP.
- Then you cannot reuse it cleanly from another delivery method.

## 12. Step 5: Write `handler.go`

```go
package workouts

import (
   "errors"

   "github.com/degeta10/workout-assistant-api/internal/pkg/apperrors"
   "github.com/degeta10/workout-assistant-api/internal/pkg/responses"
   "github.com/gin-gonic/gin"
)

type Handler struct {
   svc Service
}

func NewHandler(svc Service) *Handler {
   return &Handler{svc: svc}
}

func (h *Handler) RegisterProtectedRoutes(group *gin.RouterGroup) {
   group.POST("/workouts", h.Create)
   group.GET("/workouts", h.List)
}

// Create godoc
// @Summary Create a workout
// @Tags workouts
// @Accept json
// @Produce json
// @Param workout body CreateWorkoutRequest true "Workout payload"
// @Success 201 {object} responses.APIResponse
// @Failure 401 {object} responses.APIResponse
// @Failure 422 {object} responses.APIResponse
// @Failure 500 {object} responses.APIResponse
// @Router /workouts [post]
func (h *Handler) Create(c *gin.Context) {
   var req CreateWorkoutRequest
   if err := c.ShouldBindJSON(&req); err != nil {
      _ = c.Error(apperrors.Validation(err))
      return
   }

   id, err := h.svc.Create(c.Request.Context(), req.Name, req.Notes)
   if err != nil {
      if errors.Is(err, ErrUserIDMissing) {
         _ = c.Error(apperrors.Unauthorized("Invalid credentials", err))
         return
      }
      if errors.Is(err, ErrWorkoutNameRequired) {
         _ = c.Error(apperrors.Validation(err))
         return
      }
      _ = c.Error(apperrors.Internal("Failed to create workout", err))
      return
   }

   responses.Created(c, "Workout created successfully", map[string]any{"id": id})
}

// List godoc
// @Summary List workouts for the authenticated user
// @Tags workouts
// @Produce json
// @Success 200 {object} responses.APIResponse
// @Failure 401 {object} responses.APIResponse
// @Failure 500 {object} responses.APIResponse
// @Router /workouts [get]
func (h *Handler) List(c *gin.Context) {
   workouts, err := h.svc.List(c.Request.Context())
   if err != nil {
      if errors.Is(err, ErrUserIDMissing) {
         _ = c.Error(apperrors.Unauthorized("Invalid credentials", err))
         return
      }
      _ = c.Error(apperrors.Internal("Failed to fetch workouts", err))
      return
   }

   responses.OK(c, "Workouts fetched successfully", workouts)
}
```

### Detailed explanation

`ShouldBindJSON`

- Reads JSON body into a Go struct
- Runs validation using binding tags

`c.Request.Context()`

- Always pass this into service
- It carries timeouts, cancellation, request-scoped values, and user identity

`c.Error(...)`

- Handlers do not need to manually build every error response
- They delegate error formatting to `middleware.ErrorHandler()`

`responses.Created` and `responses.OK`

- Use the shared response format
- Keep success bodies consistent everywhere

## 13. Step 6: Wire It In `main.go`

Inside `setupRouter()`:

```go
workoutRepo := workouts.NewRepository(db)
workoutSvc := workouts.NewService(workoutRepo)
workoutHandler := workouts.NewHandler(workoutSvc)

protected := r.Group("/v1")
protected.Use(auth.RequireAuth(cfg.JWTSecret))
{
   authHandler.RegisterProtectedRoutes(protected)
   workoutHandler.RegisterProtectedRoutes(protected)
}
```

Why this is the correct place:

- `main.go` owns construction
- Feature files should not create their own DB connections
- Feature files should not read global config directly if `main` can inject it

## 14. Step 7: Add the Database Migration

You also need a migration for the new table.

Example idea:

```sql
create table if not exists public.workouts (
   id uuid primary key default gen_random_uuid(),
   user_id uuid not null references public.users(id) on delete cascade,
   name text not null,
   notes text not null default '',
   created_at timestamptz not null default now()
);
```

Then push the migration using the project command flow.

## 15. Step 8: Regenerate Docs and Build

After adding routes:

- `make swag-init`
- `go build ./...`

Why:

- Swagger comments must be regenerated
- Build catches missing imports, interface mismatches, and compile errors

## 16. Common Mistakes To Avoid

1. Importing Gin inside `service.go`
   - Wrong because business logic becomes HTTP-dependent.
2. Writing SQL inside `handler.go`
   - Wrong because handler should only do transport work.
3. Returning HTTP responses from `service.go`
   - Wrong because service should return values and errors, not JSON.
4. Forgetting `context.Context` in service and repository methods
   - Wrong because request cancellation and deadlines are lost.
5. Creating the DB connection inside a feature
   - Wrong because the app should have one shared pool.
6. Forgetting Swagger comments on handlers
   - Wrong because docs become incomplete.
7. Forgetting to register the module in `main.go`
   - Common compile-success but route-missing bug.

## 17. Fast Checklist When Adding a Module

- Created `internal/<feature>/`
- Added `models.go`
- Added `repository.go`
- Added `service.go`
- Added `handler.go`
- Added constructor functions
- Used `context.Context` in service and repository methods
- Used the `responses` package for success responses
- Used app errors or `c.Error(...)` for failures
- Wired `repo -> service -> handler` in `main.go`
- Registered routes in public or protected group
- Added a migration if a DB table is needed
- Ran `make swag-init`
- Ran `go build ./...`

## 18. What To Read Next

Best reading order in the current codebase:

1. `cmd/api/main.go`
2. `internal/auth/handler.go`
3. `internal/auth/service.go`
4. `internal/auth/repository.go`
5. `internal/auth/middleware.go`
6. `internal/pkg/middleware/error_handler.go`
7. `internal/pkg/responses/responses.go`
8. `internal/health/service.go`

If you understand those files, you understand the architecture and you can build a new module correctly.
