package service

import (
	"errors"
	"time"

	"github.com/bernardofernandezz/scheduling-api/internal/models"
	"github.com/bernardofernandezz/scheduling-api/internal/repository"
)

// AppointmentService interface defines methods for appointment service
type AppointmentService interface {
	Create(appointment *models.Appointment) error
	GetByID(id uint) (*models.Appointment, error)
	Update(appointment *models.Appointment) error
	Delete(id uint) error
	List(filters repository.AppointmentFilters) ([]models.Appointment, int64, error)
	UpdateStatus(id uint, status models.AppointmentStatus, reason string) error
	GetBySupplier(supplierID uint, filters repository.AppointmentFilters) ([]models.Appointment, int64, error)
	GetByEmployee(employeeID uint, filters repository.AppointmentFilters) ([]models.Appointment, int64, error)
	GetByOperation(operationID uint, filters repository.AppointmentFilters) ([]models.Appointment, int64, error)
	GetByDateRange(start, end time.Time, filters repository.AppointmentFilters) ([]models.Appointment, int64, error)
	GetUpcoming(limit int) ([]models.Appointment, error)
	GetStatistics() (*repository.AppointmentStatistics, error)
	CheckAvailability(operationID, employeeID uint, start, end time.Time) (bool, error)
}

// appointmentService implements AppointmentService interface
type appointmentService struct {
	appointmentRepo repository.AppointmentRepository
	employeeRepo    repository.EmployeeRepository
	supplierRepo    repository.SupplierRepository
	operationRepo   repository.OperationRepository
	productRepo     repository.ProductRepository
}

// NewAppointmentService creates a new appointment service
func NewAppointmentService(
	appointmentRepo repository.AppointmentRepository,
	employeeRepo repository.EmployeeRepository,
	supplierRepo repository.SupplierRepository,
	operationRepo repository.OperationRepository,
	productRepo repository.ProductRepository,
) AppointmentService {
	return &appointmentService{
		appointmentRepo: appointmentRepo,
		employeeRepo:    employeeRepo,
		supplierRepo:    supplierRepo,
		operationRepo:   operationRepo,
		productRepo:     productRepo,
	}
}

// Create creates a new appointment
func (s *appointmentService) Create(appointment *models.Appointment) error {
	// Check if supplier exists
	_, err := s.supplierRepo.FindByID(appointment.SupplierID)
	if err != nil {
		return errors.New("invalid supplier: " + err.Error())
	}

	// Check if employee exists
	_, err = s.employeeRepo.FindByID(appointment.EmployeeID)
	if err != nil {
		return errors.New("invalid employee: " + err.Error())
	}

	// Check if operation exists
	operation, err := s.operationRepo.FindByID(appointment.OperationID)
	if err != nil {
		return errors.New("invalid operation: " + err.Error())
	}

	// Check if product exists
	_, err = s.productRepo.FindByID(appointment.ProductID)
	if err != nil {
		return errors.New("invalid product: " + err.Error())
	}

	// Check if appointment is within operation hours
	startTime := appointment.ScheduledStart
	endTime := appointment.ScheduledEnd
	
	// Extract just the time portion for comparison
	startTimeOfDay := time.Date(2000, 1, 1, startTime.Hour(), startTime.Minute(), 0, 0, startTime.Location())
	endTimeOfDay := time.Date(2000, 1, 1, endTime.Hour(), endTime.Minute(), 0, 0, endTime.Location())
	operationOpen := time.Date(2000, 1, 1, operation.OpeningHour.Hour(), operation.OpeningHour.Minute(), 0, 0, operation.OpeningHour.Location())
	operationClose := time.Date(2000, 1, 1, operation.ClosingHour.Hour(), operation.ClosingHour.Minute(), 0, 0, operation.ClosingHour.Location())
	
	if startTimeOfDay.Before(operationOpen) || endTimeOfDay.After(operationClose) {
		return errors.New("appointment must be within operation hours")
	}

	// Set default status if not provided
	if appointment.Status == "" {
		appointment.Status = models.StatusPending
	}

	// Create appointment
	return s.appointmentRepo.Create(appointment)
}

// GetByID gets an appointment by ID
func (s *appointmentService) GetByID(id uint) (*models.Appointment, error) {
	return s.appointmentRepo.FindByID(id)
}

// Update updates an appointment
func (s *appointmentService) Update(appointment *models.Appointment) error {
	// Check if appointment exists
	existing, err := s.appointmentRepo.FindByID(appointment.ID)
	if err != nil {
		return err
	}

	// Check if status allows updates
	if existing.Status == models.StatusCancelled || existing.Status == models.StatusCompleted {
		return errors.New("cannot update cancelled or completed appointments")
	}

	// Check if supplier exists
	_, err = s.supplierRepo.FindByID(appointment.SupplierID)
	if err != nil {
		return errors.New("invalid supplier: " + err.Error())
	}

	// Check if employee exists
	_, err = s.employeeRepo.FindByID(appointment.EmployeeID)
	if err != nil {
		return errors.New("invalid employee: " + err.Error())
	}

	// Check if operation exists
	operation, err := s.operationRepo.FindByID(appointment.OperationID)
	if err != nil {
		return errors.New("invalid operation: " + err.Error())
	}

	// Check if product exists
	_, err = s.productRepo.FindByID(appointment.ProductID)
	if err != nil {
		return errors.New("invali

