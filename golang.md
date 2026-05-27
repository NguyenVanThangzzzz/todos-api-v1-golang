# Todo API — Golang cho người chuyển từ Node.js

Tài liệu này giải thích **công nghệ**, **cấu trúc Clean Architecture**, và **flow chạy code** của project. Mọi khái niệm đều có đối chiếu sang Node.js để bạn dễ hình dung.

---

## 1. Công nghệ sử dụng

| Mục đích | Thư viện Go | Tương đương ở Node.js |
|---|---|---|
| HTTP framework | [`gin-gonic/gin`](https://github.com/gin-gonic/gin) | Express / Fastify |
| Validation | [`go-playground/validator`](https://github.com/go-playground/validator) | **Zod**, Joi, class-validator |
| Logger | [`uber-go/zap`](https://github.com/uber-go/zap) | Pino, Winston |
| UUID | [`google/uuid`](https://github.com/google/uuid) | `uuid` package |
| Storage | `map[string]*Todo` + `sync.RWMutex` | `Map` của JS (nhưng JS single-thread, Go cần mutex) |

### Vì sao chọn các thư viện này?

- **Gin**: nhanh nhất trong các framework Go phổ biến, API đơn giản giống Express, được dùng rộng rãi trong industry.
- **go-playground/validator**: chuẩn de-facto của hệ sinh thái Go. Khai báo validate ngay trên struct bằng tag — tương đương `z.object({...})` của Zod, nhưng nằm cạnh field thay vì file schema riêng.
- **Zap**: structured logger nhanh nhất Go, xuất JSON chuẩn để stream vào ELK/Loki/Datadog.

---

## 2. Cấu trúc thư mục (Clean Architecture)

```
todo_api_v1/
├── cmd/
│   └── api/
│       └── main.go              # Entry point — wire dependencies, start server
├── internal/                    # Code "nội bộ" — Go cấm package ngoài import vào đây
│   ├── domain/
│   │   └── todo.go              # Entity (model thuần, không phụ thuộc gì)
│   ├── dto/
│   │   └── todo_dto.go          # Request/Response shape + rule validate
│   ├── repository/
│   │   └── todo_repository.go   # Lớp lưu trữ (in-memory map)
│   ├── usecase/
│   │   └── todo_usecase.go      # Business logic
│   ├── handler/
│   │   └── todo_handler.go      # HTTP controller — decode req, gọi usecase
│   ├── middleware/
│   │   └── logger.go            # Request logging + panic recovery
│   └── router/
│       └── router.go            # Đăng ký endpoints
├── pkg/                         # Code có thể tái sử dụng cho project khác
│   ├── logger/
│   ├── validator/
│   └── response/
├── go.mod                       # Như package.json
└── golang.md
```

### Quy tắc Clean Architecture (chiều phụ thuộc)

```
handler  →  usecase  →  repository  →  domain
   ↑           ↑             ↑
   └── dto     └── dto       (chỉ phụ thuộc domain)
```

- **Lớp trong không biết lớp ngoài.** `domain` không biết `usecase`; `usecase` không biết `handler`.
- `usecase` phụ thuộc **interface** `TodoRepository`, không phụ thuộc implementation. → Đổi sang Postgres/Mongo chỉ cần viết struct mới implement interface, không sửa usecase.

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

### Pointer (`*`) - khái niệm mới cần nắm

```go
func (u *TodoUsecase) Create(...)   // (u *TodoUsecase) = "u" là con trỏ tới TodoUsecase
*req.Title                           // dereference: lấy giá trị mà con trỏ trỏ tới
&todo                                // lấy địa chỉ của biến todo
```

- Dùng `*Type` khi bạn muốn **sửa** giá trị hoặc tránh copy struct lớn.
- Trong DTO `UpdateTodoRequest`, tôi dùng `*string` để phân biệt **"không truyền"** (`nil`) với **"truyền chuỗi rỗng"** (`""`) — đây là pattern PATCH partial update.

---

## 4. Flow chạy một request

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
│ 4. inMemoryRepo.Create (repository/todo_repository.go)      │
│    - Lock mutex (chặn goroutine khác)                       │
│    - r.store[t.ID] = t                                      │
│    - Unlock                                                 │
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
2. repository.NewInMemoryTodoRepository  // ┐
3. usecase.NewTodoUsecase(repo)          // │ Dependency Injection thủ công
4. validator.New()                       // │ (không có decorator như Nest)
5. handler.NewTodoHandler(uc, v, log)    // ┘
6. router.New(handler, log)              // gắn route + middleware
7. srv.ListenAndServe() trong goroutine  // server chạy non-blocking
8. signal.Notify(quit, SIGINT, SIGTERM)  // chờ Ctrl+C
9. srv.Shutdown(ctx)                     // graceful shutdown (drain connections)
```

---

## 5. API Endpoints

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

## 6. Cách chạy

```bash
# Tải dependencies (đọc go.mod, ghi go.sum)
go mod tidy

# Chạy app
go run ./cmd/api

# Build binary (1 file, không cần Node runtime)
go build -o todo-api ./cmd/api
./todo-api

# Đổi port (mặc định 8080)
PORT=3000 go run ./cmd/api
```

### Test nhanh bằng curl

```bash
# Create
curl -X POST localhost:8080/api/v1/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"buy milk","description":"2L oat milk"}'

# List
curl localhost:8080/api/v1/todos

# Update
curl -X PATCH localhost:8080/api/v1/todos/<id> \
  -H "Content-Type: application/json" \
  -d '{"completed":true}'

# Delete
curl -X DELETE localhost:8080/api/v1/todos/<id>
```

---

## 7. Những điểm dễ vấp khi mới chuyển từ Node

1. **Public/Private dựa vào chữ hoa**: `func foo()` chỉ dùng trong package; `func Foo()` mới export ra ngoài. Đây là lý do tất cả constructor đặt là `New<Type>` (viết hoa).
2. **Không có `null`, có `nil`** — và `nil` chỉ dùng được với pointer, slice, map, channel, function, interface.
3. **Concurrent by default**: mỗi request là 1 goroutine. Nếu chia sẻ state (như map của repo) **phải** dùng mutex/channel. Node không cần vì event loop single-thread.
4. **Error là giá trị, không phải exception**: hàm trả `(result, error)`, bạn phải check `if err != nil` ngay. Không có `try/catch` (chỉ có `panic/recover` cho lỗi không-thể-cứu).
5. **Không có `undefined`**: field không khai báo trong JSON sẽ về zero-value (`""`, `0`, `false`). Đó là lý do PATCH dùng pointer (`*string`) để biết client có truyền hay không.
6. **`go.mod` tracking exact**: thêm import mới chạy `go mod tidy` để tải về (giống `npm install` nhưng tự động dựa trên code).
7. **Interface implicit**: bạn không khai báo `class X implements Y`. Chỉ cần struct có đủ method, Go tự nhận là implement interface đó. → Tách layer cực dễ.

---

## 8. Hướng mở rộng

- **Thay in-memory bằng Postgres**: viết `postgresRepo struct {}` implement `TodoRepository`, đổi `NewInMemoryTodoRepository` → `NewPostgresTodoRepository(db)` trong `main.go`. Không sửa usecase/handler.
- **Thêm auth**: viết middleware kiểm JWT, đặt trước route cần bảo vệ (`v1.Use(authMiddleware)`).
- **Config từ env**: dùng `github.com/spf13/viper` hoặc `github.com/caarlos0/env`.
- **Test**: Go có sẵn `testing` package — đặt file `*_test.go` cạnh code, chạy `go test ./...`. Mock repository dễ vì usecase phụ thuộc interface.
- **Production logger**: đổi `zap.NewDevelopmentConfig()` thành `zap.NewProductionConfig()` để xuất JSON gọn nhẹ.
