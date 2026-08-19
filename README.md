# eCommerce System in Go

A comprehensive eCommerce system built with Go, featuring RESTful API, MySQL database, and a clean layered architecture.

## Features

- **Product Management**: CRUD operations for products with inventory tracking
- **Shopping Cart**: Add, remove, and manage items in shopping cart
- **Order Processing**: Create orders, track status, and manage order lifecycle
- **Payment Processing**: Multiple payment methods with transaction tracking
- **Authentication**: Signup/login with bcrypt password hashing, JWT access tokens and rotating refresh tokens
- **User Management**: User registration and profile management
- **MySQL Database**: Persistent storage with GORM ORM
- **RESTful API**: Clean API endpoints using Gin framework
- **Swagger Documentation**: Interactive API documentation with Swagger UI
- **Environment Configuration**: Secure configuration management via .env file

## Project Structure

```
ecommerce/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── config/
│   └── config.go               # Configuration management
├── database/
│   ├── db.go                   # Database connection
│   └── migrate.go              # Database migrations
├── models/
│   ├── product.go              # Product model
│   ├── cart.go                 # Cart and CartItem models
│   ├── order.go                # Order and OrderItem models
│   ├── user.go                 # User model
│   └── payment.go              # Payment model
├── controllers/
│   ├── product_controller.go   # Product HTTP handlers
│   ├── cart_controller.go      # Cart HTTP handlers
│   └── order_controller.go     # Order HTTP handlers
├── services/
│   ├── product_service.go      # Product business logic
│   ├── order_service.go        # Order business logic
│   ├── payment_service.go     # Payment business logic
│   └── shipping_service.go    # Shipping business logic
├── routes/
│   └── routes.go               # API route configuration
├── docs/
│   ├── docs.go                 # Swagger generated docs
│   ├── swagger.json            # Swagger JSON specification
│   └── swagger.yaml            # Swagger YAML specification
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
├── env.example                 # Environment variables example
└── README.md                   # This file
```

## Prerequisites

- Go 1.25.5 or higher (see `go.mod`)
- MySQL 5.7 or higher (tested on 8.0)
- Git

## Setup Instructions

### 1. Clone the Repository

```bash
cd D:\projects\go-ecommerce
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure Database

Create a MySQL database:

```sql
CREATE DATABASE ecommerce;
```

### 4. Configure Environment Variables

Copy the example environment file:

```bash
copy env.example .env
```

Edit `.env` file with your database credentials:

```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password_here
DB_NAME=ecommerce

SERVER_HOST=localhost
SERVER_PORT=4444
```

### 5. Generate Swagger Documentation

```bash
swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
```

Or if swag is not in your PATH:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
$env:PATH += ";$env:GOPATH\bin"
swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
```

### 6. Run the Application

```bash
go run ./cmd/server
```

Or build and run:

```bash
go build -o server.exe ./cmd/server
./server.exe
```

The server will start on `http://localhost:4444`

## Swagger Documentation

The API includes interactive Swagger documentation. Once the server is running, access it at:

**Swagger UI**: http://localhost:4444/swagger/index.html

The Swagger documentation provides:
- Complete API endpoint documentation
- Request/response schemas
- Try-it-out functionality
- Model definitions

To regenerate Swagger docs after making changes:

```bash
swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
```

## API Endpoints

### Swagger Documentation
- `GET /swagger/index.html` - Interactive Swagger UI documentation

### Health Check
- `GET /health` - Health check endpoint

### Authentication
Public:
- `POST /api/auth/signup` - Register a new account, returns a token pair
- `POST /api/auth/login` - Log in, returns a token pair
- `POST /api/auth/refresh` - Exchange a refresh token for a new pair
- `POST /api/auth/logout` - Revoke one refresh token

Requires `Authorization: Bearer <access_token>`:
- `GET /api/auth/me` - Current user's profile
- `POST /api/auth/change-password` - Change password (signs out other sessions)
- `POST /api/auth/logout-all` - Revoke every session for the user

### Products
- `GET /api/products` - Get all products
- `GET /api/products/:id` - Get product by ID
- `POST /api/products` - Create a new product
- `PUT /api/products/:id` - Update a product
- `DELETE /api/products/:id` - Delete a product

### Cart
- `GET /api/carts/user/:user_id` - Get or create cart for user
- `GET /api/carts/:cart_id` - Get cart by ID
- `POST /api/carts/:cart_id/items` - Add item to cart
- `DELETE /api/carts/:cart_id/items/:item_id` - Remove item from cart

### Orders
- `POST /api/orders` - Create a new order
- `GET /api/orders/:id` - Get order by ID
- `GET /api/orders/user/:user_id` - Get all orders for a user
- `PUT /api/orders/:id/status` - Update order status
- `POST /api/orders/:id/cancel` - Cancel an order

## Authentication

### Configuration

`JWT_SECRET` is **required** — the server refuses to start without it, and there is
no insecure default to fall back on.

Generate one locally. Pick whichever command suits your shell:

```bash
# Git Bash, macOS, Linux
openssl rand -base64 48
```

```powershell
# Windows PowerShell (no openssl needed)
$bytes = New-Object byte[] 48; [System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($bytes); [Convert]::ToBase64String($bytes)
```

```bash
# Python
python -c "import secrets; print(secrets.token_urlsafe(48))"
```

Paste the result into `.env` unquoted:

```
JWT_SECRET=XyPuyb4ie8KEDfCBrvSjY+0ZIXg8l72jYYdZq2B0FJls6req+i0ovMdarYjjzk4h
```

Three rules for this value:

- **Generate it locally, never on a website.** This is a signing key: anyone who
  knows it can forge a valid token for any user, including one with
  `"role": "admin"`. A generator site cannot be audited and may log what it
  hands you. If you dislike the command line, use a password manager's
  random-password generator — it runs offline.
- **Generate your own; do not share or reuse one.** A shared secret means a token
  minted on one machine is accepted on every other. Dev and production must
  never use the same value.
- **Minimum 32 characters**, enforced at startup. The commands above produce 64.

| Variable | Default | Purpose |
|---|---|---|
| `JWT_SECRET` | *(required)* | Signs access tokens, minimum 32 characters |
| `JWT_ISSUER` | `go-ecommerce` | Expected `iss` claim |
| `JWT_ACCESS_TTL` | `15m` | Access token lifetime |
| `JWT_REFRESH_TTL` | `720h` | Refresh token lifetime (30 days) |
| `AUTH_MAX_FAILED_LOGINS` | `5` | Failed logins before lockout |
| `AUTH_LOCKOUT_DURATION` | `15m` | How long an account stays locked |

### Signing up and logging in

```bash
curl -X POST http://localhost:4444/api/auth/signup   -H "Content-Type: application/json"   -d '{"name":"Kiran Kumar","email":"kiran@example.com","password":"Str0ng!Passphrase"}'
```

