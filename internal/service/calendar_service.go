package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bernardofernandezz/scheduling-api/internal/config"
	"github.com/bernardofernandezz/scheduling-api/internal/models"
	"github.com/bernardofernandezz/scheduling-api/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// CalendarProvider defines supported calendar providers
type CalendarProvider string

const (
	// GoogleCalendar represents Google Calendar provider
	GoogleCalendar CalendarProvider = "google"
	
	// OutlookCalendar represents Microsoft Outlook Calendar provider
	OutlookCalendar CalendarProvider = "outlook"
	
	// ICalFormat represents standard iCalendar format
	ICalFormat CalendarProvider = "ical"
)

// CalendarService defines the interface for calendar operations
type CalendarService interface {
	// iCalendar operations
	GenerateICalForAppointment(appointment *models.Appointment) (string, error)
	GenerateICalForRecurringAppointment(recurringAppointment *models.RecurringAppointment) (string, error)
	
	// Google Calendar operations
	CreateGoogleCalendarEvent(ctx context.Context, appointment *models.Appointment, calendarID string, accessToken string) (string, error)
	UpdateGoogleCalendarEvent(ctx context.Context, appointment *models.Appointment, eventID string, calendarID string, accessToken string) error
	DeleteGoogleCalendarEvent(ctx context.Context, eventID string, calendarID string, accessToken string) error
	
	// Calendar sync operations
	SyncAppointmentToCalendar(ctx context.Context, appointment *models.Appointment, userID uint, provider CalendarProvider) (string, error)
	RemoveAppointmentFromCalendar(ctx context.Context, appointment *models.Appointment, userID uint, provider CalendarProvider) error
	
	// Calendar link generation
	GenerateGoogleCalendarLink(appointment *models.Appointment) string
	GenerateOutlookCalendarLink(appointment *models.Appointment) string
	GenerateICalDownloadLink(appointment *models.Appointment) string
	
	// User calendar preferences
	GetUserCalendarPreferences(userID uint) (map[string]interface{}, error)
	UpdateUserCalendarPreferences(userID uint, preferences map[string]interface{}) error
}

// calendarService implements the CalendarService interface
type calendarService struct {
	appointmentRepo   repository.AppointmentRepository
	employeeRepo      repository.EmployeeRepository
	supplierRepo      repository.SupplierRepository
	userRepo          repository.UserRepository
	calendarSyncRepo  repository.CalendarSyncRepository
	config            *config.Config
	baseURL           string
}

// NewCalendarService creates a new calendar service
func NewCalendarService(
	appointmentRepo repository.AppointmentRepository,
	employeeRepo repository.EmployeeRepository,
	supplierRepo repository.SupplierRepository,
	userRepo repository.UserRepository,
	calendarSyncRepo repository.CalendarSyncRepository,
	config *config.Config,
) CalendarService {
	baseURL := "https://scheduling-api.example.com"
	if config != nil && config.Server != nil && config.Server.BaseURL != "" {
		baseURL = config.Server.BaseURL
	}
	
	return &calendarService{
		appointmentRepo:   appointmentRepo,
		employeeRepo:      employeeRepo,
		supplierRepo:      supplierRepo,
		userRepo:          userRepo,
		calendarSyncRepo:  calendarSyncRepo,
		config:            config,
		baseURL:           baseURL,
	}
}

