package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/bernardofernandezz/scheduling-api/internal/models"
	"github.com/bernardofernandezz/scheduling-api/internal/service"
)

// AppointmentHandler handles appointment-related requests
type AppointmentHandler struct {
	appointmentService service.AppointmentService
}

// NewAppointmentHandler creates a new appointment handler
func NewAppointmentHandler(appointmentService service.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{
		appointmentService: appointmentService,
	}
}

// CreateAppointmentRequest is the request body for creating an appointment
type CreateAppointmentRequest struct {
	SupplierID        uint      `json:"supplier_id" binding:"required"`
	EmployeeID        uint      `json:"employee_id" binding:"required"`
	OperationID       uint      `json:"operation_id" binding:"required"`
	ProductID         uint      `json:"product_id" binding:"required"`
	ScheduledStart    time.Time `json:"scheduled_start" binding:"required"`
	ScheduledEnd      time.Time `json:"scheduled_end" binding:"required"`
	Notes             string    `json:"notes"`
	QuantityToDeliver int       `json:"quantity_to_deliver" binding:"required,min=1"`
}

// UpdateAppointmentRequest is the request body for updating an appointment
type UpdateAppointmentRequest struct {
	SupplierID        uint                   `json:"supplier_id"`
	EmployeeID        uint                   `json:"employee_id"`
	OperationID       uint                   `json:"operation_id"`
	ProductID         uint                   `json:"product_id"`
	ScheduledStart    time.Time              `json:"scheduled_start"`
	ScheduledEnd      time.Time              `json:"scheduled_end"`
	Status            models.AppointmentStatus `json:"status"`
	Notes             string                 `json:"notes"`
	QuantityToDeliver int                    `json:"quantity_to_deliver" binding:"min=1"`
	CancellationReason string                `json:"cancellation_reason"`
}

// UpdateStatusRequest is the request body for updating an appointment status
type UpdateStatusRequest struct {
	Status models.AppointmentStatus `json:"status" binding:"required"`
	Reason string                  `json:"reason"`
}

// CheckAvailabilityRequest is the request body for checking appointment availability
type CheckAvailabilityRequest struct {
	OperationID    uint      `json:"operation_id" binding:"required"`
	EmployeeID     uint      `json:"employee_id" binding:"required"`
	ScheduledStart time.Time `json:"scheduled_start" binding:"required"`
	ScheduledEnd   time.Time `json:"scheduled_end" binding:"required"`
}

// GetAppointmentFilters parses appointment filters from query parameters
func GetAppointmentFilters(c *gin.Context) repository.AppointmentFilters {
	// Initialize filters
	filters := repository.AppointmentFilters{}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	filters.Page = page
	filters.Limit = limit

	// Parse sorting parameters
	filters.SortBy = c.DefaultQuery("sort_by", "scheduled_start")
	filters.SortOrder = c.DefaultQuery("sort_order", "asc")

	// Parse status filter
	if status := c.Query("status"); status != "" {
		appointmentStatus := models.AppointmentStatus(status)
		filters.Status = &appointmentStatus
	}

	// Parse date filters
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filters.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filters.EndDate = &endDate
		}
	}

	return filters
}

// Create handles creating a new appointment
func (h *AppointmentHandler) Create(c *gin.Context) {
	var req CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Get authenticated user from context (for authorization checks)
	userObj, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user, ok := userObj.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user object"})
		return
	}

	// If user is a supplier, check if they're creating an appointment for themselves
	if user.Role == "supplier" {
		// Find supplier ID for this user
		var supplierID uint = 0
		// In a real app, you'd fetch this from the supplier repository
		// For now, we'll assume supplier validation happens in the service layer

		// If the user is trying to create an appointment for a different supplier
		if supplierID != 0 && supplierID != req.SupplierID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Suppliers can only create appointments for themselves"})
			return
		}
	}

	// Create appointment model from request
	appointment := &models.Appointment{
		SupplierID:        req.SupplierID,
		EmployeeID:        req.EmployeeID,
		OperationID:       req.OperationID,
		ProductID:         req.ProductID,
		ScheduledStart:    req.ScheduledStart,
		ScheduledEnd:      req.ScheduledEnd,
		Notes:             req.Notes,
		QuantityToDeliver: req.QuantityToDeliver,
		Status:            models.StatusPending,
	}

	// Create appointment
	if err := h.appointmentService.Create(appointment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"appointment": appointment})
}

