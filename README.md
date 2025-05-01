# ms-tester

`ms-tester` is a microservice designed for managing programming competitions. It provides backend functionality for handling problems, contests, participants, and their submissions. The service is developed in Go. PostgreSQL serves as the relational database. Pandoc is used to convert problem statements from LaTeX to HTML.

For understanding the architecture, see the [documentation](https://github.com/Vyacheslav1557/docs).

### Prerequisites

Before you begin, ensure you have the following dependencies installed:

* **Docker** and **Docker Compose**: To run PostgreSQL, Pandoc.
* **Goose**: For applying database migrations (`go install github.com/pressly/goose/v3/cmd/goose@latest`).
* **oapi-codegen**: For generating OpenAPI code (`go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest`).

### 1. Running Dependencies

You can run PostgreSQL and Pandoc using Docker Compose, for example:

```yaml
version: '3.8'
services:
  pandoc:
    image: pandoc/latex
    ports:
      - "4000:3030"  # Exposes Pandoc server on port 4000 locally
    command: "server" # Runs Pandoc in server mode
  postgres:
    image: postgres:14.1-alpine # Uses PostgreSQL 14.1 Alpine image
    restart: always # Ensures the container restarts if it stops
    environment:
      POSTGRES_USER: postgres # Default user
      POSTGRES_PASSWORD: supersecretpassword # Default password (change for production!)
      POSTGRES_DB: postgres # Default database name
    ports:
      - '5432:5432' # Exposes PostgreSQL on the standard port 5432
    volumes:
      - ./postgres-data:/var/lib/postgresql/data # Persists database data locally
    healthcheck:
      test: pg_isready -U postgres -d postgres # Command to check if PostgreSQL is ready
      interval: 10s # Check every 10 seconds
      timeout: 3s # Wait 3 seconds for the check to respond
      retries: 5 # Try 5 times before marking as unhealthy
volumes:
  postgres-data: # Defines the named volume for data persistence
```

Start the services in detached mode:

```bash
docker-compose up -d
```

### 2. Configuration

The application uses environment variables for configuration. Create a `.env` file in the project root. The minimum required variables are:

```dotenv
# Environment type (development or production)
ENV=dev # or prod

# Address of the running Pandoc service
PANDOC=http://localhost:4000

# Address and port where the ms-tester service will listen
ADDRESS=localhost:8080

# PostgreSQL connection string (Data Source Name)
# Format: postgres://user:password@host:port/database?sslmode=disable
POSTGRES_DSN=postgres://username:supersecretpassword@localhost:5432/db_name?sslmode=disable

# Secret key for signing and verifying JWT tokens
JWT_SECRET=your_super_secret_jwt_key
```

**Important:** Replace `supersecretpassword` and `your_super_secret_jwt_key` with secure, unique values, especially for a production environment.

### 3. Database Migrations

The project uses `goose` to manage the database schema.

1.  Ensure `goose` is installed:
    ```bash
    go install github.com/pressly/goose/v3/cmd/goose@latest
    ```
2.  Apply the migrations to the running PostgreSQL database. Make sure the connection string in the command matches the `POSTGRES_DSN` from your `.env` file:
    ```bash
    goose -dir ./migrations postgres "postgres://postgres:supersecretpassword@localhost:5432/postgres?sslmode=disable" up
    ```

### 4. OpenAPI Code Generation

The project uses OpenAPI to define its API. Go code for handlers and models is generated based on this specification using `oapi-codegen`.

Run the generation command:

```bash
make gen
```

### 5. Running the Application

Start the `ms-tester` service:

```bash
go run ./main.go
```

After starting, the service will be available at the address specified in the `ADDRESS` variable in your `.env` file (e.g., `http://localhost:8080`).