// GenerateICalForAppointment generates an iCalendar (RFC 5545) format string for an appointment
func (s *calendarService) GenerateICalForAppointment(appointment *models.Appointment) (string, error) {
	// Retrieve related entities for more detailed calendar entry
	var supplierName, employeeName, operationName, productName string
	
	// Get supplier name
	supplier, err := s.supplierRepo.GetByID(appointment.SupplierID)
	if err == nil && supplier != nil {
		supplierName = supplier.Name
	}
	
	// Get employee name
	employee, err := s.employeeRepo.GetByID(appointment.EmployeeID)
	if err == nil && employee != nil {
		employeeName = employee.Name
	}
	
	// Generate a unique identifier for the event
	uid := fmt.Sprintf("appointment-%d@%s", appointment.ID, strings.Replace(s.baseURL, "https://", "", 1))
	
	// Format the start and end times in iCalendar format (UTC)
	startTime := appointment.ScheduledStart.UTC().Format("20060102T150405Z")
	endTime := appointment.ScheduledEnd.UTC().Format("20060102T150405Z")
	
	// Get current time for DTSTAMP
	now := time.Now().UTC().Format("20060102T150405Z")
	
	// Create appointment summary
	summary := fmt.Sprintf("Delivery: %s", productName)
	if supplierName != "" {
		summary = fmt.Sprintf("Delivery from %s", supplierName)
	}
	
	// Create appointment description
	description := fmt.Sprintf("Supplier: %s\nEmployee: %s\nOperation: %s\nProduct: %s\nQuantity: %d\nStatus: %s",
		supplierName, employeeName, operationName, productName, appointment.QuantityToDeliver, appointment.Status)
	
	if appointment.Notes != "" {
		description += fmt.Sprintf("\n\nNotes: %s", appointment.Notes)
	}
	
	// Build the iCalendar content
	var buffer bytes.Buffer
	buffer.WriteString("BEGIN:VCALENDAR\r\n")
	buffer.WriteString("VERSION:2.0\r\n")
	buffer.WriteString("PRODID:-//Scheduling API//Delivery Appointment//EN\r\n")
	buffer.WriteString("CALSCALE:GREGORIAN\r\n")
	buffer.WriteString("METHOD:PUBLISH\r\n")
	buffer.WriteString("BEGIN:VEVENT\r\n")
	buffer.WriteString(fmt.Sprintf("UID:%s\r\n", uid))
	buffer.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", now))
	buffer.WriteString(fmt.Sprintf("DTSTART:%s\r\n", startTime))
	buffer.WriteString(fmt.Sprintf("DTEND:%s\r\n", endTime))
	buffer.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", summary))
	
	// Add description with line folding (lines must be <= 75 chars)
	descLines := foldLines(fmt.Sprintf("DESCRIPTION:%s", description))
	buffer.WriteString(descLines)
	
	// Add location if available
	if operationName != "" {
		buffer.WriteString(fmt.Sprintf("LOCATION:%s\r\n", operationName))
	}
	
	// Add status
	var status string
	switch appointment.Status {
	case models.StatusConfirmed:
		status = "CONFIRMED"
	case models.StatusCancelled:
		status = "CANCELLED"
	default:
		status = "TENTATIVE"
	}
	buffer.WriteString(fmt.Sprintf("STATUS:%s\r\n", status))
	
	// Add URL to the appointment in the system
	buffer.WriteString(fmt.Sprintf("URL:%s/appointments/%d\r\n", s.baseURL, appointment.ID))
	
	// Add organizer if we have employee information
	if employeeName != "" {
		buffer.WriteString(fmt.Sprintf("ORGANIZER;CN=%s:mailto:noreply@example.com\r\n", employeeName))
	}
	
	buffer.WriteString("END:VEVENT\r\n")
	buffer.WriteString("END:VCALENDAR\r\n")
	
	return buffer.String(), nil
}