// Get handles getting an appointment by ID
func (h *AppointmentHandler) Get(c *gin.Context) {
	// Parse appointment ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment ID"})
		return
	}

	// Get appointment
	appointment, err := h.appointmentService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Authorization check - user should be related to this appointment or an admin
	userObj, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user, ok := userObj.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user object"})
		return
	}

	// Admin can view all appointments
	if user.Role != "admin" {
		// Suppliers can only view their own appointments
		if user.Role == "supplier" {
			// Check if this supplier is related to the appointment
			// In a real app, you'd fetch supplier ID for this user
			var supplierID uint = 0
			if supplierID != 0 && supplierID != appointment.SupplierID {
				c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this appointment"})
				return
			}
		}

		// Employees can only view appointments where they are the assigned employee
		if user.Role == "employee" {
			// Check if this employee is related to the appointment
			// In a real app, you'd fetch employee ID for this user
			var employeeID uint = 0
			if employeeID != 0 && employeeID != appointment.EmployeeID {
				c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this appointment"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"appointment": appointment})
}

// Update handles updating an appointment
func (h *AppointmentHandler) Update(c *gin.Context) {
	// Parse appointment ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment ID"})
		return
	}

	// Get existing appointment
	existingAppointment, err := h.appointmentService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Authorization check
	userObj, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user, ok := userObj.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user object"})
		return
	}

	// Only admins and the supplier who created the appointment can update it
	if user.Role != "admin" {
		if user.Role == "supplier" {
			// In a real app, you'd fetch supplier ID for this user
			var supplierID uint = 0
			if supplierID != 0 && supplierID != existingAppointment.SupplierID {
				c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this appointment"})
				return
			}
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins and the appointment's supplier can update appointments"})
			return
		}
	}

	// Parse request
	var req UpdateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Update appointment fields that were provided
	if req.SupplierID != 0 {
		existingAppointment.SupplierID = req.SupplierID
	}
	if req.EmployeeID != 0 {
		existingAppointment.EmployeeID = req.EmployeeID
	}
	if req.OperationID != 0 {
		existingAppointment.OperationID = req.OperationID
	}
	if req.ProductID != 0 {
		existingAppointment.ProductID = req.ProductID
	}
	if !req.ScheduledStart.IsZero() {
		existingAppointment.ScheduledStart = req.ScheduledStart
	}
	if !req.ScheduledEnd.IsZero() {
		existingAppointment.ScheduledEnd = req.ScheduledEnd
	}
	if req.Status != "" {
		existingAppointment.Status = req.Status
	}
	if req.Notes != "" {
		existingAppointment.Notes = req.Notes
	}
	if req.QuantityToDeliver > 0 {
		existingAppointment.QuantityToDeliver = req.QuantityToDeliver
	}
	if req.CancellationReason != "" {
		existingAppointment.CancellationReason = req.CancellationReason
	}

	// Update appointment
	if err := h.appointmentService.Update(existingAppointment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"appointment": existingAppointment})
}

// Delete handles deleting an appointment
func (h *AppointmentHandler) Delete(c *gin.Context) {
	// Parse appointment ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment ID"})
		return
	}

	// Get existing appointment
	existingAppointment, err := h.appointmentService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Authorization check
	userObj, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user, ok := userObj.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user object"})
		return
	}

	// Only admins and the supplier who created the appointment can delete it
	if user.Role != "admin" {
		if user.Role == "supplier" {
			// In a real app, you'd fetch supplier ID for this user
			var supplierID uint = 0
			if supplierID != 0 && supplierID != existingAppointment.SupplierID {
				c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this appointment"})
				return
			}
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins and the appointment's supplier can delete appointments"})
			return
		}
	}

	// Delete appointment
	if err := h.appointmentService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Appointment deleted successfully"})
}

// List handles listing appointments with filters
func (h *AppointmentHandler) List(c *gin.Context) {
	// Get filters from query parameters
	filters := GetAppointmentFilters(c)

	// Authorization check
	userObj, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user, ok := userObj.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user object"})
		return
	}

	// Get appointments
	appointments, total, err := h.appointmentService.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If not admin, filter appointments based on user role
	if user.Role != "admin" {
		// This filtering should ideally happen at the service/repository level
		// based on the authenticated user, but we're keeping it simple here
	}

	c.JSON(http.StatusOK, gin.H{
		"appointments": appointments,
		"total":        total,
		"page":         filters.Page,
		"limit":        filters.Limit,
		"total_pages":  (total + int64(filters.Limit) - 1) / int64(filters.Limit),
	})
}

