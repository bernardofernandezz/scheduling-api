package models

import (
    "time"
    "errors"
)

// AvailabilitySlot represents a time slot when an employee is available for appointments
type AvailabilitySlot struct {
    ID           uint      `json:"id" gorm:"primaryKey"`
    EmployeeID   uint      `json:"employee_id" gorm:"not null;index"`
    Employee     Employee  `json:"employee" gorm:"foreignKey:EmployeeID"`
    OperationID  uint      `json:"operation_id" gorm:"not null;index"`
    Operation    Operation `json:"operation" gorm:"foreignKey:OperationID"`
    DayOfWeek    int       `json:"day_of_week" gorm:"not null"` // 0=Sunday, 1=Monday, etc.
    StartTime    string    `json:"start_time" gorm:"not null"`  // Format: "HH:MM"
    EndTime      string    `json:"end_time" gorm:"not null"`    // Format: "HH:MM"
    IsRecurring  bool      `json:"is_recurring" gorm:"default:true"`
    SpecificDate *time.Time `json:"specific_date"`  // Used for non-recurring slots
    Active       bool      `json:"active" gorm:"default:true"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

// Validate performs validation on the availability slot
func (a *AvailabilitySlot) Validate() error {
    if a.EmployeeID == 0 {
        return errors.New("employee is required")
    }
    if a.OperationID == 0 {
        return errors.New("operation is required")
    }
    if a.DayOfWeek < 0 || a.DayOfWeek > 6 {
        return errors.New("day of week must be between 0 and 6")
    }
    if a.StartTime == "" {
        return errors.New("start time is required")
    }
    if a.EndTime == "" {
        return errors.New("end time is required")
    }

    // For non-recurring slots, a specific date is required
    if !a.IsRecurring && a.SpecificDate == nil {
        return errors.New("specific date is required for non-recurring slots")
    }
    
    return nil
}

// BeforeCreate is called by GORM before creating a new record
func (a *AvailabilitySlot) BeforeCreate() error {
    return a.Validate()
}

// BeforeUpdate is called by GORM before updating a record
func (a *AvailabilitySlot) BeforeUpdate() error {
    return a.Validate()
}

// OverlapsWith checks if this availability slot overlaps with another
func (a *AvailabilitySlot) OverlapsWith(other *AvailabilitySlot) bool {
    // Different days don't overlap
    if a.DayOfWeek != other.DayOfWeek {
        return false
    }
    
    // For non-recurring slots, check specific dates
    if !a.IsRecurring && !other.IsRecurring {
        if a.SpecificDate != nil && other.SpecificDate != nil {
            // Different dates don't overlap
            if !a.SpecificDate.Equal(*other.SpecificDate) {
                return false
            }
        }
    }
    
    // Check time overlap
    return !(a.EndTime <= other.StartTime || a.StartTime >= other.EndTime)
}

