# Go-Saas

A modular and extensible Go SaaS framework that provides a solid foundation for building scalable and maintainable SaaS applications.

## Description

Go-Saas is a framework designed to simplify the development of SaaS applications in Go. It provides a clean architecture with clear separation of concerns, hooks for extensibility, and essential SaaS features like authentication, database abstraction, and email services.

## Features

- **Clean Architecture**: Well-organized codebase with separation of concerns
- **Authentication System**: Built-in JWT-based authentication
- **Database Abstraction**: Database-agnostic persistence layer
- **Extensibility**: Hook system for easy extension and customization
- **Configuration Management**: Flexible configuration via environment variables
- **API Server**: HTTP server with middleware support and route management
- **Email Service**: Integrated email functionality
- **Docker Support**: Ready-to-use Docker and docker-compose files

## Installation

### Using the CLI Tool

We provide a CLI tool to generate new Go-SaaS projects quickly:

```bash
# Install the CLI tool from the v2 branch
go install github.com/tuongaz/go-saas/tools/gosaas-cli@v2

# Create a new project
gosaas new my-saas-project
```

Alternatively, you can clone the repository and run the installation script:

```bash
git clone https://github.com/tuongaz/go-saas.git
git checkout v2
cd go-saas/tools/gosaas-cli
./install.sh
```

### Manual Setup

```bash
git clone https://github.com/tuongaz/go-saas.git
git checkout v2
cd go-saas
go mod tidy
```

## Architecture

Go-Saas follows a clean architecture pattern with the following components:

- **cmd**: Application entry points
- **config**: Configuration management
- **core**: Core application logic and bootstrap
- **pkg**: Shared utility packages
- **server**: HTTP server implementation
- **service**: Business logic services
- **store**: Data access layer

### Overall Architecture
<img src="./docs/overall_architecture.png" width="100%">

## Usage

The framework is designed to be extended through its hook system and interface-based components. Here are some common usage patterns:

### Registering Routes

```go
app.PublicRoute("/api/v1", func(r chi.Router) {
    r.Get("/health", healthCheck)
})

app.PrivateRoute("/api/v1/protected", func(r chi.Router) {
    r.Get("/profile", getProfile)
})
```

### Using Hooks

```go
app.OnBeforeServe().Register(func(event *core.OnBeforeServeEvent) error {
    fmt.Println("Server is about to start!")
    return nil
})
```

### Database Operations

```go
users := app.Store().Collection("users")
user := &User{ID: "123", Name: "John Doe"}
err := users.Create(ctx, user)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT