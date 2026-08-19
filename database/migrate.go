package database

import (
	"fmt"
	"log"

	"go-ecommerce/models"
)

// SeedIfEmpty seeds the database if it's empty
func SeedIfEmpty() {
	var count int64
	DB.Model(&models.Product{}).Count(&count)
	if count == 0 {
		if err := SeedDatabase(); err != nil {
			log.Printf("Warning: Failed to seed database: %v", err)
		}
	} else {
		log.Println("Database already contains data, skipping seed.")
	}
}

// Migrate runs database migrations
func Migrate() error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	log.Println("Running database migrations...")

	// Set default storage engine to InnoDB for this session
	sqlDB, err := DB.DB()
	if err == nil {
		sqlDB.Exec("SET default_storage_engine=InnoDB")
	}

	err = DB.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.Product{},
		&models.Cart{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.Payment{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Convert existing MyISAM tables to InnoDB and ensure all tables use InnoDB
	tables := []string{"users", "refresh_tokens", "products", "carts", "cart_items", "orders", "order_items", "payments"}
	for _, table := range tables {
		// Check current engine and convert to InnoDB if needed
		var engine string
		result := DB.Raw("SELECT ENGINE FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?", table).Scan(&engine)
		if result.Error == nil && engine != "" && engine != "InnoDB" {
			if err := DB.Exec(fmt.Sprintf("ALTER TABLE `%s` ENGINE=InnoDB", table)).Error; err != nil {
				log.Printf("Warning: Failed to convert table %s to InnoDB: %v", table, err)
			} else {
				log.Printf("Converted table %s from %s to InnoDB", table, engine)
			}
		}
	}

	log.Println("Database migrations completed successfully (all tables using InnoDB)")
	return nil
}