// UpdateStatus handles updating an appointment's status
func (h *AppointmentHandler) UpdateStatus(c *gin.Context) {
	// Parse appointment ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment ID"})
		return
	}

	// Get existing appointment
	existingAppointment, err := h.appointmentService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Authorization check
	userObj, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user, ok := userObj.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user object"})
		return
	}

	// Parse request
	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Check permissions based on the requested status change
	if !hasStatusChangePermission(user, existingAppointment, req.Status) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to change the status to " + string(req.Status)})
		return
	}

	// Update status
	if err := h.appointmentService.UpdateStatus(uint(id), req.Status, req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get updated appointment
	updatedAppointment, err := h.appointmentService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated appointment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"appointment": updatedAppointment})
}

// GetBySupplier handles getting appointments for a specific supplier
func (h *AppointmentHandler) GetBySupplier(c *gin.Context) {
	// Parse supplier ID from path
	idStr := c.Param("supplier_id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid supplier ID"})
		return
	}

	// Get filters from query parameters
	filters := GetAppointmentFilters(c)

	// Authorization check
	userObj, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user, ok := userObj.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user object"})
		return
	}

	// Only admins and the supplier themselves can view their appointments
	if user.Role != "admin" {
		if user.Role == "supplier" {
			// In a real app, you'd fetch supplier ID for this user
			var supplierID uint = 0
			if supplierID != 0 && supplierID != uint(id) {
				c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view these appointments"})
				return
			}
		} else {
			// Employees can view appointments for any supplier
			// This might be adjusted based on business rules
		}
	}

	// Get appointments for the supplier
	appointments, total, err := h.appointmentService.GetBySupplier(uint(id), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"appointments": appointments,
		"total":        total,
		"page":         filters.Page,
		"limit":        filters.Limit,
		"total_pages":  (total + int64(filters.Limit) - 1) / int64(filters.Limit),
	})
}

// GetByEmployee handles getting appointments for a specific employee
func (h *AppointmentHandler) GetByEmployee(c *gin.Context) {
	// Parse employee ID from path
	idStr := c.Param("employee_id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	// Get filters from query parameters
	filters := GetAppointmentFilters(c)

	// Authorization check
	userObj, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user, ok := userObj.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user object"})
		return
	}

	// Only admins and the employee themselves can view their appointments
	if user.Role != "admin" {
		if user.Role == "employee" {
			// In a real app, you'd fetch employee ID for this user
			var employeeID uint = 0
			if employeeID != 0 && employeeID != uint(id) {
				c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view these appointments"})
				return
			}
		} else if user.Role == "supplier" {
			// Suppliers can only view appointments they're involved in
			// This filtering should ideally happen at the service level
		}
	}

	// Get appointments for the employee
	appointments, total, err := h.appointmentService.GetByEmployee(uint(id), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"appointments": appointments,
		"total":        total,
		"page":         filters.Page,
		"limit":        filters.Limit,
		"total_pages":  (total + int64(filters.Limit) - 1) / int64(filters.Limit),
	})
}

// GetByOperation handles getting appointments for a specific operation
func (h *AppointmentHandler) GetByOperation(c *gin.Context) {
	// Parse operation ID from path
	idStr := c.Param("operation_id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid operation ID"})
		return
	}

	// Get filters from query parameters
	filters := GetAppointmentFilters(c)

	// Authorization check
	userObj, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user, ok := userObj.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user object"})
		return
	}

	// Admins can view all operations
	// Employees can view operations they're assigned to
	// Suppliers can view operations they're delivering to
	if user.Role != "admin" {
		// In a real app, you would check if the user has access to this operation
		// This might involve checking if an employee belongs to this operation
		// or if a supplier has any appointments at this operation
	}

	// Get appointments for the operation
	appointments, total, err := h.appointmentService.GetByOperation(uint(id), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"appointments": appointments,
		"total":        total,
		"page":         filters.Page,
		"limit":        filters.Limit,
		"total_pages":  (total + int64(filters.Limit) - 1) / int64(filters.Limit),
	})
}

