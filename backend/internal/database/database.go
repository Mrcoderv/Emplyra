package database

import (
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/emplyra/backend/internal/config"
	"github.com/emplyra/backend/internal/models"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	logLevel := logger.Warn
	if cfg.Environment == "development" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			slog.NewLogLogger(slog.Default().Handler(), slog.LevelDebug),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logLevel,
				IgnoreRecordNotFoundError: true,
				ParameterizedQueries:      true,
			},
		),
	})
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(allModels()...)
}

func allModels() []interface{} {
	return []interface{}{
		&models.Permission{},
		&models.Role{},
		&models.User{},
		&models.RefreshToken{},
		&models.AuditLog{},
		&models.Department{},
		&models.Designation{},
		&models.Employee{},
		&models.SalaryStructure{},
		&models.Attendance{},
		&models.LeaveType{},
		&models.LeaveBalance{},
		&models.Leave{},
		&models.Holiday{},
		&models.Payroll{},
		&models.JobPost{},
		&models.Candidate{},
		&models.Application{},
		&models.Interview{},
		&models.Onboarding{},
		&models.GoogleFormIntegration{},
		&models.GoogleFormResponse{},
		&models.GoogleOAuthToken{},
		&models.Goal{},
		&models.KPI{},
		&models.PerformanceReview{},
		&models.TrainingProgram{},
		&models.TrainingSchedule{},
		&models.TrainingEnrollment{},
		&models.Document{},
		&models.Notification{},
	}
}
