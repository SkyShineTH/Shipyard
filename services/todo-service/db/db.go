package db

import (
	"fmt"
	"os"

	"github.com/SkyShineTH/shipyard/todo-service/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// migrateTodoUserID handles existing rows before AutoMigrate adds NOT NULL user_id.
// Old todos get user_id = 0 (orphaned; JWT users start at 1).
func migrateTodoUserID(db *gorm.DB) error {
	var tableExists bool
	if err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = current_schema() AND table_name = 'todos'
		)`).Scan(&tableExists).Error; err != nil {
		return fmt.Errorf("check todos table: %w", err)
	}
	if !tableExists {
		return nil
	}

	var colExists bool
	if err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = current_schema() AND table_name = 'todos' AND column_name = 'user_id'
		)`).Scan(&colExists).Error; err != nil {
		return fmt.Errorf("check user_id column: %w", err)
	}

	if !colExists {
		if err := db.Exec(`ALTER TABLE todos ADD COLUMN user_id bigint NOT NULL DEFAULT 0`).Error; err != nil {
			return fmt.Errorf("add user_id: %w", err)
		}
		if err := db.Exec(`ALTER TABLE todos ALTER COLUMN user_id DROP DEFAULT`).Error; err != nil {
			return fmt.Errorf("drop user_id default: %w", err)
		}
		return db.Exec(`CREATE INDEX IF NOT EXISTS idx_todos_user_id ON todos (user_id)`).Error
	}

	if err := db.Exec(`UPDATE todos SET user_id = 0 WHERE user_id IS NULL`).Error; err != nil {
		return fmt.Errorf("backfill user_id: %w", err)
	}
	return db.Exec(`ALTER TABLE todos ALTER COLUMN user_id SET NOT NULL`).Error
}

func Connect() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
	)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	if err := migrateTodoUserID(database); err != nil {
		return nil, fmt.Errorf("migrate user_id: %w", err)
	}

	if err := database.AutoMigrate(&model.Todo{}); err != nil {
		return nil, fmt.Errorf("run automigrate: %w", err)
	}

	return database, nil
}
