package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"sync"
	"text/template"
	"time"

	"github.com/bernardofernandezz/scheduling-api/internal/config"
	"github.com/bernardofernandezz/scheduling-api/internal/models"
	"github.com/bernardofernandezz/scheduling-api/internal/repository"
)

// NotificationService defines the interface for notification operations
type NotificationService interface {
	// Notification creation and management
	CreateNotification(notification *models.Notification) error
	GetNotificationByID(id uint) (*models.Notification, error)
	GetNotificationsByRecipient(recipientType models.NotificationRecipientType, recipientID uint) ([]models.Notification, error)
	UpdateNotificationStatus(id uint, status models.NotificationStatus, errorMsg *string) error
	CancelNotification(id uint) error
	
	// Template management
	GetTemplateByEvent(event models.NotificationEvent, recipientType models.NotificationRecipientType, notificationType models.NotificationType) (*models.NotificationTemplate, error)
	RenderTemplate(template *models.NotificationTemplate, data map[string]interface{}) (subject string, bodyText string, bodyHTML string, err error)
	
	// Notification sending
	SendNotification(notification *models.Notification) error
	SendEmail(to string, subject string, bodyText string, bodyHTML string) error
	SendSMS(to string, message string) error
	SendPush(userID uint, title string, message string, data map[string]interface{}) error
	
	// Queue management
	EnqueueNotification(notification *models.Notification, queueName string, priority int) error
	ProcessQueue(queueName string, batchSize int) error
	
	// Appointment event notifications
	NotifyAppointmentCreated(appointment *models.Appointment) error
	NotifyAppointmentUpdated(appointment *models.Appointment, changes map[string]interface{}) error
	NotifyAppointmentStatusChanged(appointment *models.Appointment, oldStatus models.AppointmentStatus) error
	ScheduleAppointmentReminder(appointment *models.Appointment, hoursBeforeAppointment int) error
}

// notificationService implements the NotificationService interface
type notificationService struct {
	notificationRepo   repository.NotificationRepository
	templateRepo       repository.NotificationTemplateRepository
	queueRepo          repository.NotificationQueueRepository
	preferenceRepo     repository.NotificationPreferenceRepository
	userRepo           repository.UserRepository
	employeeRepo       repository.EmployeeRepository
	supplierRepo       repository.SupplierRepository
	config             *config.Config
	
	// Worker pool for processing notifications
	workerPool         chan struct{}
	workerPoolSize     int
	workerMutex        sync.Mutex
	workerID           string
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	notificationRepo repository.NotificationRepository,
	templateRepo repository.NotificationTemplateRepository,
	queueRepo repository.NotificationQueueRepository,
	preferenceRepo repository.NotificationPreferenceRepository,
	userRepo repository.UserRepository,
	employeeRepo repository.EmployeeRepository,
	supplierRepo repository.SupplierRepository,
	config *config.Config,
) NotificationService {
	// Initialize worker pool
	workerPoolSize := 5 // Default worker pool size
	if config != nil && config.Notification != nil && config.Notification.WorkerPoolSize > 0 {
		workerPoolSize = config.Notification.WorkerPoolSize
	}

	return &notificationService{
		notificationRepo:   notificationRepo,
		templateRepo:       templateRepo,
		queueRepo:          queueRepo,
		preferenceRepo:     preferenceRepo,
		userRepo:           userRepo,
		employeeRepo:       employeeRepo,
		supplierRepo:       supplierRepo,
		config:             config,
		workerPool:         make(chan struct{}, workerPoolSize),
		workerPoolSize:     workerPoolSize,
		workerID:           fmt.Sprintf("worker-%d", time.Now().UnixNano()),
	}
}