// GetByDateRange handles getting appointments within a date range
func (h *AppointmentHandler) GetByDateRange(c *gin.Context) {
	// Parse date range parameters
	startDateStr := c.Query("start_date")
	if startDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date is required"})
		return
	}

	endDateStr := c.Query("end_date")
	if endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date is required"})
		return
	}

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use RFC3339 format (e.g., 2025-05-23T10:00:00Z)"})
		return
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use RFC3339 format (e.g., 2025-05-23T18:00:00Z)"})
		return
	}

	if startDate.After(endDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date must be before end_date"})
		return
	}

	// Get filters from query parameters
	filters := GetAppointmentFilters(c)

	// Authorization check
	userObj, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user, ok := userObj.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user object"})
		return
	}

	// For non-admin users, we might want to limit the date range to prevent too many results
	if user.Role != "admin" {
		// Limit to a maximum of 31 days for non-admin users
		maxDateRange := 31 * 24 * time.Hour
		if endDate.Sub(startDate) > maxDateRange {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Date range cannot exceed 31 days for non-admin users"})
			return
		}
	}

	// Get appointments within the date range
	appointments, total, err := h.appointmentService.GetByDateRange(startDate, endDate, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"appointments": appointments,
		"total":        total,
		"page":         filters.Page,
		"limit":        filters.Limit,
		"total_pages":  (total + int64(filters.Limit) - 1) / int64(filters.Limit),
		"start_date":   startDate,
		"end_date":     endDate,
	})
}

// GetUpcoming handles getting upcoming appointments
func (h *AppointmentHandler) GetUpcoming(c *gin.Context) {
	// Parse limit parameter
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10 // Default to 10 if invalid
	}

	// Authorization check
	userObj, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user, ok := userObj.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user object"})
		return
	}

	// Get upcoming appointments
	appointments, err := h.appointmentService.GetUpcoming(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Filter appointments based on user role if not admin
	// In a real app, this filtering would be done at the service layer
	if user.Role != "admin" {
		// Just a placeholder for role-based filtering
	}

	c.JSON(http.StatusOK, gin.H{
		"appointments": appointments,
		"count":        len(appointments),
	})
}

// GetStatistics handles getting appointment statistics
func (h *AppointmentHandler) GetStatistics(c *gin.Context) {
	// Authorization check - only admins can view statistics
	userObj, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user, ok := userObj.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user object"})
		return
	}

	if user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only administrators can view statistics"})
		return
	}

	// Get appointment statistics
	statistics, err := h.appointmentService.GetStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"statistics": statistics})
}

// CheckAvailability handles checking if a time slot is available
func (h *AppointmentHandler) CheckAvailability(c *gin.Context) {
	var req CheckAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Validate time range
	if req.ScheduledStart.After(req.ScheduledEnd) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Start time must be before end time"})
		return
	}

	// Check if the time slot is at least 1 hour
	if req.ScheduledEnd.Sub(req.ScheduledStart) < time.Hour {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Appointment must be at least 1 hour long"})
		return
	}

	// Check if the start time is in the future
	if req.ScheduledStart.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Appointment must be scheduled for a future date"})
		return
	}

	// Check availability
	available, err := h.appointmentService.CheckAvailability(
		req.OperationID,
		req.EmployeeID,
		req.ScheduledStart,
		req.ScheduledEnd,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"available":       available,
		"scheduled_start": req.ScheduledStart,
		"scheduled_end":   req.ScheduledEnd,
		"operation_id":    req.OperationID,
		"employee_id":     req.EmployeeID,
	})
}

// hasStatusChangePermission checks if a user has permission to change an appointment to the requested status
func hasStatusChangePermission(user *models.User, appointment *models.Appointment, newStatus models.AppointmentStatus) bool {
    // Admins can change to any status
    if user.Role == "admin" {
        return true
    }

    // Check current status transitions
    switch appointment.Status {
    case models.StatusPending:
        // Pending can be confirmed by employee or cancelled by supplier
        if newStatus == models.StatusConfirmed && user.Role == "employee" {
            return true
        }
        if newStatus == models.StatusCancelled && user.Role == "supplier" {
            // Check if this is the supplier's appointment
            return user.ID == appointment.SupplierID
        }
    case models.StatusConfirmed:
        // Confirmed can be completed by employee or cancelled/rescheduled by supplier
        if newStatus == models.StatusCompleted && user.Role == "employee" {
            return user.ID == appointment.EmployeeID
        }
        if (newStatus == models.StatusCancelled || newStatus == models.StatusRescheduled) && user.Role == "supplier" {
            return user.ID == appointment.SupplierID
        }
    case models.StatusCancelled:
        // Cancelled appointments cannot transition to any other status
        return false
    case models.StatusCompleted:
        // Completed appointments cannot transition to any other status
        return false
    case models.StatusRescheduled:
        // Rescheduled appointments can only go back to pending
        if newStatus == models.StatusPending && user.Role == "supplier" {
            return user.ID == appointment.SupplierID
        }
    }

    return false
}
