package infrastructure

import (
	"fmt"
	"log"
	"os"
	"time"
	"tridorian-ztna/internal/models"
	"tridorian-ztna/pkg/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupDatabase() *gorm.DB {

	dbHost := utils.GetEnv("DB_HOST", "localhost")
	dbPort := utils.GetEnv("DB_PORT", "5432")
	dbUser := utils.GetEnv("DB_USER", "devuser")
	dbPass := utils.GetEnv("DB_PASSWORD", "P@ssw0rd")
	dbName := utils.GetEnv("DB_NAME", "trivpn-trimanaged")
	appEnv := utils.GetEnv("APP_ENV", "development")

	sslMode := "disable"
	logLevel := logger.Info

	if appEnv == "production" {
		sslMode = "require"
		logLevel = logger.Error
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Bangkok",
		dbHost, dbUser, dbPass, dbName, dbPort, sslMode,
	)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  appEnv != "production",
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:      newLogger,
		PrepareStmt: true,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get SQL DB object: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if appEnv != "production" {

		all_model := []any{&models.Tenant{},
			&models.Administrator{},
			&models.NodeSku{},
			&models.Node{},
			&models.PolicyNode{},
			&models.PolicyCondition{},
			&models.AccessPolicy{},
			&models.SignInPolicy{},
			&models.Application{},
			&models.ApplicationCIDR{},
			&models.CustomDomain{},
			&models.BackofficeUser{},
			&models.AccessPolicyNode{},
		}

		// db.Migrator().DropTable(all_model...)

		err = db.AutoMigrate(
			all_model...,
		)
		if err != nil {
			log.Fatalf("Failed to get SQL DB object: %v", err)
		}

		// backofficeEmail := utils.GetEnv("BACKOFFICE_USER", "tri")
		// backofficePass := utils.GetEnv("BACKOFFICE_PASSWORD", "password")
		// backofficeUser := &models.BackofficeUser{
		// 	Name:  "System Admin",
		// 	Email: backofficeEmail,
		// }
		// backofficeUser.SetPassword(backofficePass)
		// db.Create(backofficeUser)
		// log.Printf("🔑 Backoffice created: User=%s, Pass=%s", backofficeEmail, backofficePass)

		// nodeSku := &models.NodeSku{
		// 	Name:        "Trial SKU",
		// 	Description: "Trial SKU for testing purposes",
		// 	MaxUsers:    100,
		// 	Bandwidth:   10,
		// 	PriceCents:  0,
		// }
		// db.Create(nodeSku)

	}

	return db
}
