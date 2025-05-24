package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// RecurrencePattern defines how often an appointment recurs
type RecurrencePattern string

const (
	// RecurrenceDaily represents a daily recurrence
	RecurrenceDaily RecurrencePattern = "daily"
	
	// RecurrenceWeekly represents a weekly recurrence
	RecurrenceWeekly RecurrencePattern = "weekly"
	
	// RecurrenceMonthly represents a monthly recurrence
	RecurrenceMonthly RecurrencePattern = "monthly"
	
	// RecurrenceBiweekly represents a biweekly (every 2 weeks) recurrence
	RecurrenceBiweekly RecurrencePattern = "biweekly"
)

// WeekDay represents a day of the week
type WeekDay int

const (
	Sunday WeekDay = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

// RecurringAppointment represents a template for generating multiple appointments
// with a specified recurrence pattern
type RecurringAppointment struct {
	gorm.Model
	
	// Basic information
	SupplierID        uint               `json:"supplier_id" gorm:"not null"`
	EmployeeID        uint               `json:"employee_id" gorm:"not null"`
	OperationID       uint               `json:"operation_id" gorm:"not null"`
	ProductID         uint               `json:"product_id" gorm:"not null"`
	QuantityToDeliver int                `json:"quantity_to_deliver" gorm:"not null"`
	Notes             string             `json:"notes"`
	
	// Recurrence details
	Pattern           RecurrencePattern  `json:"pattern" gorm:"not null"`
	StartDate         time.Time          `json:"start_date" gorm:"not null"`
	EndDate           *time.Time         `json:"end_date"`
	MaxOccurrences    *int               `json:"max_occurrences"`
	
	// Time of day for the appointment (stored in minutes from midnight)
	StartTimeMinutes  int                `json:"start_time_minutes" gorm:"not null"`
	DurationMinutes   int                `json:"duration_minutes" gorm:"not null"`
	
	// For weekly/biweekly patterns, which days of the week
	WeekDays          []WeekDay          `json:"week_days" gorm:"-"`
	WeekDaysString    string             `json:"-" gorm:"column:week_days"`
	
	// For monthly patterns, which day of the month (1-31)
	MonthDay          *int               `json:"month_day"`
	
	// For all patterns, specific dates to exclude
	ExclusionDates    []time.Time        `json:"exclusion_dates" gorm:"-"`
	ExclusionJSON     string             `json:"-" gorm:"column:exclusion_dates"`
	
	// Related entities
	Supplier          Supplier           `json:"supplier" gorm:"foreignKey:SupplierID"`
	Employee          Employee           `json:"employee" gorm:"foreignKey:EmployeeID"`
	Operation         Operation          `json:"operation" gorm:"foreignKey:OperationID"`
	Product           Product            `json:"product" gorm:"foreignKey:ProductID"`
	
	// Generated appointments (one-to-many relationship)
	Appointments      []Appointment      `json:"appointments" gorm:"foreignKey:RecurringAppointmentID"`
}

// Validate ensures the recurring appointment data is valid
func (ra *RecurringAppointment) Validate() error {
	// Check required fields
	if ra.SupplierID == 0 {
		return errors.New("supplier is required")
	}
	if ra.EmployeeID == 0 {
		return errors.New("employee is required")
	}
	if ra.OperationID == 0 {
		return errors.New("operation is required")
	}
	if ra.ProductID == 0 {
		return errors.New("product is required")
	}
	if ra.QuantityToDeliver <= 0 {
		return errors.New("quantity to deliver must be greater than zero")
	}
	
	// Validate recurrence pattern
	switch ra.Pattern {
	case RecurrenceDaily, RecurrenceWeekly, RecurrenceBiweekly, RecurrenceMonthly:
		// Valid pattern
	default:
		return errors.New("invalid recurrence pattern")
	}
	
	// Validate start date is not in the past
	if ra.StartDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return errors.New("start date cannot be in the past")
	}
	
	// Validate end date is after start date if specified
	if ra.EndDate != nil && ra.EndDate.Before(ra.StartDate) {
		return errors.New("end date must be after start date")
	}
	
	// Either end date or max occurrences must be specified
	if ra.EndDate == nil && ra.MaxOccurrences == nil {
		return errors.New("either end date or maximum occurrences must be specified")
	}
	
	// Validate max occurrences if specified
	if ra.MaxOccurrences != nil && *ra.MaxOccurrences <= 0 {
		return errors.New("maximum occurrences must be greater than zero")
	}
	
	// Validate appointment duration
	if ra.DurationMinutes < 30 {
		return errors.New("appointment duration must be at least 30 minutes")
	}
	if ra.DurationMinutes > 480 {
		return errors.New("appointment duration cannot exceed 8 hours")
	}
	
	// Validate start time is within a day
	if ra.StartTimeMinutes < 0 || ra.StartTimeMinutes > 1439 {
		return errors.New("start time must be between 0 and 1439 minutes")
	}
	
	// Validate pattern-specific fields
	switch ra.Pattern {
	case RecurrenceWeekly, RecurrenceBiweekly:
		if len(ra.WeekDays) == 0 {
			return errors.New("at least one weekday must be selected for weekly/biweekly patterns")
		}
	case RecurrenceMonthly:
		if ra.MonthDay == nil {
			return errors.New("day of month must be specified for monthly pattern")
		}
		if *ra.MonthDay < 1 || *ra.MonthDay > 31 {
			return errors.New("day of month must be between 1 and 31")
		}
	}
	
	return nil
}

