# Todo API — Golang cho người chuyển từ Node.js

Tài liệu này giải thích **công nghệ**, **cấu trúc Clean Architecture**, và **flow chạy code** của project. Mọi khái niệm đều có đối chiếu sang Node.js để bạn dễ hình dung.

> **Cập nhật:** project đã chuyển từ lưu trữ in-memory (`map`) sang **PostgreSQL** thông qua ORM **GORM**. Các phần liên quan đến storage bên dưới đã được cập nhật theo.

---

## 1. Công nghệ sử dụng

| Mục đích | Thư viện Go | Tương đương ở Node.js |
|---|---|---|
| HTTP framework | [`gin-gonic/gin`](https://github.com/gin-gonic/gin) | Express / Fastify |
| Validation | [`go-playground/validator`](https://github.com/go-playground/validator) | **Zod**, Joi, class-validator |
| Logger | [`uber-go/zap`](https://github.com/uber-go/zap) | Pino, Winston |
| UUID | [`google/uuid`](https://github.com/google/uuid) | `uuid` package |
| **ORM** | [`gorm.io/gorm`](https://gorm.io) | **Prisma**, TypeORM, Sequelize |
| **Postgres driver** | [`gorm.io/driver/postgres`](https://github.com/go-gorm/postgres) (pgx) | `pg` / `postgres` |
| **Database** | **PostgreSQL** (`todo_api_golang`) | PostgreSQL / MySQL |

### Vì sao chọn các thư viện này?

- **Gin**: nhanh nhất trong các framework Go phổ biến, API đơn giản giống Express, được dùng rộng rãi trong industry.
- **go-playground/validator**: chuẩn de-facto của hệ sinh thái Go. Khai báo validate ngay trên struct bằng tag — tương đương `z.object({...})` của Zod, nhưng nằm cạnh field thay vì file schema riêng.
- **Zap**: structured logger nhanh nhất Go, xuất JSON chuẩn để stream vào ELK/Loki/Datadog.
- **GORM**: ORM phổ biến nhất của Go. Map struct ↔ bảng bằng tag `gorm:"..."`, có `AutoMigrate`, query builder, hooks — tương tự Prisma/TypeORM bên Node.

---

## 2. Cấu trúc thư mục (Clean Architecture)

```
todo_api_v1/
├── main.go                       # Entry point — kết nối DB, wire dependencies, start server
├── internal/                     # Code "nội bộ" — Go cấm package ngoài import vào đây
│   ├── domain/
│   │   └── todo.go              # Entity (model) + tag gorm map xuống cột Postgres
│   ├── dto/
│   │   └── todo_dto.go          # Request/Response shape + rule validate
│   ├── repository/
│   │   └── todo_repository.go   # Lớp lưu trữ — PostgreSQL qua GORM
│   ├── usecase/
│   │   └── todo_usecase.go      # Business logic
│   ├── handler/
│   │   └── todo_handler.go      # HTTP controller — decode req, gọi usecase
│   ├── middleware/
│   │   └── logger.go            # Request logging + panic recovery
│   └── router/
│       └── router.go            # Đăng ký endpoints
├── pkg/                          # Code có thể tái sử dụng cho project khác
│   ├── database/
│   │   └── postgres.go          # Mở kết nối GORM tới PostgreSQL (đọc DSN từ env)
│   ├── logger/
│   ├── validator/
│   └── response/
├── go.mod                        # Như package.json
└── golang.md
```

### Quy tắc Clean Architecture (chiều phụ thuộc)

```
handler  →  usecase  →  repository  →  domain
   ↑           ↑             ↑              ↑
   └── dto     └── dto       └── gorm.DB    (entity thuần)
```

- **Lớp trong không biết lớp ngoài.** `domain` không biết `usecase`; `usecase` không biết `handler`.
- `usecase` phụ thuộc **interface** `TodoRepository`, không phụ thuộc implementation. → Đây chính là lý do việc đổi từ in-memory sang Postgres chỉ cần viết struct `postgresRepo` mới implement interface, **không sửa usecase/handler**.

---

## 3. Mapping khái niệm Node.js → Go

| Node.js | Go |
|---|---|
| `npm install` | `go mod tidy` |
| `package.json` | `go.mod` |
| `node_modules/` | Cache global (`$GOPATH/pkg/mod`), không có folder trong project |
| `class TodoService` | `type TodoUsecase struct { ... }` + method `func (u *TodoUsecase) ...` |
| `interface ITodoRepo` (TS) | `type TodoRepository interface { ... }` |
| `try/catch` | `if err != nil { return err }` |
| `throw new Error(...)` | `return errors.New(...)` hoặc `return fmt.Errorf(...)` |
| `async/await` + Promise | **Goroutine** (`go funcName()`) + channel — nhưng I/O thường blocking-style, runtime tự điều phối |
| `import { x } from './y'` | `import "github.com/.../y"` (path tuyệt đối từ module name) |
| Field public `this.name` | Field **viết hoa** chữ đầu (`Name`) là public; viết thường là private package |
| Constructor | Hàm `New<Type>(...)` trả về `*Type` (convention) |
| `JSON.stringify` | Tag `json:"field_name"` + `encoding/json` (Gin tự lo) |
| Zod schema | Tag `validate:"required,min=1"` ngay trên field struct |
| Prisma `@id`, `@default` trong schema | Tag `gorm:"primaryKey;default:..."` ngay trên field struct |
| `prisma migrate` | `db.AutoMigrate(&Todo{})` chạy lúc khởi động |

### Pointer (`*`) - khái niệm mới cần nắm

```go
func (u *TodoUsecase) Create(...)   // (u *TodoUsecase) = "u" là con trỏ tới TodoUsecase
*req.Title                           // dereference: lấy giá trị mà con trỏ trỏ tới
&todo                                // lấy địa chỉ của biến todo
```

- Dùng `*Type` khi bạn muốn **sửa** giá trị hoặc tránh copy struct lớn.
- Trong DTO `UpdateTodoRequest`, tôi dùng `*string` để phân biệt **"không truyền"** (`nil`) với **"truyền chuỗi rỗng"** (`""`) — đây là pattern PATCH partial update.

---

## 4. Kết nối Database (GORM + PostgreSQL)

### Entity map xuống bảng (`internal/domain/todo.go`)

```go
type Todo struct {
    ID          string    `gorm:"type:uuid;primaryKey" json:"id"`
    Title       string    `gorm:"type:varchar(200);not null" json:"title"`
    Description string    `gorm:"type:varchar(1000)" json:"description"`
    Completed   bool      `gorm:"not null;default:false" json:"completed"`
    CreatedAt   time.Time `json:"created_at"`   // GORM tự fill khi Create
    UpdatedAt   time.Time `json:"updated_at"`   // GORM tự cập nhật khi Save
}

func (Todo) TableName() string { return "todos" }
```

- `CreatedAt` / `UpdatedAt` là tên field "thần kỳ" của GORM: nó tự động set khi tạo / cập nhật (giống `@default(now())` / `@updatedAt` của Prisma).

### Mở kết nối (`pkg/database/postgres.go`)

DSN (chuỗi kết nối) được build từ **biến môi trường**, có sẵn giá trị mặc định để chạy local ngay:

| Biến env | Mặc định |
|---|---|
| `DB_HOST` | `localhost` |
| `DB_PORT` | `5432` |
| `DB_USER` | `postgres` |
| `DB_PASSWORD` | `123123` |
| `DB_NAME` | `todo_api_golang` |
| `DB_SSLMODE` | `disable` |

```go
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: gormlogger.Default.LogMode(gormlogger.Warn),
})
```

### Repository dùng GORM (`internal/repository/todo_repository.go`)

```go
func NewPostgresTodoRepository(db *gorm.DB) TodoRepository { return &postgresRepo{db: db} }

func (r *postgresRepo) Create(t *domain.Todo) error { return r.db.Create(t).Error }

func (r *postgresRepo) GetByID(id string) (*domain.Todo, error) {
    var t domain.Todo
    if err := r.db.First(&t, "id = ?", id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) { return nil, ErrNotFound }
        return nil, err
    }
    return &t, nil
}
// Update dùng Save (ghi đè cả Completed=false); Delete + RowsAffected==0 → ErrNotFound
```

> **Lưu ý quan trọng:** vì query DB có thể lỗi, interface đã đổi `List()` từ `[]*domain.Todo` thành `List() ([]*domain.Todo, error)`. Usecase và handler đã được cập nhật để xử lý error này.

---

## 5. Flow chạy một request

Giả sử client gửi `POST /api/v1/todos` với body `{"title":"buy milk"}`:

```
┌─────────┐
│ Client  │  POST /api/v1/todos  {"title":"buy milk"}
└────┬────┘
     │
     ▼
┌─────────────────────────────────────────────────────────────┐
│ 1. Gin Engine (router.go)                                   │
│    - Match route POST /api/v1/todos → todoH.Create          │
│    - Chạy chain middleware: Recovery → RequestLogger        │
└────┬────────────────────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. TodoHandler.Create (handler/todo_handler.go)             │
│    a. c.ShouldBindJSON(&req)   → decode JSON vào struct     │
│    b. validator.Struct(req)    → check tag `validate:"..."` │
│       Nếu fail → 400 Bad Request, dừng                      │
└────┬────────────────────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. TodoUsecase.Create (usecase/todo_usecase.go)             │
│    - Sinh UUID, set CreatedAt/UpdatedAt = now               │
│    - Gọi repo.Create(t)                                     │
└────┬────────────────────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. postgresRepo.Create (repository/todo_repository.go)      │
│    - r.db.Create(t)  → GORM sinh câu lệnh:                  │
│      INSERT INTO "todos" (...) VALUES (...)                  │
│    - Trả về .Error (nil nếu thành công)                     │
└────┬────────────────────────────────────────────────────────┘
     │  trả về (*Todo, nil)
     ▼
┌─────────────────────────────────────────────────────────────┐
│ 5. Handler trả response                                     │
│    response.OK(c, 201, "todo created", t)                   │
│    → c.JSON(201, Envelope{Success: true, Data: t})          │
└────┬────────────────────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────────────────────────┐
│ 6. Middleware RequestLogger chạy code sau c.Next():         │
│    log.Info("request", method, path, status, latency, ...)  │
└────┬────────────────────────────────────────────────────────┘
     │
     ▼
┌─────────┐
│ Client  │  201 Created  {"success":true,"data":{...}}
└─────────┘
```

### Flow khởi động (`main.go`)

```go
1. logger.New()                          // tạo logger
2. database.New()                        // ┐ mở kết nối GORM tới PostgreSQL
3. db.AutoMigrate(&domain.Todo{})        // ┘ tự tạo/cập nhật bảng "todos"
4. repository.NewPostgresTodoRepository  // ┐
5. usecase.NewTodoUsecase(repo)          // │ Dependency Injection thủ công
6. validator.New()                       // │ (không có decorator như Nest)
7. handler.NewTodoHandler(uc, v, log)    // ┘
8. router.New(handler, log)              // gắn route + middleware
9. srv.ListenAndServe() trong goroutine  // server chạy non-blocking
10. signal.Notify(quit, SIGINT, SIGTERM) // chờ Ctrl+C
11. srv.Shutdown(ctx)                     // graceful shutdown (drain connections)
```

> `AutoMigrate` **không cần** chạy lệnh migration riêng — mỗi lần khởi động nó tự tạo bảng nếu chưa có và thêm cột mới nếu bạn thêm field. Nó không xóa cột / không đổi kiểu cột đã tồn tại (tránh mất dữ liệu).

---

## 6. API Endpoints

| Method | Path | Body | Mô tả |
|---|---|---|---|
| GET | `/health` | — | Health check |
| POST | `/api/v1/todos` | `{title, description?}` | Tạo todo |
| GET | `/api/v1/todos` | — | List todos |
| GET | `/api/v1/todos/:id` | — | Lấy 1 todo |
| PATCH | `/api/v1/todos/:id` | `{title?, description?, completed?}` | Cập nhật partial |
| DELETE | `/api/v1/todos/:id` | — | Xoá todo |

**Response envelope chuẩn:**

```json
{ "success": true,  "message": "todo created", "data": { ... } }
{ "success": false, "error": "title is required" }
```

---

## 7. Cách chạy

### Yêu cầu trước

1. PostgreSQL đang chạy (mặc định `localhost:5432`).
2. Đã tạo database tên `todo_api_golang` (tạo bằng pgAdmin hoặc `CREATE DATABASE todo_api_golang;`).
3. Thông tin user/password khớp với biến env ở mục 4 (mặc định `postgres` / `123123`).

> Bảng `todos` **không cần tạo tay** — `AutoMigrate` sẽ tự tạo khi chạy app lần đầu.

```bash
# Tải dependencies (đọc go.mod, ghi go.sum)
go mod tidy

# Chạy app (entry point ở thư mục gốc)
go run .

# Build binary (1 file, không cần Node runtime)
go build -o todo-api .
./todo-api

# Đổi port (mặc định 3636)
PORT=3000 go run .
```

### Tuỳ chỉnh kết nối DB bằng env (ví dụ PowerShell)

```powershell
$env:DB_PASSWORD = "your_password"
$env:DB_NAME     = "todo_api_golang"
go run .
```

### Test nhanh bằng curl

```bash
# Create
curl -X POST localhost:3636/api/v1/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"buy milk","description":"2L oat milk"}'

# List
curl localhost:3636/api/v1/todos

# Update
curl -X PATCH localhost:3636/api/v1/todos/<id> \
  -H "Content-Type: application/json" \
  -d '{"completed":true}'

# Delete
curl -X DELETE localhost:3636/api/v1/todos/<id>
```

---

## 8. Những điểm dễ vấp khi mới chuyển từ Node

1. **Public/Private dựa vào chữ hoa**: `func foo()` chỉ dùng trong package; `func Foo()` mới export ra ngoài. Đây là lý do tất cả constructor đặt là `New<Type>` (viết hoa).
2. **Không có `null`, có `nil`** — và `nil` chỉ dùng được với pointer, slice, map, channel, function, interface.
3. **Error là giá trị, không phải exception**: hàm trả `(result, error)`, bạn phải check `if err != nil` ngay. Không có `try/catch` (chỉ có `panic/recover` cho lỗi không-thể-cứu).
4. **Không có `undefined`**: field không khai báo trong JSON sẽ về zero-value (`""`, `0`, `false`). Đó là lý do PATCH dùng pointer (`*string`) để biết client có truyền hay không. Khi update, dùng `db.Save` (ghi đè cả cột) thay vì `db.Updates` (bỏ qua zero-value) để `completed=false` vẫn được lưu.
5. **`go.mod` tracking exact**: thêm import mới chạy `go mod tidy` để tải về (giống `npm install` nhưng tự động dựa trên code).
6. **Interface implicit**: bạn không khai báo `class X implements Y`. Chỉ cần struct có đủ method, Go tự nhận là implement interface đó. → Tách layer cực dễ (đây là cách `postgresRepo` thay thế `inMemoryRepo` mà không sửa usecase).
7. **GORM `ErrRecordNotFound`**: khi `First` không tìm thấy bản ghi, GORM trả lỗi `gorm.ErrRecordNotFound`. Repository bắt lỗi này và map sang `ErrNotFound` riêng của domain để usecase/handler trả về HTTP 404.

---

## 9. Hướng mở rộng

- **Connection pooling / timeout**: lấy `sqlDB, _ := db.DB()` rồi set `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`.
- **Soft delete**: thêm field `gorm.DeletedAt` vào entity → GORM tự chuyển `DELETE` thành `UPDATE deleted_at = now()`.
- **Migration chuyên nghiệp**: thay `AutoMigrate` bằng [`golang-migrate`](https://github.com/golang-migrate/migrate) hoặc `goose` để versioning migration trong production.
- **Thêm auth**: viết middleware kiểm JWT, đặt trước route cần bảo vệ (`v1.Use(authMiddleware)`).
- **Config từ env**: gom các biến `DB_*`, `PORT` vào struct config dùng `github.com/spf13/viper` hoặc `github.com/caarlos0/env` thay vì đọc lẻ `os.Getenv`.
- **Test**: Go có sẵn `testing` package — đặt file `*_test.go` cạnh code, chạy `go test ./...`. Mock repository dễ vì usecase phụ thuộc interface (không cần DB thật khi test usecase).
- **Production logger**: đổi `zap.NewDevelopmentConfig()` thành `zap.NewProductionConfig()` để xuất JSON gọn nhẹ.
