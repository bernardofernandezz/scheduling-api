package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// User represents a system user (can be an employee or a supplier)
type User struct {
	BaseModel
	Name         string `json:"name"`
	Email        string `gorm:"uniqueIndex" json:"email"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"` // "admin", "employee", "supplier"
	Phone        string `json:"phone"`
	Active       bool   `gorm:"default:true" json:"active"`
}

// Supplier represents a supplier entity
type Supplier struct {
	BaseModel
	UserID      uint   `json:"user_id"`
	User        User   `json:"user"`
	CompanyName string `json:"company_name"`
	CNPJ        string `gorm:"uniqueIndex" json:"cnpj"`
	Address     string `json:"address"`
	Category    string `json:"category"`
}

// Employee represents an employee of the company
type Employee struct {
	BaseModel
	UserID         uint   `json:"user_id"`
	User           User   `json:"user"`
	Department     string `json:"department"`
	Position       string `json:"position"`
	EmployeeNumber string `json:"employee_number"`
	Operation      string `json:"operation"` // Which operation (branch/location) the employee belongs to
}

// Product represents a product that can be delivered
type Product struct {
	BaseModel
	Name        string  `json:"name"`
	Description string  `json:"description"`
	SKU         string  `gorm:"uniqueIndex" json:"sku"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	SupplierID  uint    `json:"supplier_id"`
	Supplier    Supplier `json:"supplier"`
	Active      bool    `gorm:"default:true" json:"active"`
}

// Operation represents a company location/branch
type Operation struct {
	BaseModel
	Name         string     `json:"name"`
	Address      string     `json:"address"`
	City         string     `json:"city"`
	State        string     `json:"state"`
	ZipCode      string     `json:"zip_code"`
	ContactPhone string     `json:"contact_phone"`
	Active       bool       `gorm:"default:true" json:"active"`
	OpeningHour  time.Time  `json:"opening_hour"`
	ClosingHour  time.Time  `json:"closing_hour"`
	Employees    []Employee `gorm:"many2many:operation_employees" json:"employees"`
}

// AppointmentStatus represents the status of an appointment
type AppointmentStatus string

const (
	StatusPending   AppointmentStatus = "pending"
	StatusConfirmed AppointmentStatus = "confirmed"
	StatusCancelled AppointmentStatus = "cancelled"
	StatusCompleted AppointmentStatus = "completed"
	StatusRescheduled AppointmentStatus = "rescheduled"
)

// Appointment represents a scheduled appointment between a supplier and an employee
type Appointment struct {
	BaseModel
	SupplierID      uint             `json:"supplier_id"`
	Supplier        Supplier         `json:"supplier"`
	EmployeeID      uint             `json:"employee_id"`
	Employee        Employee         `json:"employee"`
	OperationID     uint             `json:"operation_id"`
	Operation       Operation        `json:"operation"`
	ProductID       uint             `json:"product_id"`
	Product         Product          `json:"product"`
	ScheduledStart  time.Time        `json:"scheduled_start"`
	ScheduledEnd    time.Time        `json:"scheduled_end"`
	Status          AppointmentStatus `gorm:"default:'pending'" json:"status"`
	Notes           string           `json:"notes"`
	QuantityToDeliver int            `json:"quantity_to_deliver"`
	ConfirmedAt     *time.Time       `json:"confirmed_at"`
	CancelledAt     *time.Time       `json:"cancelled_at"`
	CompletedAt     *time.Time       `json:"completed_at"`
	CancellationReason string        `json:"cancellation_reason"`
}

// Validate validates an appointment
func (a *Appointment) Validate() error {
	if a.SupplierID == 0 {
		return errors.New("supplier is required")
	}
	if a.EmployeeID == 0 {
		return errors.New("employee is required")
	}
	if a.OperationID == 0 {
		return errors.New("operation is required")
	}
	if a.ProductID == 0 {
		return errors.New("product is required")
	}
	if a.ScheduledStart.IsZero() {
		return errors.New("scheduled start time is required")
	}
	if a.ScheduledEnd.IsZero() {
		return errors.New("scheduled end time is required")
	}
	if a.ScheduledStart.After(a.ScheduledEnd) {
		return errors.New("scheduled start time must be before scheduled end time")
	}
	if a.QuantityToDeliver <= 0 {
		return errors.New("quantity to deliver must be greater than zero")
	}
	
	// Check if the appointment is at least 1 hour
	if a.ScheduledEnd.Sub(a.ScheduledStart) < time.Hour {
		return errors.New("appointment must be at least 1 hour long")
	}

	return nil
}

// AvailabilitySlot represents a time slot when an employee is available for appointments
type AvailabilitySlot struct {
	BaseModel
	EmployeeID  uint      `json:"employee_id"`
	Employee    Employee  `json:"employee"`
	OperationID uint      `json:"operation_id"`
	Operation   Operation `json:"operation"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	DayOfWeek   int       `json:"day_of_week"` // 0=Sunday, 1=Monday, etc.
	IsRecurring bool      `gorm:"default:true" json:"is_recurring"`
}