// CreateNotification creates a new notification
func (s *notificationService) CreateNotification(notification *models.Notification) error {
	// Set default status if not provided
	if notification.Status == "" {
		notification.Status = models.NotificationStatusPending
	}
	
	// Validate notification
	if notification.Type == "" {
		return errors.New("notification type is required")
	}
	if notification.Event == "" {
		return errors.New("notification event is required")
	}
	if notification.RecipientType == "" {
		return errors.New("recipient type is required")
	}
	if notification.RecipientID == 0 {
		return errors.New("recipient ID is required")
	}
	
	// If template ID is provided but no content, render the template
	if notification.TemplateID != nil && *notification.TemplateID != "" && 
	   (notification.Subject == "" || notification.Body == "") {
		// Parse template data
		var templateData map[string]interface{}
		if notification.TemplateData != "" {
			if err := json.Unmarshal([]byte(notification.TemplateData), &templateData); err != nil {
				return fmt.Errorf("failed to parse template data: %w", err)
			}
		}
		
		// Fetch template
		template, err := s.templateRepo.GetByID(notification.TemplateID)
		if err != nil {
			return fmt.Errorf("failed to fetch template: %w", err)
		}
		
		// Render template
		subject, bodyText, bodyHTML, err := s.RenderTemplate(template, templateData)
		if err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}
		
		// Set notification content
		notification.Subject = subject
		if notification.Type == models.NotificationTypeEmail {
			notification.Body = bodyHTML
		} else {
			notification.Body = bodyText
		}
	}
	
	// Create notification in repository
	return s.notificationRepo.Create(notification)
}

// GetNotificationByID retrieves a notification by ID
func (s *notificationService) GetNotificationByID(id uint) (*models.Notification, error) {
	return s.notificationRepo.GetByID(id)
}

// GetNotificationsByRecipient retrieves notifications for a specific recipient
func (s *notificationService) GetNotificationsByRecipient(recipientType models.NotificationRecipientType, recipientID uint) ([]models.Notification, error) {
	return s.notificationRepo.GetByRecipient(recipientType, recipientID)
}

// UpdateNotificationStatus updates a notification's status
func (s *notificationService) UpdateNotificationStatus(id uint, status models.NotificationStatus, errorMsg *string) error {
	notification, err := s.notificationRepo.GetByID(id)
	if err != nil {
		return err
	}
	
	notification.Status = status
	
	if errorMsg != nil {
		notification.ErrorMessage = errorMsg
	}
	
	if status == models.NotificationStatusSent {
		now := time.Now()
		notification.SentAt = &now
	}
	
	return s.notificationRepo.Update(notification)
}

// CancelNotification cancels a pending notification
func (s *notificationService) CancelNotification(id uint) error {
	notification, err := s.notificationRepo.GetByID(id)
	if err != nil {
		return err
	}
	
	// Only pending notifications can be cancelled
	if notification.Status != models.NotificationStatusPending {
		return errors.New("only pending notifications can be cancelled")
	}
	
	notification.Status = models.NotificationStatusCancelled
	
	return s.notificationRepo.Update(notification)
}

// GetTemplateByEvent retrieves a template for a specific event, recipient type, and notification type
func (s *notificationService) GetTemplateByEvent(event models.NotificationEvent, recipientType models.NotificationRecipientType, notificationType models.NotificationType) (*models.NotificationTemplate, error) {
	return s.templateRepo.GetByEvent(event, recipientType, notificationType)
}