// GenerateICalForRecurringAppointment generates an iCalendar format string for a recurring appointment
func (s *calendarService) GenerateICalForRecurringAppointment(recurringAppointment *models.RecurringAppointment) (string, error) {
	// Retrieve related entities for more detailed calendar entry
	var supplierName, employeeName, operationName, productName string
	
	// Get supplier name
	supplier, err := s.supplierRepo.GetByID(recurringAppointment.SupplierID)
	if err == nil && supplier != nil {
		supplierName = supplier.Name
	}
	
	// Get employee name
	employee, err := s.employeeRepo.GetByID(recurringAppointment.EmployeeID)
	if err == nil && employee != nil {
		employeeName = employee.Name
	}
	
	// Generate a unique identifier for the event
	uid := fmt.Sprintf("recurring-%d@%s", recurringAppointment.ID, strings.Replace(s.baseURL, "https://", "", 1))
	
	// Get current time for DTSTAMP
	now := time.Now().UTC().Format("20060102T150405Z")
	
	// Calculate the start and end times for each occurrence
	startHour := recurringAppointment.StartTimeMinutes / 60
	startMinute := recurringAppointment.StartTimeMinutes % 60
	
	// Format the start date
	startDate := recurringAppointment.StartDate
	
	// Format the start time
	startTime := time.Date(
		startDate.Year(),
		startDate.Month(),
		startDate.Day(),
		startHour,
		startMinute,
		0, 0,
		time.UTC,
	).Format("20060102T150405Z")
	
	// Calculate end time by adding duration
	endTime := time.Date(
		startDate.Year(),
		startDate.Month(),
		startDate.Day(),
		startHour,
		startMinute,
		0, 0,
		time.UTC,
	).Add(time.Duration(recurringAppointment.DurationMinutes) * time.Minute).Format("20060102T150405Z")
	
	// Create appointment summary
	summary := fmt.Sprintf("Recurring Delivery: %s", productName)
	if supplierName != "" {
		summary = fmt.Sprintf("Recurring Delivery from %s", supplierName)
	}
	
	// Create appointment description
	description := fmt.Sprintf("Supplier: %s\nEmployee: %s\nOperation: %s\nProduct: %s\nQuantity: %d",
		supplierName, employeeName, operationName, productName, recurringAppointment.QuantityToDeliver)
	
	if recurringAppointment.Notes != "" {
		description += fmt.Sprintf("\n\nNotes: %s", recurringAppointment.Notes)
	}
	
	// Build the iCalendar content
	var buffer bytes.Buffer
	buffer.WriteString("BEGIN:VCALENDAR\r\n")
	buffer.WriteString("VERSION:2.0\r\n")
	buffer.WriteString("PRODID:-//Scheduling API//Recurring Delivery Appointment//EN\r\n")
	buffer.WriteString("CALSCALE:GREGORIAN\r\n")
	buffer.WriteString("METHOD:PUBLISH\r\n")
	buffer.WriteString("BEGIN:VEVENT\r\n")
	buffer.WriteString(fmt.Sprintf("UID:%s\r\n", uid))
	buffer.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", now))
	buffer.WriteString(fmt.Sprintf("DTSTART:%s\r\n", startTime))
	buffer.WriteString(fmt.Sprintf("DTEND:%s\r\n", endTime))
	buffer.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", summary))
	
	// Add description with line folding
	descLines := foldLines(fmt.Sprintf("DESCRIPTION:%s", description))
	buffer.WriteString(descLines)
	
	// Add location if available
	if operationName != "" {
		buffer.WriteString(fmt.Sprintf("LOCATION:%s\r\n", operationName))
	}
	
	// Add recurrence rule based on pattern
	var rrule string
	switch recurringAppointment.Pattern {
	case models.RecurrenceDaily:
		rrule = "RRULE:FREQ=DAILY"
	
	case models.RecurrenceWeekly:
		// Convert weekdays to iCalendar format (SU,MO,TU,WE,TH,FR,SA)
		var days []string
		for _, day := range recurringAppointment.WeekDays {
			switch day {
			case models.Sunday:
				days = append(days, "SU")
			case models.Monday:
				days = append(days, "MO")
			case models.Tuesday:
				days = append(days, "TU")
			case models.Wednesday:
				days = append(days, "WE")
			case models.Thursday:
				days = append(days, "TH")
			case models.Friday:
				days = append(days, "FR")
			case models.Saturday:
				days = append(days, "SA")
			}
		}
		rrule = fmt.Sprintf("RRULE:FREQ=WEEKLY;BYDAY=%s", strings.Join(days, ","))
	
	case models.RecurrenceBiweekly:
		// Similar to weekly but with interval=2
		var days []string
		for _, day := range recurringAppointment.WeekDays {
			switch day {
			case models.Sunday:
				days = append(days, "SU")
			case models.Monday:
				days = append(days, "MO")
			case models.Tuesday:
				days = append(days, "TU")
			case models.Wednesday:
				days = append(days, "WE")
			case models.Thursday:
				days = append(days, "TH")
			case models.Friday:
				days = append(days, "FR")
			case models.Saturday:
				days = append(days, "SA")
			}
		}
		rrule = fmt.Sprintf("RRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=%s", strings.Join(days, ","))
	
	case models.RecurrenceMonthly:
		if recurringAppointment.MonthDay != nil {
			rrule = fmt.Sprintf("RRULE:FREQ=MONTHLY;BYMONTHDAY=%d", *recurringAppointment.MonthDay)
		} else {
			rrule = "RRULE:FREQ=MONTHLY"
		}
	}
	
	// Add end date or count if specified
	if recurringAppointment.EndDate != nil {
		endDate := recurringAppointment.EndDate.UTC().Format("20060102T235959Z")
		rrule += fmt.Sprintf(";UNTIL=%s", endDate)
	} else if recurringAppointment.MaxOccurrences != nil {
		rrule += fmt.Sprintf(";COUNT=%d", *recurringAppointment.MaxOccurrences)
	}
	
	buffer.WriteString(rrule + "\r\n")
	
	// Add exclusion dates if any
	if len(recurringAppointment.ExclusionDates) > 0 {
		var exDates []string
		for _, date := range recurringAppointment.ExclusionDates {
			exDates = append(exDates, date.Format("20060102"))
		}
		buffer.WriteString(fmt.Sprintf("EXDATE;VALUE=DATE:%s\r\n", strings.Join(exDates, ",")))
	}
	
	// Add status
	buffer.WriteString("STATUS:CONFIRMED\r\n")
	
	// Add URL to the recurring appointment in the system
	buffer.WriteString(fmt.Sprintf("URL:%s/recurring-appointments/%d\r\n", s.baseURL, recurringAppointment.ID))
	
	// Add organizer if we have employee information
	if employeeName != "" {
		buffer.WriteString(fmt.Sprintf("ORGANIZER;CN=%s:mailto:noreply@example.com\r\n", employeeName))
	}
	
	buffer.WriteString("END:VEVENT\r\n")
	buffer.WriteString("END:VCALENDAR\r\n")
	
	return buffer.String(), nil
}

