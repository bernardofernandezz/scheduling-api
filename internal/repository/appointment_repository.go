		Preload("Employee").Preload("Employee.User").
		Preload("Operation").Preload("Product").
		Find(&appointments).Error

	if err != nil {
		return nil, 0, err
	}

	return appointments, count, nil
}

// FindByOperation finds appointments by operation
func (r *appointmentRepository) FindByOperation(operationID uint, filters AppointmentFilters) ([]models.Appointment, int64, error) {
	var appointments []models.Appointment
	var count int64

	// Base query with operation filter
	query := r.db.Model(&models.Appointment{}).Where("operation_id = ?", operationID)

	// Apply other filters
	query = applyAppointmentFilters(query, filters)

	// Count total records
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	if filters.Page > 0 && filters.Limit > 0 {
		offset := (filters.Page - 1) * filters.Limit
		query = query.Offset(offset).Limit(filters.Limit)
	}

	// Apply sorting
	if filters.SortBy != "" {
		sortOrder := "ASC"
		if filters.SortOrder == "desc" {
			sortOrder = "DESC"
		}
		query = query.Order(filters.SortBy + " " + sortOrder)
	} else {
		// Default sorting by scheduled start time
		query = query.Order("scheduled_start ASC")
	}

	// Fetch appointments with preloaded relations
	err := query.Preload("Supplier").Preload("Supplier.User").
		Preload("Employee").Preload("Employee.User").
		Preload("Operation").Preload("Product").
		Find(&appointments).Error

	if err != nil {
		return nil, 0, err
	}

	return appointments, count, nil
}