// RenderTemplate renders a notification template with the provided data
func (s *notificationService) RenderTemplate(template *models.NotificationTemplate, data map[string]interface{}) (subject string, bodyText string, bodyHTML string, err error) {
	// Render subject
	subjectTmpl, err := textTemplate.New("subject").Parse(template.Subject)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse subject template: %w", err)
	}
	
	var subjectBuf bytes.Buffer
	if err := subjectTmpl.Execute(&subjectBuf, data); err != nil {
		return "", "", "", fmt.Errorf("failed to render subject template: %w", err)
	}
	subject = subjectBuf.String()
	
	// Render body text
	bodyTextTmpl, err := textTemplate.New("bodyText").Parse(template.BodyText)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse body text template: %w", err)
	}
	
	var bodyTextBuf bytes.Buffer
	if err := bodyTextTmpl.Execute(&bodyTextBuf, data); err != nil {
		return "", "", "", fmt.Errorf("failed to render body text template: %w", err)
	}
	bodyText = bodyTextBuf.String()
	
	// Render body HTML if available
	if template.BodyHTML != "" {
		bodyHTMLTmpl, err := htmlTemplate.New("bodyHTML").Parse(template.BodyHTML)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to parse body HTML template: %w", err)
		}
		
		var bodyHTMLBuf bytes.Buffer
		if err := bodyHTMLTmpl.Execute(&bodyHTMLBuf, data); err != nil {
			return "", "", "", fmt.Errorf("failed to render body HTML template: %w", err)
		}
		bodyHTML = bodyHTMLBuf.String()
	}
	
	return subject, bodyText, bodyHTML, nil
}

