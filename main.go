package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/YakovlevIgA/forozon/graph/repository"
	"github.com/jackc/pgx/v4"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/YakovlevIgA/forozon/graph"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

const defaultPort = "8080"

func main() {
	ctx := context.Background()

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	storageType := os.Getenv("STORAGE")
	var storage graph.Storage

	// Инициализация репозитория нужного типа
	switch storageType {
	case "postgres":
		storage = initPG(ctx)
		log.Println("Используется postgres хранилище")
	default:
		storage = repository.NewInMemoryRepository()
		log.Println("Используется in-memory хранилище")
	}

	// Инициализация сервиса
	resolver := graph.NewResolver(storage)

	// Инициализация GraphQL сервера и playground для него
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.Use(extension.Introspection{})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("Подключитесь к http://localhost:%s/ для GraphQL Playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// initPG инициализация postgres
func initPG(ctx context.Context) graph.Storage {
	connStr := os.Getenv("POSTGRES_URL")

	if connStr == "" {
		log.Fatal("POSTGRES_URL не указан")
	}

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("не удалось подключиться к базе данных: %v", err)
	}

	if err := conn.Ping(ctx); err != nil {
		log.Fatalf("не удалось подключиться к базе данных: %v", err)
	}

	log.Println("Подключились к postgres")

	if err := runMigrations(); err != nil {
		log.Fatalf("Не удалось применить миграции: %v", err)
	}

	log.Println("Миграции успешно выполнены")

	storage, err := repository.NewPostgresRepository(conn)
	if err != nil {
		log.Fatalf("Ошибка инициализации PostgreSQL: %v", err)
	}

	return storage
}

// runMigrations применение миграций к postgres
func runMigrations() error {
	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		log.Fatalf("Ошибка при подключении к базе данных: %v", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs("./migrations")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", filepath.ToSlash(absPath)),
		os.Getenv("POSTGRES_DB"),
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