Both signup and login return:

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "3Qk9m1s...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": { "id": 1, "email": "kiran@example.com", "role": "customer" }
}
```

### The sample users cannot log in

`database/seed.go` creates four sample users (`john.doe@example.com` and
friends) for the product and order examples. They were created before
authentication existed, so they have **no password** and login will always
return `401 invalid email or password`.

This is intentional rather than a bug: an account with an empty password hash is
rejected outright instead of being treated as "any password works". Sign up your
own account to test the auth endpoints.

Send the access token on protected requests:

```bash
curl http://localhost:4444/api/auth/me -H "Authorization: Bearer <access_token>"
```

When it expires after 15 minutes, exchange the refresh token for a new pair:

```bash
curl -X POST http://localhost:4444/api/auth/refresh   -H "Content-Type: application/json"   -d '{"refresh_token":"<refresh_token>"}'
```

**Refresh tokens rotate on every use.** The old one is invalidated immediately, so
always store the token returned by the most recent call. Presenting an
already-used token is treated as theft and revokes every session for that user.

### Using Swagger with authentication

1. Open http://localhost:4444/swagger/index.html
2. Call `POST /auth/signup` or `POST /auth/login` and copy the `access_token`
3. Click **Authorize** (top right) and paste **just the token** — the `Bearer `
   prefix is added for you by a `requestInterceptor` in
   [routes/swagger.go](routes/swagger.go). Typing `Bearer <token>` yourself also
   works and is not doubled up.
4. Endpoints marked with a padlock will now work

Authorization persists across page reloads, so you only paste once per token.

Note this is a UI convenience only. The request on the wire still carries the
standard `Authorization: Bearer <token>` header, and the server still rejects
anything without the scheme — see [middleware/auth.go](middleware/auth.go). If
you call the API with curl or from your own client, include `Bearer ` yourself.

### Password policy

At least 8 characters, no more than 72 bytes (bcrypt's limit), must contain a
letter plus a digit or symbol, must not be a common password, and must not echo
your own name or email.

### Security properties

| Concern | How it is handled |
|---|---|
| Password storage | bcrypt, cost 12; `PasswordHash` is `json:"-"` so it can never be serialized |
| User enumeration | Wrong password and unknown email return an identical 401, and a dummy bcrypt compare equalizes response time |
| Brute force (one account) | Lockout after `AUTH_MAX_FAILED_LOGINS` failures |
| Brute force (credential stuffing) | Per-IP rate limit of 20 requests/minute on the public auth endpoints |
| Stolen refresh token | Rotation on use plus replay detection that revokes the whole token family |
| Database dump | Only SHA-256 hashes of refresh tokens are stored |
| Algorithm confusion | Signing method pinned to HS256 at verification time |
| Privilege escalation | `role` is assigned server-side and never read from the request body |
| Error leakage | Only known sentinel errors reach the client; unexpected errors are logged and answered generically |

### Protecting your own routes

```go
// in routes/routes.go
admin := api.Group("/admin", requireAuth, middleware.RequireRole(models.RoleAdmin))
admin.POST("/products", productController.CreateProduct)
```

Inside a handler, read the caller's identity from the token, never from a
parameter the client controls:

```go
userID, ok := middleware.UserID(ctx)
```

Note that the existing product, cart and order routes are still public, so that
current clients keep working. Wrap them as shown above when you are ready.

## Example API Requests

### Create a Product

```bash
curl -X POST http://localhost:4444/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Laptop",
    "description": "High-performance laptop",
    "price": 999.99,
    "stock": 10,
    "category": "Electronics"
  }'
```

### Add Item to Cart

```bash
curl -X POST http://localhost:4444/api/carts/1/items \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 1,
    "quantity": 2
  }'
```

### Create an Order

```bash
curl -X POST http://localhost:4444/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "cart_id": 1,
    "shipping_cost": 15.00,
    "payment_method": "Credit Card"
  }'
```

## Database Schema

The application uses GORM for database management. Tables are automatically created on first run:

- `users` - User accounts
- `products` - Product catalog
- `carts` - Shopping carts
- `cart_items` - Items in shopping carts
- `orders` - Customer orders
- `order_items` - Items in orders
- `payments` - Payment transactions

## Architecture

The project follows a clean layered architecture:

1. **Models**: Database entities with GORM tags
2. **Services**: Business logic layer
3. **Controllers**: HTTP request handlers
4. **Routes**: API endpoint configuration
5. **Database**: Database connection and migrations
6. **Config**: Configuration management

## Development

### Running Migrations

Migrations run automatically on application startup. To manually run migrations:

```go
// In main.go or a migration script
database.Migrate()
```

### Adding New Features

1. Create model in `models/` package
2. Add service methods in `services/` package
3. Create controller handlers in `controllers/` package
4. Add routes in `routes/routes.go`

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `3306` |
| `DB_USER` | Database user | `root` |
| `DB_PASSWORD` | Database password | (required) |
| `DB_NAME` | Database name | `ecommerce` |
| `SERVER_HOST` | Server host | `localhost` |
| `SERVER_PORT` | Server port | `4444` |

## Dependencies

- **Gin**: HTTP web framework
- **GORM**: ORM library for Go
- **MySQL Driver**: MySQL database driver
- **godotenv**: Environment variable management

## Future Enhancements

- Authentication and authorization (JWT)
- User registration and login endpoints
- Product reviews and ratings
- Discount codes and promotions
- Email notifications
- Admin dashboard
- Unit and integration tests
- Docker containerization
- CI/CD pipeline

## License

This is a demonstration project for educational purposes.