// SendNotification sends a notification based on its type
func (s *notificationService) SendNotification(notification *models.Notification) error {
	// Update notification status to sending
	notification.Status = models.NotificationStatusSending
	if err := s.notificationRepo.Update(notification); err != nil {
		return err
	}
	
	var err error
	errorMsg := ""
	
	// Get recipient contact information based on recipient type
	var email string
	var phoneNumber string
	var userID uint
	
	switch notification.RecipientType {
	case models.RecipientSupplier:
		supplier, err := s.supplierRepo.GetByID(notification.RecipientID)
		if err != nil {
			errorMsg = fmt.Sprintf("failed to get supplier: %s", err.Error())
			goto updateStatus
		}
		
		// Get user associated with this supplier
		user, err := s.userRepo.GetByID(supplier.UserID)
		if err != nil {
			errorMsg = fmt.Sprintf("failed to get supplier user: %s", err.Error())
			goto updateStatus
		}
		
		email = user.Email
		userID = user.ID
		
		// Get phone from notification preferences if available
		prefs, err := s.preferenceRepo.GetByUserID(user.ID)
		if err == nil && prefs != nil {
			phoneNumber = prefs.PhoneNumber
		}
		
	case models.RecipientEmployee:
		employee, err := s.employeeRepo.GetByID(notification.RecipientID)
		if err != nil {
			errorMsg = fmt.Sprintf("failed to get employee: %s", err.Error())
			goto updateStatus
		}
		
		// Get user associated with this employee
		user, err := s.userRepo.GetByID(employee.UserID)
		if err != nil {
			errorMsg = fmt.Sprintf("failed to get employee user: %s", err.Error())
			goto updateStatus
		}
		
		email = user.Email
		userID = user.ID
		
		// Get phone from notification preferences if available
		prefs, err := s.preferenceRepo.GetByUserID(user.ID)
		if err == nil && prefs != nil {
			phoneNumber = prefs.PhoneNumber
		}
		
	case models.RecipientAdmin:
		// Get admin user
		user, err := s.userRepo.GetByID(notification.RecipientID)
		if err != nil {
			errorMsg = fmt.Sprintf("failed to get admin user: %s", err.Error())
			goto updateStatus
		}
		
		email = user.Email
		userID = user.ID
		
		// Get phone from notification preferences if available
		prefs, err := s.preferenceRepo.GetByUserID(user.ID)
		if err == nil && prefs != nil {
			phoneNumber = prefs.PhoneNumber
		}
	}
	
	// Check user notification preferences
	prefs, err := s.preferenceRepo.GetByUserID(userID)
	if err == nil && prefs != nil {
		// Parse event preferences
		var eventPrefs map[string]bool
		if prefs.EventPrefs != "" {
			if err := json.Unmarshal([]byte(prefs.EventPrefs), &eventPrefs); err == nil {
				// Check if this event type is disabled
				if enabled, exists := eventPrefs[string(notification.Event)]; exists && !enabled {
					errorMsg = "notification disabled by user preferences"
					goto updateStatus
				}
			}
		}
		
		// Check if this notification type is enabled
		switch notification.Type {
		case models.NotificationTypeEmail:
			if !prefs.EmailEnabled {
				errorMsg = "email notifications disabled by user preferences"
				goto updateStatus
			}
		case models.NotificationTypeSMS:
			if !prefs.SMSEnabled {
				errorMsg = "SMS notifications disabled by user preferences"
				goto updateStatus
			}
		case models.NotificationTypePush:
			if !prefs.PushEnabled {
				errorMsg = "push notifications disabled by user preferences"
				goto updateStatus
			}
		}
	}
	
	// Send notification based on type
	switch notification.Type {
	case models.NotificationTypeEmail:
		if email == "" {
			errorMsg = "recipient email address not available"
			goto updateStatus
		}
		
		// Extract or generate HTML version if needed
		bodyHTML := notification.Body
		bodyText := notification.Body
		
		// Attempt to extract metadata for additional content
		var metadata map[string]interface{}
		if notification.Metadata != "" {
			if err := json.Unmarshal([]byte(notification.Metadata), &metadata); err == nil {
				if textContent, ok := metadata["text_content"].(string); ok {
					bodyText = textContent
				}
			}
		}
		
		err = s.SendEmail(email, notification.Subject, bodyText, bodyHTML)
		if err != nil {
			errorMsg = fmt.Sprintf("failed to send email: %s", err.Error())
		}
		
	case models.NotificationTypeSMS:
		if phoneNumber == "" {
			errorMsg = "recipient phone number not available"
			goto updateStatus
		}
		
		err = s.SendSMS(phoneNumber, notification.Body)
		if err != nil {
			errorMsg = fmt.Sprintf("failed to send SMS: %s", err.Error())
		}
		
	case models.NotificationTypePush:
		// Extract additional data from metadata if available
		var pushData map[string]interface{}
		if notification.Metadata != "" {
			if err := json.Unmarshal([]byte(notification.Metadata), &pushData); err != nil {
				pushData = make(map[string]interface{})
			}
		}
		
		err = s.SendPush(userID, notification.Subject, notification.Body, pushData)
		if err != nil {
			errorMsg = fmt.Sprintf("failed to send push notification: %s", err.Error())
		}
	}
	
updateStatus:
	// Update notification status based on result
	if errorMsg != "" {
		notification.Status = models.NotificationStatusFailed
		notification.ErrorMessage = &errorMsg
		notification.RetryCount++
		
		// If retry count is less than max retries, requeue for later
		if notification.RetryCount < notification.MaxRetries {
			// Schedule retry with exponential backoff (5 min, 15 min, 45 min)
			backoffMinutes := 5 * (1 << (notification.RetryCount - 1))
			scheduledFor := time.Now().Add(time.Duration(backoffMinutes) * time.Minute)
			notification.ScheduledFor = &scheduledFor
			notification.Status = models.NotificationStatusPending
		}
	} else {
		// Success
		notification.Status = models.NotificationStatusSent
		now := time.Now()
		notification.SentAt = &now
	}
	
	return s.notificationRepo.Update(notification)
}

// SendEmail sends an email notification
func (s *notificationService) SendEmail(to string, subject string, bodyText string, bodyHTML string) error {
	// For this example, we'll log the email rather than actually sending it
	// In a real implementation, you would integrate with an email provider (SendGrid, Mailgun, etc.)
	log.Printf("EMAIL TO: %s, SUBJECT: %s\nTEXT: %s\nHTML: %s", to, subject, bodyText, bodyHTML)
	
	// TODO: Implement actual email sending logic
	// This would typically integrate with a third-party email service
	
	return nil
}