// CreateGoogleCalendarEvent creates a new event in Google Calendar
func (s *calendarService) CreateGoogleCalendarEvent(ctx context.Context, appointment *models.Appointment, calendarID string, accessToken string) (string, error) {
	// Initialize Google Calendar API client
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}))
	
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", fmt.Errorf("failed to create Google Calendar service: %v", err)
	}
	
	// Retrieve related entities for more detailed calendar entry
	var supplierName, employeeName, operationName, productName string
	
	// Get supplier name
	supplier, err := s.supplierRepo.GetByID(appointment.SupplierID)
	if err == nil && supplier != nil {
		supplierName = supplier.Name
	}
	
	// Get employee name
	employee, err := s.employeeRepo.GetByID(appointment.EmployeeID)
	if err == nil && employee != nil {
		employeeName = employee.Name
	}
	
	// Create appointment summary and description
	summary := fmt.Sprintf("Delivery: %s", productName)
	if supplierName != "" {
		summary = fmt.Sprintf("Delivery from %s", supplierName)
	}
	
	description := fmt.Sprintf("Supplier: %s\nEmployee: %s\nOperation: %s\nProduct: %s\nQuantity: %d\nStatus: %s",
		supplierName, employeeName, operationName, productName, appointment.QuantityToDeliver, appointment.Status)
	
	if appointment.Notes != "" {
		description += fmt.Sprintf("\n\nNotes: %s", appointment.Notes)
	}
	
	// Add a link back to the appointment in our system
	description += fmt.Sprintf("\n\nView in Scheduling Portal: %s/appointments/%d", s.baseURL, appointment.ID)
	
	// Create Google Calendar event
	event := &calendar.Event{
		Summary:     summary,
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: appointment.ScheduledStart.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		End: &calendar.EventDateTime{
			DateTime: appointment.ScheduledEnd.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		Location: operationName,
		Source: &calendar.EventSource{
			Title: "Scheduling Portal",
			Url:   fmt.Sprintf("%s/appointments/%d", s.baseURL, appointment.ID),
		},
	}
	
	// Set status based on appointment status
	switch appointment.Status {
	case models.StatusConfirmed:
		event.Status = "confirmed"
	case models.StatusCancelled:
		event.Status = "cancelled"
	default:
		event.Status = "tentative"
	}
	
	// Insert the event
	createdEvent, err := srv.Events.Insert(calendarID, event).Do()
	if err != nil {
		return "", fmt.Errorf("failed to create Google Calendar event: %v", err)
	}
	
	return createdEvent.Id, nil
}

