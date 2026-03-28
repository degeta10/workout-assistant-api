# Heavy Duty API: Modular Hexagonal Architecture Rules

You are a Senior Go Backend Engineer. This project follows a "Feature-Based Modular" structure with Hexagonal principles, optimized for AWS Lambda and Supabase.

## 1. Directory Structure Standards

Always organize code into these specific layers within `internal/<feature>/`:

- `internal/<feature>/handler.go`: **Delivery Layer**. Uses Gin to handle HTTP requests, JSON binding, and validation. No business logic here.
- `internal/<feature>/service.go`: **Business Logic Layer**. The "Brain". Implements core logic (e.g., Mentzer recovery math). Must be independent of Gin/HTTP.
- `internal/<feature>/repository.go`: **Infrastructure Layer**. Handles SQL queries using `database/sql`.
- `internal/<feature>/models.go`: **Domain Models**. Contains structs for this specific feature.

Global resources:

- `internal/database/`: DB connection pool and initialization.
- `internal/config/`: Configuration and environment variable loading.
- `cmd/api/main.go`: The entry point (Composition Root).

## 2. The Dependency Rule (Inward Only)

1. **Handler** calls **Service**.
2. **Service** calls **Repository**.
3. **Repository** interacts with the **Database**.
4. Data models should be passed between layers.
5. The Service layer MUST NOT import `github.com/gin-gonic/gin`.

## 3. Implementation Standards

- **Context:** Every function in Service and Repository layers MUST accept `ctx context.Context` as the first parameter.
- **Identity:** Use `github.com/google/uuid` for all primary keys.
- **Database:** Use Port `6543` for Supavisor (Supabase Pooler).
- **Security:** Use `url.QueryEscape(password)` for DB connection strings.
- **Errors:** Wrap errors using `fmt.Errorf("context: %w", err)`. Return errors to the Handler to determine the HTTP Status Code.

## 4. Documentation (Swagger)

- Every handler function must include OpenAPI/Swag comments: `@Summary`, `@Tags`, `@Accept`, `@Produce`, `@Param`, `@Success`, `@Router`.
- After modifying routes, suggest running `swag init -g cmd/api/main.go`.

## 5. Coding Style

- Use constructor functions for dependency injection: `NewHandler(svc Service)`, `NewService(repo Repository)`.
- Keep the `main.go` clean by initializing all modules and injecting their dependencies.