// SendSMS sends an SMS notification
func (s *notificationService) SendSMS(to string, message string) error {
	// For this example, we'll log the SMS rather than actually sending it
	// In a real implementation, you would integrate with an SMS provider (Twilio, etc.)
	log.Printf("SMS TO: %s, MESSAGE: %s", to, message)
	
	// TODO: Implement actual SMS sending logic
	// This would typically integrate with a third-party SMS service
	
	return nil
}

// SendPush sends a push notification
func (s *notificationService) SendPush(userID uint, title string, message string, data map[string]interface{}) error {
	// For this example, we'll log the push notification rather than actually sending it
	// In a real implementation, you would integrate with a push provider (Firebase, etc.)
	dataJson, _ := json.Marshal(data)
	log.Printf("PUSH TO USER: %d, TITLE: %s, MESSAGE: %s, DATA: %s", userID, title, message, dataJson)
	
	// TODO: Implement actual push notification sending logic
	// This would typically integrate with a third-party push notification service
	
	return nil
}

// EnqueueNotification adds a notification to the processing queue
func (s *notificationService) EnqueueNotification(notification *models.Notification, queueName string, priority int) error {
	// Create notification if it doesn't exist
	if notification.ID == 0 {
		if err := s.CreateNotification(notification); err != nil {
			return fmt.Errorf("failed to create notification: %w", err)
		}
	}
	
	// Create queue entry
	queue := &models.NotificationQueue{
		QueueName:      queueName,
		Priority:       priority,
		NotificationID: notification.ID,
		Status:         models.NotificationStatusPending,
	}
	
	return s.queueRepo.Create(queue)
}

// ProcessQueue processes notifications from the queue
func (s *notificationService) ProcessQueue(queueName string, batchSize int) error {
	// Lock to prevent multiple workers from processing the same queue
	s.workerMutex.Lock()
	defer s.workerMutex.Unlock()
	
	// Get the next batch of notifications to process, ordered by priority and creation time
	queueItems, err := s.queueRepo.GetPendingByQueue(queueName, batchSize)
	if err != nil {
		return err
	}
	
	if len(queueItems) == 0 {
		return nil // Nothing to process
	}
	
	// Process each queue item
	for _, item := range queueItems {
		// Lock this item for processing
		now := time.Now()
		lockUntil := now.Add(5 * time.Minute) // Lock for 5 minutes
		item.LockedUntil = &lockUntil
		item.ProcessorID = &s.workerID
		item.Status = models.NotificationStatusSending
		
		if err := s.queueRepo.Update(&item); err != nil {
			log.Printf("Failed to lock queue item %d: %v", item.ID, err)
			continue
		}
		
		// Get the notification
		notification, err := s.notificationRepo.GetByID(item.NotificationID)
		if err != nil {
			log.Printf("Failed to get notification %d: %v", item.NotificationID, err)
			continue
		}
		
		// If notification is scheduled for the future, skip it
		if notification.ScheduledFor != nil && notification.ScheduledFor.After(now) {
			item.Status = models.NotificationStatusPending
			item.LockedUntil = nil
			item.ProcessorID = nil
			s.queueRepo.Update(&item)
			continue
		}
		
		// Process the notification in a worker from the pool
		s.workerPool <- struct{}{} // Acquire a worker
		go func(item models.NotificationQueue, notification *models.Notification) {
			defer func() {
				<-s.workerPool // Release the worker
			}()
			
			// Send the notification
			err := s.SendNotification(notification)
			if err != nil {
				log.Printf("Failed to send notification %d: %v", notification.ID, err)
			}
			
			// Update queue item status
			processed := time.Now()
			item.ProcessedAt = &processed
			item.Status = notification.Status
			s.queueRepo.Update(&item)
		}(item, notification)
	}
	
	return nil
}

// NotifyAppointmentCreated sends notifications when a new appointment is created
func (s *notificationService) NotifyAppointmentCreated(appointment *models.Appointment) error {
	// Prepare common template data
	templateData := map[string]interface{}{
		"appointment_id":      appointment.ID,
		"supplier_id":         appointment.SupplierID,
		"employee_id":         appointment.EmployeeID,
		"operation_id":        appointment.OperationID,
		"product_id":          appointment.ProductID,
		"scheduled_start":     appointment.ScheduledStart.Format(time.RFC3339),
		"scheduled_end":       appointment.ScheduledEnd.Format(time.RFC3339),
		"scheduled_date":      appointment.ScheduledStart.Format("Monday, January 2, 2006"),
		"scheduled_time":      appointment.ScheduledStart.Format("3:04 PM"),
		"quantity_to_deliver": appointment.QuantityToDeliver,
		"status":              string(appointment.Status),
		"notes":               appointment.Notes,
	}
	
	// Convert template data to JSON
	templateDataJSON, err := json.Marshal(templateData)
	if err != nil {
		return fmt.Errorf("failed to marshal template data: %w", err)
	}
	
	// Create a notification for the supplier
	supplierTemplate, err := s.GetTemplateByEvent(
		models.EventAppointmentCreated,
		models.RecipientSupplier,
		models.NotificationTypeEmail,
	)
	if err == nil && supplierTemplate != nil {
		notification := &models.Notification{
			Type:          models.NotificationTypeEmail,
			Status:        models.NotificationStatusPending,
			Event:         models.EventAppointmentCreated,
			RecipientType: models.RecipientSupplier,
			RecipientID:   appointment.SupplierID,
			TemplateID:    &supplierTemplate.ID,
			TemplateData:  string(templateDataJSON),
			AppointmentID: &appointment.ID,
		}
		
		if err := s.EnqueueNotification(notification, "appointment_notifications", 2); err != nil {
			log.Printf("Failed to enqueue supplier notification for appointment %d: %v", appointment.ID, err)
		}
	}
	
	// Create a notification for the employee
	employeeTemplate, err := s.GetTemplateByEvent(
		models.EventAppointmentCreated,
		models.RecipientEmployee,
		models.NotificationTypeEmail,
	)
	if err == nil && employeeTemplate != nil {
		notification := &models.Notification{
			Type:          models.NotificationTypeEmail,
			Status:        models.NotificationStatusPending,
			Event:         models.EventAppointmentCreated,
			RecipientType: models.RecipientEmployee,
			RecipientID:   appointment.EmployeeID,
			TemplateID:    &employeeTemplate.ID,
			TemplateData:  string(templateDataJSON),
			AppointmentID: &appointment.ID,
		}
		
		if err := s.EnqueueNotification(notification, "appointment_notifications", 2); err != nil {
			log.Printf("Failed to enqueue employee notification for appointment %d: %v", appointment.ID, err)
		}
	}
	
	return nil
}

// NotifyAppointmentUpdated sends notifications when an appointment is updated
func (s *notificationService) NotifyAppointmentUpdated(appointment *models.Appointment, changes map[string]interface{}) error {
	// Prepare common template data
	templateData := map[string]interface{}{
		"appointment_id":      appointment.ID,
		"supplier_id":         appointment.SupplierID,
		"employee_id":         appointment.EmployeeID,
		"operation_id":        appointment.OperationID,
		"product_id":          appointment.ProductID,
		"scheduled_start":     appointment.ScheduledStart.Format(time.RFC3339),
		"scheduled_end":       appointment.ScheduledEnd.Format(time.RFC3339),
		"scheduled_date":      appointment.ScheduledStart.Format("Monday, January 2, 2006"),
		"scheduled_time":      appointment.ScheduledStart.Format("3:04 PM"),
		"quantity_to_deliver": appointment.QuantityToDeliver,
		"status":              string(appointment.Status),
		"notes":               appointment.Notes,
		"changes":             changes,
	}
	
	// Convert template data to JSON
	templateDataJSON, err := json.Marshal(templateData)
	if err != nil {
		return fmt.Errorf("failed to marshal template data: %w", err)
	}
	
	// Create a notification for the supplier
	supplierTemplate, err := s.GetTemplateByEvent(
		models.EventAppointmentUpdated,
		models.RecipientSupplier,
		models.NotificationTypeEmail,
	)
	if err == nil && supplierTemplate != nil {
		notification := &models.Notification{
			Type:          models.NotificationTypeEmail,
			Status:        models.NotificationStatusPending,
			Event:         models.EventAppointmentUpdated,
			RecipientType: models.RecipientSupplier,
			RecipientID:   appointment.SupplierID,
			TemplateID:    &supplierTemplate.ID,
			TemplateData:  string(templateDataJSON),
			AppointmentID: &appointment.ID,
		}
		
		if err := s.EnqueueNotification(notification, "appointment_notifications", 2); err != nil {
			log.Printf("Failed to enqueue supplier notification for appointment %d: %v", appointment.ID, err)
		}
	}
	
	// Create a notification for the employee
	employeeTemplate, err := s.GetTemplateByEvent(
		models.EventAppointmentUpdated,
		models.RecipientEmployee,
		models.NotificationTypeEmail,
	)
	if err == nil && employeeTemplate != nil {
		notification := &models.Notification{
			Type:          models.NotificationTypeEmail,
			Status:        models.NotificationStatusPending,
			Event:         models.EventAppointmentUpdated,
			RecipientType: models.RecipientEmployee,
			RecipientID:   appointment.EmployeeID,
			TemplateID:    &employeeTemplate.ID,
			TemplateData:  string(templateDataJSON),
			AppointmentID: &appointment.ID,
		}
		
		if err := s.EnqueueNotification(notification, "appointment_notifications", 2); err != nil {
			log.Printf("Failed to enqueue employee notification for appointment %d: %v", appointment.ID, err)
		}
	}
	
	return nil
}

// NotifyAppointmentStatusChanged sends notifications when an appointment status changes
func (s *notificationService) NotifyAppointmentStatusChanged(appointment *models.Appointment, oldStatus models.AppointmentStatus) error {
	// Determine the event type based on the new status
	var event models.NotificationEvent
	switch appointment.Status {
	case models.StatusConfirmed:
		event = models.EventAppointmentConfirmed
	case models.StatusCancelled:
		event = models.EventAppointmentCancelled
	case models.StatusCompleted:
		event = models.EventAppointmentCompleted
	default:
		event = models.EventAppointmentUpdated
	}
	
	// Prepare common template data
	templateData := map[string]interface{}{
		"appointment_id":      appointment.ID,
		"supplier_id":         appointment.SupplierID,
		"employee_id":         appointment.EmployeeID,
		"operation_id":        appointment.OperationID,
		"product_id":          appointment.ProductID,
		"scheduled_start":     appointment.ScheduledStart.Format(time.RFC3339),
		"scheduled_end":       appointment.ScheduledEnd.Format(time.RFC3339),
		"scheduled_date":      appointment.ScheduledStart.Format("Monday, January 2, 2006"),
		"scheduled_time":      appointment.ScheduledStart.Format("3:04 PM"),
		"quantity_to_deliver": appointment.QuantityToDeliver,
		"status":              string(appointment.Status),
		"old_status":          string(oldStatus),
		"notes":               appointment.Notes,
	}
	
	// Add cancellation reason if available
	if appointment.Status == models.StatusCancelled && appointment.CancellationReason != "" {
		templateData["cancellation_reason"] = appointment.CancellationReason
	}
	
	// Convert template data to JSON
	templateDataJSON, err := json.Marshal(templateData)
	if err != nil {
		return fmt.Errorf("failed to marshal template data: %w", err)
	}
	
	// Create a notification for the supplier
	supplierTemplate, err := s.GetTemplateByEvent(
		event,
		models.Recipient