// UpdateGoogleCalendarEvent updates an existing event in Google Calendar
func (s *calendarService) UpdateGoogleCalendarEvent(ctx context.Context, appointment *models.Appointment, eventID string, calendarID string, accessToken string) error {
	// Initialize Google Calendar API client
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}))
	
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("failed to create Google Calendar service: %v", err)
	}
	
	// Retrieve the existing event
	existingEvent, err := srv.Events.Get(calendarID, eventID).Do()
	if err != nil {
		return fmt.Errorf("failed to retrieve Google Calendar event: %v", err)
	}
	
	// Retrieve related entities for more detailed calendar entry
	var supplierName, employeeName, operationName, productName string
	
	// Get supplier name
	supplier, err := s.supplierRepo.GetByID(appointment.SupplierID)
	if err == nil && supplier != nil {
		supplierName = supplier.Name
	}
	
	// Get employee name
	employee, err := s.employeeRepo.GetByID(appointment.EmployeeID)
	if err == nil && employee != nil {
		employeeName = employee.Name
	}
	
	// Create appointment summary and description
	summary := fmt.Sprintf("Delivery: %s", productName)
	if supplierName != "" {
		summary = fmt.Sprintf("Delivery from %s", supplierName)
	}
	
	description := fmt.Sprintf("Supplier: %s\nEmployee: %s\nOperation: %s\nProduct: %s\nQuantity: %d\nStatus: %s",
		supplierName, employeeName, operationName, productName, appointment.QuantityToDeliver, appointment.Status)
	
	if appointment.Notes != "" {
		description += fmt.Sprintf("\n\nNotes: %s", appointment.Notes)
	}
	
	// Add a link back to the appointment in our system
	description += fmt.Sprintf("\n\nView in Scheduling Portal: %s/appointments/%d", s.baseURL, appointment.ID)
	
	// Update event fields
	existingEvent.Summary = summary
	existingEvent.Description = description
	existingEvent.Start = &calendar.EventDateTime{
		DateTime: appointment.ScheduledStart.Format(time.RFC3339),
		TimeZone: "UTC",
	}
	existingEvent.End = &calendar.EventDateTime{
		DateTime: appointment.ScheduledEnd.Format(time.RFC3339),
		TimeZone: "UTC",
	}
	existingEvent.Location = operationName
	
	// Set status based on appointment status
	switch appointment.Status {
	case models.StatusConfirmed:
		existingEvent.Status = "confirmed"
	case models.StatusCancelled:
		existingEvent.Status = "cancelled"
	default:
		existingEvent.Status = "tentative"
	}
	
	// Update the event
	_, err = srv.Events.Update(calendarID, eventID, existingEvent).Do()
	if err != nil {
		return fmt.Errorf("failed to update Google Calendar event: %v", err)
	}
	
	return nil
}

// DeleteGoogleCalendarEvent deletes an event from Google Calendar
func (s *calendarService) DeleteGoogleCalendarEvent(ctx context.Context, eventID string, calendarID string, accessToken string) error {
	// Initialize Google Calendar API client
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}))
	
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("failed to create Google Calendar service: %v", err)
	}
	
	// Delete the event
	err = srv.Events.Delete(calendarID, eventID).Do()
	if err != nil {
		return fmt.Errorf("failed to delete Google Calendar event: %v", err)
	}
	
	return nil
}

