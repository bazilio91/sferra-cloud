# sferra-cloud

Cloud layer for calculating the cost of manufacturing parts.

## Prerequisites

- Go 1.23.1 or higher
- Docker and Docker Compose
- Protocol Buffers compiler (`protoc`)
- Make sure `protoc` plugins for Go are installed:
    - `protoc-gen-go`
    - `protoc-gen-go-grpc`

## Installation

1. Clone the repository:

```bash
git clone https://github.com/bazilio91/sferra-cloud.git
cd sferra-cloud
```

1. Install the Go dependencies:

```bash
go mod download
```

1. Install tools:

```bash
# Install protobuf compiler (protoc) if not already installed
# For MacOS (using Homebrew)
brew install protobuf

# For Ubuntu
sudo apt-get install -y protobuf-compiler

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Install swag for Swagger documentation
go install github.com/swaggo/swag/cmd/swag@latest
```

1. Create .env file:

```bash
cp .env.example .env
```


## Running the Application in Development Mode

```bash
# Generate self-signed certificates (for development purposes only)
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes -subj "/CN=localhost"
# Start the services
docker-compose up -d

# Generate Protobuf Files:
make proto

# Generate Swagger Documentation:
make swagger

# Run the REST API Server
go run cmd/api/main.go

# Run the gRPC Server
go run cmd/grpc/main.go
```

## Running Tests
    
```bash
go test ./...
```

## Directory Structure
- cmd/: Contains the entry points of the application.
  - api/: Entry point for the REST API server.
  - grpc/: Entry point for the gRPC server.
- pkg/: Contains the application packages.
  - api/: REST API server code.
  - grpc/: gRPC server code.
  - auth/: JWT authentication utilities.
  - config/: Configuration loading and validation.
  - db/: Database initialization.
  - models/: Database models.
  - pb/: Protobuf definitions and generated code.
