package repository

import (
	"fmt"

	"github.com/bernardofernandezz/scheduling-api/internal/config"
	"github.com/bernardofernandezz/scheduling-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Repositories holds all repositories
type Repositories struct {
	db               *gorm.DB
	UserRepo         UserRepository
	SupplierRepo     SupplierRepository
	EmployeeRepo     EmployeeRepository
	ProductRepo      ProductRepository
	OperationRepo    OperationRepository
	AppointmentRepo  AppointmentRepository
	AvailabilityRepo AvailabilityRepository
}

// NewDBConnection creates a new database connection
func NewDBConnection(config config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Name, config.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// NewRepositories creates new instances of all repositories
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		db:               db,
		UserRepo:         NewUserRepository(db),
		SupplierRepo:     NewSupplierRepository(db),
		EmployeeRepo:     NewEmployeeRepository(db),
		ProductRepo:      NewProductRepository(db),
		OperationRepo:    NewOperationRepository(db),
		AppointmentRepo:  NewAppointmentRepository(db),
		AvailabilityRepo: NewAvailabilityRepository(db),
	}
}

// AutoMigrate migrates all models
func (r *Repositories) AutoMigrate() error {
	return r.db.AutoMigrate(
		&models.User{},
		&models.Supplier{},
		&models.Employee{},
		&models.Product{},
		&models.Operation{},
		&models.Appointment{},
		&models.AvailabilitySlot{},
	)
}

// GetDB returns the database instance
func (r *Repositories) GetDB() *gorm.DB {
	return r.db
}