// FindByDateRange finds appointments within a date range
func (r *appointmentRepository) FindByDateRange(start, end time.Time, filters AppointmentFilters) ([]models.Appointment, int64, error) {
	var appointments []models.Appointment
	var count int64

	// Base query with date range filter
	query := r.db.Model(&models.Appointment{}).
		Where("scheduled_start >= ? AND scheduled_start <= ?", start, end)

	// Apply other filters
	query = applyAppointmentFilters(query, filters)

	// Count total records
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	if filters.Page > 0 && filters.Limit > 0 {
		offset := (filters.Page - 1) * filters.Limit
		query = query.Offset(offset).Limit(filters.Limit)
	}

	// Apply sorting
	if filters.SortBy != "" {
		sortOrder := "ASC"
		if filters.SortOrder == "desc" {
			sortOrder = "DESC"
		}
		query = query

package repository

import (
    "time"
    "github.com/bernardofernandezz/scheduling-api/internal/models"
    "gorm.io/gorm"
)

// AppointmentFilters defines filters for querying appointments
type AppointmentFilters struct {
    Status     *models.AppointmentStatus
    StartDate  *time.Time
    EndDate    *time.Time
    Page       int
    Limit      int
    SortBy     string
    SortOrder  string
}

// AppointmentRepository interface defines methods for appointment repository
type AppointmentRepository interface {
    Repository
    FindBySupplier(supplierID uint, filters AppointmentFilters) ([]models.Appointment, int64, error)
    FindByEmployee(employeeID uint, filters AppointmentFilters) ([]models.Appointment, int64, error)
    FindByOperation(operationID uint, filters AppointmentFilters) ([]models.Appointment, int64, error)
    FindByDateRange(start, end time.Time, filters AppointmentFilters) ([]models.Appointment, int64, error)
    FindUpcoming(limit int) ([]models.Appointment, error)
    HasConflict(appointment *models.Appointment) (bool, error)
    UpdateStatus(id uint, status models.AppointmentStatus, reason string) error
    List(filters AppointmentFilters) ([]models.Appointment, int64, error)
}

// appointmentRepository implements AppointmentRepository interface
type appointmentRepository struct {
    BaseRepository
}

// NewAppointmentRepository creates a new appointment repository
func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
    return &appointmentRepository{
        BaseRepository: NewBaseRepository(db),
    }
}

// List returns a paginated list of appointments
func (r *appointmentRepository) List(filters AppointmentFilters) ([]models.Appointment, int64, error) {
    var appointments []models.Appointment
    var total int64

    query := r.db.Model(&models.Appointment{})

    // Apply filters
    if filters.Status != nil {
        query = query.Where("status = ?", *filters.Status)
    }
    if filters.StartDate != nil {
        query = query.Where("scheduled_start >= ?", *filters.StartDate)
    }
    if filters.EndDate != nil {
        query = query.Where("scheduled_end <= ?", *filters.EndDate)
    }

    // Get total count
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // Apply pagination
    if filters.Page > 0 && filters.Limit > 0 {
        offset := (filters.Page - 1) * filters.Limit
        query = query.Offset(offset).Limit(filters.Limit)
    }

    // Apply sorting
    if filters.SortBy != "" {
        direction := "ASC"
        if filters.SortOrder == "desc" {
            direction = "DESC"
        }
        query = query.Order(filters.SortBy + " " + direction)
    } else {
        query = query.Order("scheduled_start ASC")
    }

    // Execute query with preloaded relationships
    err := query.
        Preload("Supplier").Preload("Supplier.User").
        Preload("Employee").Preload("Employee.User").
        Preload("Operation").
        Preload("Product").
        Find(&appointments).Error

    return appointments, total, err
}

// FindBySupplier finds appointments for a specific supplier
func (r *appointmentRepository) FindBySupplier(supplierID uint, filters AppointmentFilters) ([]models.Appointment, int64, error) {
    var appointments []models.Appointment
    var total int64

    query := r.db.Model(&models.Appointment{}).Where("supplier_id = ?", supplierID)

    // Apply filters
    if filters.Status != nil {
        query = query.Where("status = ?", *filters.Status)
    }
    if filters.StartDate != nil {
        query = query.Where("scheduled_start >= ?", *filters.StartDate)
    }
    if filters.EndDate != nil {
        query = query.Where("scheduled_end <= ?", *filters.EndDate)
    }

    // Get total count
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // Apply pagination
    if filters.Page > 0 && filters.Limit > 0 {
        offset := (filters.Page - 1) * filters.Limit
        query = query.Offset(offset).Limit(filters.Limit)
    }

    // Apply sorting
    if filters.SortBy != "" {
        direction := "ASC"
        if filters.SortOrder == "desc" {
            direction = "DESC"
        }
        query = query.Order(filters.SortBy + " " + direction)
    } else {
        query = query.Order("scheduled_start ASC")
    }

    // Execute query with preloaded relationships
    err := query.
        Preload("Supplier").Preload("Supplier.User").
        Preload("Employee").Preload("Employee.User").
        Preload("Operation").
        Preload("Product").
        Find(&appointments).Error

    return appointments, total, err
}

// FindByEmployee finds appointments for a specific employee
func (r *appointmentRepository) FindByEmployee(employeeID uint, filters AppointmentFilters) ([]models.Appointment, int64, error) {
    var appointments []models.Appointment
    var total int64

    query := r.db.Model(&models.Appointment{}).Where("employee_id = ?", employeeID)

    // Apply filters
    if filters.Status != nil {
        query = query.Where("status = ?", *filters.Status)
    }
    if filters.StartDate != nil {
        query = query.Where("scheduled_start >= ?", *filters.StartDate)
    }
    if filters.EndDate != nil {
        query = query.Where("scheduled_end <= ?", *filters.EndDate)
    }

    // Get total count
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // Apply pagination
    if filters.Page > 0 && filters.Limit > 0 {
        offset := (filters.Page - 1) * filters.Limit
        query = query.Offset(offset).Limit(filters.Limit)
    }

    // Apply sorting
    if filters.SortBy != "" {
        direction := "ASC"
        if filters.SortOrder == "desc" {
            direction = "DESC"
        }
        query = query.Order(filters.SortBy + " " + direction)
    } else {
        query = query.Order("scheduled_start ASC")
    }

    // Execute query with preloaded relationships
    err := query.
        Preload("Supplier").Preload("Supplier.User").
        Preload("Employee").Preload("Employee.User").
        Preload("Operation").
        Preload("Product").
        Find(&appointments).Error

    return appointments, total, err
}

// FindByOperation finds appointments for a specific operation
func (r *appointmentRepository) FindByOperation(operationID uint, filters AppointmentFilters) ([]models.Appointment, int64, error) {
    var appointments []models.Appointment
    var total int64

    query := r.db.Model(&models.Appointment{}).Where("operation_id = ?", operationID)

    // Apply filters
    if filters.Status != nil {
        query = query.Where("status = ?", *filters.Status)
    }
    if filters.StartDate != nil {
        query = query.Where("scheduled_start >= ?", *filters.StartDate)
    }
    if filters.EndDate != nil {
        query = query.Where("scheduled_end <= ?", *filters.EndDate)
    }

    // Get total count
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // Apply pagination
    if filters.Page > 0 && filters.Limit > 0 {
        offset := (filters.Page - 1) * filters.Limit
        query = query.Offset(offset).Limit(filters.Limit)
    }

    // Apply sorting
    if filters.SortBy != "" {
        direction := "ASC"
        if filters.SortOrder == "desc" {
            direction = "DESC"
        }
        query = query.Order(filters.SortBy + " " + direction)
    } else {
        query = query.Order("scheduled_start ASC")
    }

    // Execute query with preloaded relationships
    err := query.
        Preload("Supplier").Preload("Supplier.User").
        Preload("Employee").Preload("Employee.User").
        Preload("Operation").
        Preload("Product").
        Find(&appointments).Error

    return appointments, total, err
}

// FindByDateRange finds appointments within a specific date range
func (r *appointmentRepository) FindByDateRange(start, end time.Time, filters AppointmentFilters) ([]models.Appointment, int64, error) {
    var appointments []models.Appointment
    var total int64

    query := r.db.Model(&models.Appointment{}).
        Where("scheduled_start >= ? AND scheduled_end <= ?", start, end)

    // Apply filters
    if filters.Status != nil {
        query = query.Where("status = ?", *filters.Status)
    }

    // Get total count
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // Apply pagination
    if filters.Page > 0 && filters.Limit > 0 {
        offset := (filters.Page - 1) * filters.Limit
        query = query.Offset(offset).Limit(filters.Limit)
    }

    // Apply sorting
    if filters.SortBy != "" {
        direction := "ASC"
        if filters.SortOrder == "desc" {
            direction = "DESC"
        }
        query = query.Order(filters.SortBy + " " + direction)
    } else {
        query = query.Order("scheduled_start ASC")
    }

    // Execute query with preloaded relationships
    err := query.
        Preload("Supplier").Preload("Supplier.User").
        Preload("Employee").Preload("Employee.User").
        Preload("Operation").
        Preload("Product").
        Find(&appointments).Error

    return appointments, total, err
}

// FindUpcoming finds upcoming appointments
func (r *appointmentRepository) FindUpcoming(limit int) ([]models.Appointment, error) {
    var appointments []models.Appointment

    query := r.db.Model(&models.Appointment{}).
        Where("scheduled_start > ? AND status != ?", time.Now(), models.StatusCancelled).
        Order("scheduled_start ASC")

    if limit > 0 {
        query = query.Limit(limit)
    }

    err := query.
        Preload("Supplier").Preload("Supplier.User").
        Preload("Employee").Preload("Employee.User").
        Preload("Operation").
        Preload("Product").
        Find(&appointments).Error

    return appointments, err
}

// HasConflict checks if an appointment conflicts with existing appointments
func (r *appointmentRepository) HasConflict(appointment *models.Appointment) (bool, error) {
    var count int64

    // Check for conflicts with employee's schedule
    query := r.db.Model(&models.Appointment{}).
        Where("employee_id = ? AND id != ?", appointment.EmployeeID, appointment.ID).
        Where("status NOT IN ?", []models.AppointmentStatus{models.StatusCancelled}).
        Where("(scheduled_start < ? AND scheduled_end > ?) OR "+
            "(scheduled_start >= ? AND scheduled_start < ?) OR "+
            "(scheduled_end > ? AND scheduled_end <= ?)",
            appointment.ScheduledEnd, appointment.ScheduledStart,
            appointment.ScheduledStart, appointment.ScheduledEnd,
            appointment.ScheduledStart, appointment.ScheduledEnd)

    if err := query.Count(&count).Error; err != nil {
        return false, err
    }

    if count > 0 {
        return true, nil
    }

    // Check for conflicts with supplier's schedule
    query = r.db.Model(&models.Appointment{}).
        Where("supplier_id = ? AND id != ?", appointment.SupplierID, appointment.ID).
        Where("status NOT IN ?", []models.AppointmentStatus{models.StatusCancelled}).
        Where("(scheduled_start < ? AND scheduled_end > ?) OR "+
            "(scheduled_start >= ? AND scheduled_start < ?) OR "+
            "(scheduled_end > ? AND scheduled_end <= ?)",
            appointment.ScheduledEnd, appointment.ScheduledStart,
            appointment.ScheduledStart, appointment.ScheduledEnd,
            appointment.ScheduledStart, appointment.ScheduledEnd)

    if err := query.Count(&count).Error; err != nil {
        return false, err
    }

    return count > 0, nil
}

// UpdateStatus updates the status of an appointment
func (r *appointmentRepository) UpdateStatus(id uint, status models.AppointmentStatus, reason string) error {
    var appointment models.Appointment
    if err := r.db.First(&appointment, id).

package repository

import (
	"errors"
	"time"

	"github.com/bernardofernandezz/scheduling-api/internal/models"
	"gorm.io/gorm"
)

// AppointmentRepository interface defines methods for appointment repository
type AppointmentRepository interface {
	Create(appointment *models.Appointment) error
	FindByID(id uint) (*models.Appointment, error)
	Update(appointment *models.Appointment) error
	Delete(id uint) error
	List(filters AppointmentFilters) ([]models.Appointment, int64, error)
	UpdateStatus(id uint, status models.AppointmentStatus, reason string) error
	HasConflict(appointment *models.Appointment) (bool, error)
	FindBySupplier(supplierID uint, filters AppointmentFilters) ([]models.Appointment, int64, error)
	FindByEmployee(employeeID uint, filters AppointmentFilters) ([]models.Appointment, int64, error)
	FindByOperation(operationID uint, filters AppointmentFilters) ([]models.Appointment, int64, error)
	FindByDateRange(start, end time.Time, filters AppointmentFilters) ([]models.Appointment, int64, error)
	FindUpcoming(limit int) ([]models.Appointment, error)
	GetStatistics() (*AppointmentStatistics, error)
}

// AppointmentFilters defines filters for appointment queries
type AppointmentFilters struct {
	Status    *models.AppointmentStatus
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	Limit     int
	SortBy    string
	SortOrder string
}

// AppointmentStatistics represents appointment statistics
type AppointmentStatistics struct {
	TotalAppointments   int64
	PendingAppointments int64
	ConfirmedAppointments int64
	CancelledAppointments int64
	CompletedAppointments int64
	RescheduledAppointments int64
	AppointmentsByDay    map[string]int64
	AppointmentsByMonth  map[string]int64
}

// appointmentRepository implements AppointmentRepository interface
type appointmentRepository struct {
	db *gorm.DB
}

// NewAppointmentRepository creates a new appointment repository
func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
	return &appointmentRepository{db: db}
}

// Create creates a new appointment with conflict checking
func (r *appointmentRepository) Create(appointment *models.Appointment) error {
	// Validate appointment
	if err := appointment.Validate(); err != nil {
		return err
	}

	// Check for conflicts
	hasConflict, err := r.HasConflict(appointment)
	if err != nil {
		return err
	}
	if hasConflict {
		return errors.New("appointment conflicts with an existing appointment")
	}

	// Create appointment in a transaction
	err = r.db.Transaction(func(tx *gorm.DB) error {
		// Create appointment
		if err := tx.Create(appointment).Error; err != nil {
			return err
		}
		return nil
	})

	return err
}

// FindByID finds an appointment by ID
func (r *appointmentRepository) FindByID(id uint) (*models.Appointment, error) {
	var appointment models.Appointment
	err := r.db.Preload("Supplier").Preload("Supplier.User").
		Preload("Employee").Preload("Employee.User").
		Preload("Operation").Preload("Product").
		First(&appointment, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("appointment not found")
		}
		return nil, err
	}
	return &appointment, nil
}