// SyncAppointmentToCalendar syncs an appointment to a user's external calendar
func (s *calendarService) SyncAppointmentToCalendar(ctx context.Context, appointment *models.Appointment, userID uint, provider CalendarProvider) (string, error) {
	// Get user's calendar preferences
	preferences, err := s.GetUserCalendarPreferences(userID)
	if err != nil {
		return "", fmt.Errorf("failed to get calendar preferences: %v", err)
	}
	
	// Check if sync is enabled for this provider
	enabledKey := fmt.Sprintf("%s_enabled", provider)
	if enabled, ok := preferences[enabledKey].(bool); !ok || !enabled {
		return "", fmt.Errorf("calendar sync not enabled for provider: %s", provider)
	}
	
	var externalEventID string
	
	// Sync based on provider
	switch provider {
	case GoogleCalendar:
		// Get Google Calendar settings
		accessToken, ok := preferences["google_access_token"].(string)
		if !ok || accessToken == "" {
			return "", fmt.Errorf("Google Calendar access token not found")
		}
		
		calendarID, ok := preferences["google_calendar_id"].(string)
		if !ok || calendarID == "" {
			calendarID = "primary" // Default to primary calendar
		}
		
		// Check if this appointment is already synced
		existingSync, err := s.calendarSyncRepo.GetByAppointmentAndProvider(appointment.ID, string(GoogleCalendar))
		if err == nil && existingSync != nil {
			// Update existing event
			err = s.UpdateGoogleCalendarEvent(ctx, appointment, existingSync.ExternalEventID, calendarID, accessToken)
			if err != nil {
				return "", fmt.Errorf("failed to update Google Calendar event: %v", err)
			}
			externalEventID = existingSync.ExternalEventID
		} else {
			// Create new event
			externalEventID, err = s.CreateGoogleCalendarEvent(ctx, appointment, calendarID, accessToken)
			if err != nil {
				return "", fmt.Errorf("failed to create Google Calendar event: %v", err)
			}
			
			// Save the sync record
			syncRecord := &models.CalendarSync{
				UserID:          userID,
				AppointmentID:   appointment.ID,
				Provider:        string(GoogleCalendar),
				ExternalEventID: externalEventID,
				LastSynced:      time.Now(),
			}
			
			err = s.calendarSyncRepo.Create(syncRecord)
			if err != nil {
				// Try to delete the event if we couldn't save the sync record
				s.DeleteGoogleCalendarEvent(ctx, externalEventID, calendarID, accessToken)
				return "", fmt.Errorf("failed to save calendar sync record: %v", err)
			}
		}
		
	case OutlookCalendar:
		// Outlook integration would be implemented similarly to Google
		return "", fmt.Errorf("Outlook Calendar integration not yet implemented")
		
	case ICalFormat:
		// For iCal, we just generate the file and return an ID that can be used to download it
		externalEventID = uuid.New().String()
		
		// Save the sync record
		syncRecord := &models.CalendarSync{
			UserID:          userID,
			AppointmentID:   appointment.ID,
			Provider:        string(ICalFormat),
			ExternalEventID: externalEventID,
			LastSynced:      time.Now(),
		}
		
		err = s.calendarSyncRepo.Create(syncRecord)
		if err != nil {
			return "", fmt.Errorf("failed to save calendar sync record: %v", err)
		}
		
	default:
		return "", fmt.Errorf("unsupported calendar provider: %s", provider)
	}
	
	return externalEventID, nil
}

// RemoveAppointmentFromCalendar removes an appointment from a user's external calendar
func (s *calendarService) RemoveAppointmentFromCalendar(ctx context.Context, appointment *models.Appointment, userID uint, provider CalendarProvider) error {
	// Get user's calendar preferences
	preferences, err := s.GetUserCalendarPreferences(userID)
	if err != nil {
		return fmt.Errorf("failed to get calendar preferences: %v", err)
	}
	
	// Check if this appointment is synced
	existingSync, err := s.calendarSyncRepo.GetByAppointmentAndProvider(appointment.ID, string(provider))
	if err != nil || existingSync == nil {
		return fmt.Errorf("appointment not synced with this calendar provider")
	}
	
	// Remove based on provider
	switch provider {
	case GoogleCalendar:
		// Get Google Calendar settings
		accessToken, ok := preferences["google_access_token"].(string)
		if !ok || accessToken == "" {
			return fmt.Errorf("Google Calendar access token not found")
		}
		
		calendarID, ok := preferences["google_calendar_id"].(string)
		if !ok || calendarID == "" {
			calendarID = "primary" // Default to primary calendar
		}
		
		// Delete the event
		err = s.DeleteGoogleCalendarEvent(ctx, existingSync.ExternalEventID, calendarID, accessToken)
		if err != nil {
			return fmt.Errorf("failed to delete Google Calendar event: %v", err)
		}
		
	case OutlookCalendar:
		// Outlook integration would be implemented similarly to Google
		return fmt.Errorf("Outlook Calendar integration not yet implemented")
		
	case ICalFormat:
		// For iCal, we don't need to do anything with external systems
		// The record just helps us track which appointments have iCal files generated
		
	default:
		return fmt.Errorf("unsupported calendar provider: %s", provider)
	}
	
	// Delete the sync record
	return s.calendarSyncRepo.Delete(existingSync.ID)
}

// GenerateGoogleCalendarLink generates a URL for adding an appointment to Google Calendar
func (s *calendarService) GenerateGoogleCalendarLink(appointment *models.Appointment) string {
	// Retrieve related entities for more detailed calendar entry
	var supplierName, employeeName, operationName, productName string
	
	// Get supplier name
	supplier, err := s.supplierRepo.GetByID(appointment.SupplierID)
	if err == nil && supplier != nil {
		supplierName = supplier.Name
	}
	
	// Get employee name
	employee, err := s.employeeRepo.GetByID(appointment.EmployeeID)
	if err == nil && employee != nil {
		employeeName = employee.Name
	}
	
	// Create appointment summary and description
	summary := fmt.Sprintf("Delivery:

