package models

import (
	"time"

	"gorm.io/gorm"
)

// NotificationType defines the type of notification to send
type NotificationType string

const (
	// NotificationTypeEmail indicates an email notification
	NotificationTypeEmail NotificationType = "email"
	
	// NotificationTypeSMS indicates an SMS notification
	NotificationTypeSMS NotificationType = "sms"
	
	// NotificationTypePush indicates a push notification
	NotificationTypePush NotificationType = "push"
)

// NotificationStatus defines the status of a notification
type NotificationStatus string

const (
	// NotificationStatusPending indicates a notification that is waiting to be sent
	NotificationStatusPending NotificationStatus = "pending"
	
	// NotificationStatusSending indicates a notification that is in the process of being sent
	NotificationStatusSending NotificationStatus = "sending"
	
	// NotificationStatusSent indicates a notification that has been successfully sent
	NotificationStatusSent NotificationStatus = "sent"
	
	// NotificationStatusFailed indicates a notification that failed to send
	NotificationStatusFailed NotificationStatus = "failed"
	
	// NotificationStatusCancelled indicates a notification that was cancelled before sending
	NotificationStatusCancelled NotificationStatus = "cancelled"
)

// NotificationEvent defines the event that triggered the notification
type NotificationEvent string

const (
	// EventAppointmentCreated is triggered when a new appointment is created
	EventAppointmentCreated NotificationEvent = "appointment_created"
	
	// EventAppointmentUpdated is triggered when an appointment is updated
	EventAppointmentUpdated NotificationEvent = "appointment_updated"
	
	// EventAppointmentCancelled is triggered when an appointment is cancelled
	EventAppointmentCancelled NotificationEvent = "appointment_cancelled"
	
	// EventAppointmentConfirmed is triggered when an appointment is confirmed
	EventAppointmentConfirmed NotificationEvent = "appointment_confirmed"
	
	// EventAppointmentCompleted is triggered when an appointment is completed
	EventAppointmentCompleted NotificationEvent = "appointment_completed"
	
	// EventAppointmentReminder is triggered to remind about upcoming appointments
	EventAppointmentReminder NotificationEvent = "appointment_reminder"
)

// NotificationRecipientType defines the type of recipient
type NotificationRecipientType string

const (
	// RecipientSupplier indicates the notification is for a supplier
	RecipientSupplier NotificationRecipientType = "supplier"
	
	// RecipientEmployee indicates the notification is for an employee
	RecipientEmployee NotificationRecipientType = "employee"
	
	// RecipientAdmin indicates the notification is for an admin
	RecipientAdmin NotificationRecipientType = "admin"
)

// Notification represents a notification to be sent
type Notification struct {
	gorm.Model
	
	// Basic information
	Type            NotificationType       `json:"type" gorm:"not null"`
	Status          NotificationStatus     `json:"status" gorm:"not null"`
	Event           NotificationEvent      `json:"event" gorm:"not null"`
	RecipientType   NotificationRecipientType `json:"recipient_type" gorm:"not null"`
	RecipientID     uint                   `json:"recipient_id" gorm:"not null"`
	
	// Content information
	Subject         string                 `json:"subject" gorm:"not null"`
	Body            string                 `json:"body" gorm:"not null;type:text"`
	TemplateID      *string                `json:"template_id"`
	TemplateData    string                 `json:"template_data" gorm:"type:text"` // JSON string of template variables
	
	// Related resources
	AppointmentID   *uint                  `json:"appointment_id"`
	Appointment     *Appointment           `json:"appointment" gorm:"foreignKey:AppointmentID"`
	
	// Delivery tracking
	ScheduledFor    *time.Time             `json:"scheduled_for"` // For delayed notifications
	SentAt          *time.Time             `json:"sent_at"`
	DeliveredAt     *time.Time             `json:"delivered_at"`
	ErrorMessage    *string                `json:"error_message"`
	RetryCount      int                    `json:"retry_count" gorm:"default:0"`
	MaxRetries      int                    `json:"max_retries" gorm:"default:3"`
	
	// Metadata
	Metadata        string                 `json:"metadata" gorm:"type:text"` // JSON string for additional data
}

// NotificationTemplate defines templates for different notification events
type NotificationTemplate struct {
	gorm.Model
	
	// Basic information
	Name            string                 `json:"name" gorm:"not null;unique"`
	Description     string                 `json:"description"`
	
	// Content templates
	Subject         string                 `json:"subject" gorm:"not null"`
	BodyText        string                 `json:"body_text" gorm:"type:text;not null"` // Plain text version
	BodyHTML        string                 `json:"body_html" gorm:"type:text"`          // HTML version for email
	
	// Classification
	Type            NotificationType       `json:"type" gorm:"not null"`
	Event           NotificationEvent      `json:"event" gorm:"not null"`
	RecipientType   NotificationRecipientType `json:"recipient_type" gorm:"not null"`
	
	// Status
	IsActive        bool                   `json:"is_active" gorm:"default:true"`
	
	// Variables used in the template, stored as JSON array string
	Variables       string                 `json:"variables" gorm:"type:text"`
}

// NotificationPreference defines user preferences for notifications
type NotificationPreference struct {
	gorm.Model
	
	// User relationship
	UserID          uint                   `json:"user_id" gorm:"not null"`
	User            User                   `json:"user" gorm:"foreignKey:UserID"`
	
	// Notification types enabled
	EmailEnabled    bool                   `json:"email_enabled" gorm:"default:true"`
	SMSEnabled      bool                   `json:"sms_enabled" gorm:"default:false"`
	PushEnabled     bool                   `json:"push_enabled" gorm:"default:false"`
	
	// Event preferences (JSON string mapping event types to bool)
	EventPrefs      string                 `json:"event_prefs" gorm:"type:text"`
	
	// Contact information
	Email           string                 `json:"email"`
	PhoneNumber     string                 `json:"phone_number"`
	
	// Reminder settings
	ReminderHours   int                    `json:"reminder_hours" gorm:"default:24"` // Hours before appointment to send reminder
}

// NotificationQueue represents a queue for processing notifications
type NotificationQueue struct {
	gorm.Model
	
	// Queue information
	QueueName       string                 `json:"queue_name" gorm:"not null"`
	Priority        int                    `json:"priority" gorm:"default:1"` // Higher number = higher priority
	
	// Notification relationship
	NotificationID  uint                   `json:"notification_id" gorm:"not null"`
	Notification    Notification           `json:"notification" gorm:"foreignKey:NotificationID"`
	
	// Processing status
	ProcessedAt     *time.Time             `json:"processed_at"`
	Status          NotificationStatus     `json:"status" gorm:"not null"`
	LockedUntil     *time.Time             `json:"locked_until"` // For distributed processing
	ProcessorID     *string                `json:"processor_id"` // ID of the worker processing this notification
}