// Update updates an appointment
func (r *appointmentRepository) Update(appointment *models.Appointment) error {
	// Validate appointment
	if err := appointment.Validate(); err != nil {
		return err
	}

	// Check for conflicts (only if dates are changing)
	existingAppointment, err := r.FindByID(appointment.ID)
	if err != nil {
		return err
	}

	// If start or end time has changed, check for conflicts
	if !existingAppointment.ScheduledStart.Equal(appointment.ScheduledStart) ||
		!existingAppointment.ScheduledEnd.Equal(appointment.ScheduledEnd) {
		hasConflict, err := r.HasConflict(appointment)
		if err != nil {
			return err
		}
		if hasConflict {
			return errors.New("updated appointment conflicts with an existing appointment")
		}
	}

	// Update appointment
	return r.db.Save(appointment).Error
}

// Delete soft deletes an appointment
func (r *appointmentRepository) Delete(id uint) error {
	return r.db.Delete(&models.Appointment{}, id).Error
}

// UpdateStatus updates an appointment's status
func (r *appointmentRepository) UpdateStatus(id uint, status models.AppointmentStatus, reason string) error {
	appointment, err := r.FindByID(id)
	if err != nil {
		return err
	}

	// Update status and related fields
	appointment.Status = status
	now := time.Now()

	switch status {
	case models.StatusConfirmed:
		appointment.ConfirmedAt = &now
	case models.StatusCancelled:
		appointment.CancelledAt = &now
		appointment.CancellationReason = reason
	case models.StatusCompleted:
		appointment.CompletedAt = &now
	}

	return r.db.Save(appointment).Error
}

// HasConflict checks if an appointment conflicts with existing appointments
func (r *appointmentRepository) HasConflict(appointment *models.Appointment) (bool, error) {
	var count int64

	// Check for employee conflicts
	query := r.db.Model(&models.Appointment{}).
		Where("employee_id = ? AND id != ?", appointment.EmployeeID, appointment.ID).
		Where("status NOT IN ?", []models.AppointmentStatus{models.StatusCancelled}).
		Where(
			"(scheduled_start < ? AND scheduled_end > ?) OR "+
				"(scheduled_start >= ? AND scheduled_start < ?) OR "+
				"(scheduled_end > ? AND scheduled_end <= ?)",
			appointment.ScheduledEnd, appointment.ScheduledStart,
			appointment.ScheduledStart, appointment.ScheduledEnd,
			appointment.ScheduledStart, appointment.ScheduledEnd,
		)

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	// Check for supplier conflicts
	query = r.db.Model(&models.Appointment{}).
		Where("supplier_id = ? AND id != ?", appointment.SupplierID, appointment.ID).
		Where("status NOT IN ?", []models.AppointmentStatus{models.StatusCancelled}).
		Where(
			"(scheduled_start < ? AND scheduled_end > ?) OR "+
				"(scheduled_start >= ? AND scheduled_start < ?) OR "+
				"(scheduled_end > ? AND scheduled_end <= ?)",
			appointment.ScheduledEnd, appointment.ScheduledStart,
			appointment.ScheduledStart, appointment.ScheduledEnd,
			appointment.ScheduledStart, appointment.ScheduledEnd,
		)

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// List returns a paginated list of appointments with filters
func (r *appointmentRepository) List(filters AppointmentFilters) ([]models.Appointment, int64, error) {
	var appointments []models.Appointment
	var count int64

	// Base query
	query := r.db.Model(&models.Appointment{})

	// Apply filters
	query = applyAppointmentFilters(query, filters)

	// Count total records
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	if filters.Page > 0 && filters.Limit > 0 {
		offset := (filters.Page - 1) * filters.Limit
		query = query.Offset(offset).Limit(filters.Limit)
	}

	// Apply sorting
	if filters.SortBy != "" {
		sortOrder := "ASC"
		if filters.SortOrder == "desc" {
			sortOrder = "DESC"
		}
		query = query.Order(filters.SortBy + " " + sortOrder)
	} else {
		// Default sorting by scheduled start time
		query = query.Order("scheduled_start ASC")
	}

	// Fetch appointments with preloaded relations
	err := query.Preload("Supplier").Preload("Supplier.User").
		Preload("Employee").Preload("Employee.User").
		Preload("Operation").Preload("Product").
		Find(&appointments).Error

	if err != nil {
		return nil, 0, err
	}

	return appointments, count, nil
}

// FindBySupplier finds appointments by supplier
func (r *appointmentRepository) FindBySupplier(supplierID uint, filters AppointmentFilters) ([]models.Appointment, int64, error) {
	var appointments []models.Appointment
	var count int64

	// Base query with supplier filter
	query := r.db.Model(&models.Appointment{}).Where("supplier_id = ?", supplierID)

	// Apply other filters
	query = applyAppointmentFilters(query, filters)

	// Count total records
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	if filters.Page > 0 && filters.Limit > 0 {
		offset := (filters.Page - 1) * filters.Limit
		query = query.Offset(offset).Limit(filters.Limit)
	}

	// Apply sorting
	if filters.SortBy != "" {
		sortOrder := "ASC"
		if filters.SortOrder == "desc" {
			sortOrder = "DESC"
		}
		query = query.Order(filters.SortBy + " " + sortOrder)
	} else {
		// Default sorting by scheduled start time
		query = query.Order("scheduled_start ASC")
	}

	// Fetch appointments with preloaded relations
	err := query.Preload("Supplier").Preload("Supplier.User").
		Preload("Employee").Preload("Employee.User").
		Preload("Operation").Preload("Product").
		Find(&appointments).Error

	if err != nil {
		return nil, 0, err
	}

	return appointments, count, nil
}

// FindByEmployee finds appointments by employee
func (r *appointmentRepository) FindByEmployee(employeeID uint, filters AppointmentFilters) ([]models.Appointment, int64, error) {
	var appointments []models.Appointment
	var count int64

	// Base query with employee filter
	query := r.db.Model(&models.Appointment{}).Where("employee_id = ?", employeeID)

	// Apply other filters
	query = applyAppointmentFilters(query, filters)

	// Count total records
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	if filters.Page > 0 && filters.Limit > 0 {
		offset := (filters.Page - 1) * filters.Limit
		query = query.Offset(offset).Limit(filters.Limit)
	}

	// Apply sorting
	if filters.SortBy != "" {
		sortOrder := "ASC"
		if filters.SortOrder == "desc" {
			sortOrder = "DESC"
		}
		query = query.Order(filters.SortBy + " " + sortOrder)
	} else {
		// Default sorting by scheduled start time
		query = query.Order("scheduled_start ASC")
	}

	// Fetch appointments with preloaded relations
	err := query.Preload("Supplier").Preload("Supplier.User").
		Preload("Employee").Preload("Employee.User").
		Preload("Operation").Preload("Product").
		Find(&appointments).Error

	if err != nil {
		return nil, 0, err
	}

	return appointments, count, nil
}

// FindByOperation finds appointments by operation
func (r *appointmentRepository) FindByOperation(operationID uint, filters AppointmentFilters) ([]models.Appointment, int64, error) {
	var appointments []models.Appointment
	var count int64

	// Base query with operation filter
	query := r.db.Model(&models.Appointment{}).Where("operation_id = ?", operationID)

	// Apply other filters
	query = applyAppointmentFilters(query, filters)

	// Count total records
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	if filters.Page > 0 && filters.Limit > 0 {
		offset := (filters.Page - 1) * filters.Limit
		query = query.Offset(offset).Limit(filters.Limit)
	}

	// Apply sorting
	if filters.SortBy != "" {
		sortOrder := "ASC"
		if filters.SortOrder == "desc" {
			sortOrder = "DESC"
		}
		query = query.Order(filters.SortBy + " " + sortOrder)
	} else {
		// Default sorting by scheduled start time
		query = query.Order("scheduled_start ASC")
	}

	// Fetch appointments with preloaded relations
	err := query.Preload("Supplier").Preload("Supplier.User").
		Preload("Employee").Preload("Employee.User").
		Preload("Operation").Preload("Product").
		Find(&appointments).Error

	if err != nil {
		return nil, 0, err
	}

	return appointments, count, nil
}

// FindByDateRange finds appointments within a date range
func (r *appointmentRepository) FindByDateRange(start, end time.Time, filters AppointmentFilters) ([]models.Appointment, int64, error) {
	var appointments []models.Appointment
	var count int64

	// Base query with date range filter
	query := r.db.Model(&models.Appointment{}).
		Where("scheduled_start >= ? AND scheduled_start <= ?", start, end)

	// Apply other filters
	query = applyAppointmentFilters(query, filters)

	// Count total records
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	if filters.Page > 0 && filters.Limit > 0 {
		offset := (filters.Page - 1) * filters.Limit
		query = query.Offset(offset).Limit(filters.