// BeforeSave prepares the model for saving to the database
func (ra *RecurringAppointment) BeforeSave(tx *gorm.DB) error {
	// Convert weekdays slice to string representation for storage
	if len(ra.WeekDays) > 0 {
		weekDaysStr := ""
		for i, day := range ra.WeekDays {
			if i > 0 {
				weekDaysStr += ","
			}
			weekDaysStr += string(rune('0' + int(day)))
		}
		ra.WeekDaysString = weekDaysStr
	}
	
	// Convert exclusion dates to JSON string
	if len(ra.ExclusionDates) > 0 {
		exclusionDatesStr := "["
		for i, date := range ra.ExclusionDates {
			if i > 0 {
				exclusionDatesStr += ","
			}
			exclusionDatesStr += "\"" + date.Format(time.RFC3339) + "\""
		}
		exclusionDatesStr += "]"
		ra.ExclusionJSON = exclusionDatesStr
	}
	
	return ra.Validate()
}

// AfterFind converts database representation back to usable fields
func (ra *RecurringAppointment) AfterFind(tx *gorm.DB) error {
	// Convert string representation of weekdays back to slice
	if ra.WeekDaysString != "" {
		for _, c := range ra.WeekDaysString {
			if c != ',' {
				day := WeekDay(int(c - '0'))
				ra.WeekDays = append(ra.WeekDays, day)
			}
		}
	}
	
	// Parse exclusion dates from JSON
	if ra.ExclusionJSON != "" {
		// For simplicity, we're doing basic string parsing here
		// In a real implementation, use proper JSON unmarshaling
		dateStr := ""
		inQuotes := false
		ra.ExclusionDates = []time.Time{}
		
		for _, c := range ra.ExclusionJSON {
			if c == '"' {
				if inQuotes {
					// End of a date string
					if date, err := time.Parse(time.RFC3339, dateStr); err == nil {
						ra.ExclusionDates = append(ra.ExclusionDates, date)
					}
					dateStr = ""
				} else {
					// Start of a date string
					dateStr = ""
				}
				inQuotes = !inQuotes
			} else if inQuotes {
				dateStr += string(c)
			}
		}
	}
	
	return nil
}

// GenerateOccurrences generates all appointment occurrences based on the recurrence pattern
func (ra *RecurringAppointment) GenerateOccurrences() []time.Time {
	var occurrences []time.Time
	
	// Determine the end date for generating occurrences
	endDate := time.Now().AddDate(10, 0, 0) // Default to 10 years in the future
	if ra.EndDate != nil {
		endDate = *ra.EndDate
	}
	
	// Calculate the start and end times for appointments
	startMinutesFromMidnight := ra.StartTimeMinutes
	startHour := startMinutesFromMidnight / 60
	startMinute := startMinutesFromMidnight % 60
	
	// Begin with the start date
	currentDate := ra.StartDate
	
	// Set the correct time of day
	currentDate = time.Date(
		currentDate.Year(),
		currentDate.Month(),
		currentDate.Day(),
		startHour,
		startMinute,
		0, 0,
		currentDate.Location(),
	)
	
	// Generate occurrences based on the pattern
	occurrenceCount := 0
	maxOccurrences := 1000 // Default limit
	if ra.MaxOccurrences != nil {
		maxOccurrences = *ra.MaxOccurrences
	}
	
	for currentDate.Before(endDate) && occurrenceCount < maxOccurrences {
		// Check if this date should be included based on the pattern
		includeDate := false
		
		switch ra.Pattern {
		case RecurrenceDaily:
			includeDate = true
			
		case RecurrenceWeekly:
			weekday := WeekDay(int(currentDate.Weekday()))
			for _, day := range ra.WeekDays {
				if day == weekday {
					includeDate = true
					break
				}
			}
			
		case RecurrenceBiweekly:
			// Calculate week number since start date
			weeksSinceStart := int(currentDate.Sub(ra.StartDate).Hours() / 24 / 7)
			if weeksSinceStart%2 == 0 { // Only include every other week
				weekday := WeekDay(int(currentDate.Weekday()))
				for _, day := range ra.WeekDays {
					if day == weekday {
						includeDate = true
						break
					}
				}
			}
			
		case RecurrenceMonthly:
			if ra.MonthDay != nil && currentDate.Day() == *ra.MonthDay {
				includeDate = true
			}
		}
		
		// Check if this date is in the exclusion list
		for _, exclusionDate := range ra.ExclusionDates {
			if currentDate.Year() == exclusionDate.Year() &&
				currentDate.Month() == exclusionDate.Month() &&
				currentDate.Day() == exclusionDate.Day() {
				includeDate = false
				break
			}
		}
		
		// Add to occurrences if included
		if includeDate {
			occurrences = append(occurrences, currentDate)
			occurrenceCount++
		}
		
		// Advance to the next day
		currentDate = currentDate.AddDate(0, 0, 1)
	}
	
	return occurrences
}

// GenerateAppointments creates actual Appointment models from the recurring template
func (ra *RecurringAppointment) GenerateAppointments() []Appointment {
	occurrences := ra.GenerateOccurrences()
	appointments := make([]Appointment, 0, len(occurrences))
	
	for _, occurrence := range occurrences {
		// Calculate start and end times
		startTime := occurrence
		endTime := occurrence.Add(time.Duration(ra.DurationMinutes) * time.Minute)
		
		appointment := Appointment{
			SupplierID:            ra.SupplierID,
			EmployeeID:            ra.EmployeeID,
			OperationID:           ra.OperationID,
			ProductID:             ra.ProductID,
			ScheduledStart:        startTime,
			ScheduledEnd:          endTime,
			Notes:                 ra.Notes,
			QuantityToDeliver:     ra.QuantityToDeliver,
			Status:                StatusPending,
			RecurringAppointmentID: &ra.ID,
		}
		
		appointments = append(appointments, appointment)
	}
	
	return appointments
}

