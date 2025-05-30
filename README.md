
# Scheduling API

A robust Go API for a scheduling portal where suppliers can schedule appointments with employees for product deliveries.

## 🚀 Project Overview

This API provides functionality for managing a scheduling system for suppliers to deliver products to company operations. Key features include:

- ✅ User authentication and authorization
- ✅ Appointment scheduling with conflict detection
- ✅ Role-based access control
- ✅ Comprehensive appointment management
- ✅ Statistics and reporting
- ✅ Availability checking

## 🛠️ Technologies

- Go (Golang) 1.20+
- Gin Web Framework
- GORM (with PostgreSQL)
- JWT Authentication
- Clean Architecture Pattern

## 🗂️ Project Structure

\`\`\`
schedulingAPI/
├── cmd/
│   └── api/
│       └── main.go           # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/         # HTTP request handlers
│   │   ├── middleware/       # HTTP middleware
│   │   └── routes/           # Route definitions
│   ├── config/               # Application configuration
│   ├── models/               # Domain models
│   ├── repository/           # Data access layer
│   └── service/              # Business logic layer
├── pkg/
│   ├── auth/                 # Authentication utilities
│   └── utils/                # Common utilities
├── scripts/                  # Build and deployment scripts
├── go.mod                    # Go module definition
└── README.md                 # Project documentation
\`\`\`

## 🏁 Getting Started

### Prerequisites

- Go 1.20 or higher
- PostgreSQL database
- Git

### Installation

1. Clone the repository:

\`\`\`bash
git clone https://github.com/bernardofernandezz/scheduling-api.git
cd scheduling-api
\`\`\`

2. Install dependencies:

\`\`\`bash
go mod download
\`\`\`

3. Create a \`.env\` file in the root directory:

\`\`\`
# Server settings
SERVER_ADDRESS=:8080
GIN_MODE=debug

# Database settings
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=scheduling_db
DB_SSLMODE=disable

# JWT settings
JWT_SECRET=your-secret-key
JWT_EXPIRE_HOURS=24
\`\`\`

4. Run the application:

\`\`\`bash
go run cmd/api/main.go
\`\`\`

### Using Convenience Scripts

- For Unix/Linux/MacOS:

\`\`\`bash
./scripts/run.sh
\`\`\`

- For Windows:

\`\`\`powershell
.\scriptsun.ps1
\`\`\`

## 🐳 Docker Support

You can also run the application using Docker:

1. Build the Docker image:

\`\`\`bash
docker build -t scheduling-api .
\`\`\`

2. Run the container:

\`\`\`bash
docker run -p 8080:8080 scheduling-api
\`\`\`

For production deployment, use environment variables:

\`\`\`bash
docker run -p 8080:8080 \
  -e SERVER_ADDRESS=:8080 \
  -e GIN_MODE=release \
  -e DB_HOST=your-db-host \
  -e DB_PORT=5432 \
  -e DB_USER=your-db-user \
  -e DB_PASSWORD=your-db-password \
  -e DB_NAME=scheduling_db \
  -e DB_SSLMODE=require \
  -e JWT_SECRET=your-very-secure-jwt-secret \
  -e JWT_EXPIRE_HOURS=24 \
  -e CORS_ALLOWED_ORIGINS=https://your-frontend-domain.com \
  scheduling-api
\`\`\`

## 📚 API Endpoints

### Authentication

- \`POST /api/auth/register\` - Register a new user
- \`POST /api/auth/login\` - Authenticate and get tokens
- \`POST /api/auth/refresh\` - Refresh authentication token
- \`POST /api/auth/password-reset\` - Request password reset

### Users

- \`GET /api/users/profile\` - Get authenticated user profile
- \`POST /api/users/change-password\` - Change user password

### Appointments

- \`POST /api/appointments\` - Create a new appointment
- \`GET /api/appointments\` - List appointments with filters
- \`GET /api/appointments/:id\` - Get appointment details
- \`PUT /api/appointments/:id\` - Update an appointment
- \`DELETE /api/appointments/:id\` - Delete an appointment
- \`POST /api/appointments/:id/status\` - Update appointment status
- \`POST /api/appointments/check-availability\` - Check time slot availability
- \`GET /api/appointments/upcoming\` - Get upcoming appointments
- \`GET /api/appointments/by-date-range\` - Get appointments within date range
- \`GET /api/appointments/by-supplier/:supplier_id\` - Get supplier appointments
- \`GET /api/appointments/by-employee/:employee_id\` - Get employee appointments
- \`GET /api/appointments/by-operation/:operation_id\` - Get operation appointments

### Admin

- \`GET /api/admin/statistics/appointments\` - Get appointment statistics

## 🔐 Authentication

The API uses JWT (JSON Web Token) for authentication. To access protected endpoints:

1. Obtain a token via the login endpoint.
2. Include the token in the Authorization header:

\`\`\`
Authorization: Bearer <your-token>
\`\`\`

## 🏷️ Models

### User

Roles:

- Admin: System administrator
- Employee: Company employee who receives deliveries
- Supplier: External supplier who schedules deliveries

### Appointment

Statuses:

- Pending: Initial state when created
- Confirmed: Approved by employee
- Cancelled: Cancelled by either party
- Completed: Delivery successfully completed
- Rescheduled: Appointment time changed

## 📝 License

© 2025 Your Company Name. All rights reserved.

## 🛠️ Development Setup

### With Docker

```bash
docker-compose up -d
docker-compose logs -f api
docker-compose down
```

### Manual Setup

1. Ensure Go 1.20+ is installed.
2. Create a PostgreSQL database.
3. Copy \`.env.example\` to \`.env\` and configure.
4. Run:

```bash
go mod download
go run ./cmd/api
```

or

```bash
make run
```

### Using Makefile

```bash
make build
make run
make test
make test-coverage
make docker
```

## 🚀 Deployment

### Docker Image

```bash
docker build -t scheduling-api .
docker run -p 8080:8080 \
  -e DB_HOST=your-db-host \
  -e DB_PORT=5432 \
  -e DB_USER=your-db-user \
  -e DB_PASSWORD=your-db-password \
  -e DB_NAME=scheduling_db \
  -e DB_SSLMODE=require \
  -e JWT_SECRET=your-secret-key \
  -e GIN_MODE=release \
  scheduling-api
```

### CI/CD

Includes GitHub Actions for:

- Automated testing
- Docker image build/push
- Automated deployment

## 🤝 Contributing

1. Fork the repository.
2. Create your feature branch.
3. Commit changes.
4. Push to branch.
5. Open a Pull Request.
