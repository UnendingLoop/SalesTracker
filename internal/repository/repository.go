package repository

import (
	"context"
	"database/sql"
	"log"
	"path/filepath"
	"time"

	"github.com/UnendingLoop/SalesTracker/internal/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/dbpg"
)

type OperationsRepository interface {
	Create(ctx context.Context, op *model.Operation) error
	Get(ctx context.Context, id int) (*model.Operation, error)
	List(ctx context.Context, f *model.RequestParamOperations) ([]model.Operation, error)
	Update(ctx context.Context, op *model.Operation) error
	Delete(ctx context.Context, id int) error
	AnalyticsGroup(ctx context.Context, f *model.RequestParamAnalytics) ([]model.AnalyticsQuantum, error)
	AnalyticsSummary(ctx context.Context, f *model.RequestParamAnalytics) (*model.AnalyticsSummary, error)
}

func NewOperationsRepo(dbconn *dbpg.DB) OperationsRepository {
	return &PostgresRepo{db: dbconn}
}

func ConnectWithRetries(appConfig *config.Config, retryCount int, idleTime time.Duration) *dbpg.DB {
	dbOptions := dbpg.Options{
		MaxOpenConns:    5,
		MaxIdleConns:    5,
		ConnMaxLifetime: 10 * time.Minute,
	}

	dbUser := appConfig.GetString("POSTGRES_USER")
	dbName := appConfig.GetString("POSTGRES_DB")
	dbPass := appConfig.GetString("POSTGRES_PASSWORD")
	dbContName := appConfig.GetString("DB_CONTAINER_NAME")
	if dbUser == "" || dbName == "" || dbPass == "" || dbContName == "" {
		log.Fatal("DB connection credentials, db name or DB container name are not set in env")
	}
	dsn := "postgresql://" + dbUser + ":" + dbPass + "@" + dbContName + ":5432/" + dbName + "?sslmode=disable"

	var dbConn *dbpg.DB
	var err error

	for range retryCount {
		dbConn, err = dbpg.New(dsn, nil, &dbOptions)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to PGDB: %s\nWaiting %v before next retry...", err, idleTime)
		time.Sleep(idleTime)
	}

	if err != nil {
		log.Fatal("Failed to connect to DB. Exiting the app...")
	}

	return dbConn
}

func MigrateWithRetries(db *sql.DB, migrationsPath string, retries int, idle time.Duration) {
	for i := range retries {
		log.Printf("Migration try #%d...", i)
		err := runMigrate(db, migrationsPath)
		if err == nil {
			break
		}
		switch i {
		case retries:
			log.Fatalln("Out of retries. Exiting...")
		default:
			log.Printf("Migration try #%d was unsuccessful: %v\nWaiting %v before next try...", i, err, idle)
			time.Sleep(idle)
		}
	}
}

func runMigrate(db *sql.DB, migrationsPath string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return err
	}

	sourceURL := "file://" + absPath
	log.Println("Running migrations from:", sourceURL)

	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Database migrations applied successfully")
	return nil
}
