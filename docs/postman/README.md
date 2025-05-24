# Scheduling API - Postman Testing Guide

This guide will help you test the Scheduling API using Postman. It includes environment setup, authentication flows, and examples for all major API endpoints.

## Table of Contents

- [Environment Setup](#environment-setup)
- [Authentication](#authentication)
- [Appointments](#appointments)
- [Recurring Appointments](#recurring-appointments)
- [Notifications](#notifications)
- [Calendar Integration](#calendar-integration)

## Environment Setup

### 1. Import the Collection

1. Download the [Scheduling API Postman Collection](./scheduling-api-collection.json)
2. In Postman, click on "Import" and select the downloaded file
3. The collection will be imported with all request templates

### 2. Set Up Environment Variables

Create a new environment in Postman with the following variables:

| Variable | Initial Value | Description |
|----------|--------------|-------------|
| `base_url` | `http://localhost:8080` | The base URL of your API |
| `token` | (empty) | Will store the JWT token after login |
| `refresh_token` | (empty) | Will store the refresh token after login |
| `user_id` | (empty) | Will store the current user ID |

### 3. Configure the API Server

Make sure your API server is running. By default, it runs on port 8080.

## Authentication

### Register a New User

**Request:**
```
POST {{base_url}}/api/auth/register
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "password123",
  "name": "Test User",
  "role": "supplier"
}
```

**Response:**
```json
{
  "user": {
    "id": 1,
    "email": "test@example.com",
    "name": "Test User",
    "role": "supplier",
    "created_at": "2025-05-24T00:00:00Z"
  }
}
```

### Login

**Request:**
```
POST {{base_url}}/api/auth/login
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "test@example.com",
    "name": "Test User",
    "role": "supplier"
  }
}
```

After receiving the response, the collection will automatically set the `token`, `refresh_token`, and `user_id` environment variables.

### Refresh Token

**Request:**
```
POST {{base_url}}/api/auth/refresh
Content-Type: application/json

{
  "refresh_token": "{{refresh_token}}"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Request Password Reset

**Request:**
```
POST {{base_url}}/api/auth/password-reset
Content-Type: application/json

{
  "email": "test@example.com"
}
```

**Response:**
```json
{
  "message": "Password reset instructions sent to your email"
}
```

## Appointments

All appointment endpoints require authentication. The JWT token is automatically included in the request headers.

### Create an Appointment

**Request:**
```
POST {{base_url}}/api/appointments
Content-Type: application/json
Authorization: Bearer {{token}}

{
  "supplier_id": 1,
  "employee_id": 2,
  "operation_id": 3,
  "product_id": 4,
  "scheduled_start": "2025-06-01T10:00:00Z",
  "scheduled_end": "2025-06-01T11:00:00Z",
  "notes": "Test appointment",
  "quantity_to_deliver": 100
}
```

**Response:**
```json
{
  "appointment": {
    "id": 1,
    "supplier_id": 1,
    "employee_id": 2,
    "operation_id": 3,
    "product_id": 4,
    "scheduled_start": "2025-06-01T10:00:00Z",
    "scheduled_end": "2025-06-01T11:00:00Z",
    "notes": "Test appointment",
    "quantity_to_deliver": 100,
    "status": "pending",
    "created_at": "2025-05-24T00:00:00Z",
    "updated_at": "2025-05-24T00:00:00Z"
  }
}
```

### Get Appointment by ID

**Request:**
```
GET {{base_url}}/api/appointments/1
Authorization: Bearer {{token}}
```

**Response:**
```json
{
  "appointment": {
    "id": 1,
    "supplier_id": 1,
    "employee_id": 2,
    "operation_id": 3,
    "product_id": 4,
    "scheduled_start": "2025-06-01T10:00:00Z",
    "scheduled_end": "2025-06-01T11:00:00Z",
    "notes": "Test appointment",
    "quantity_to_deliver": 100,
    "status": "pending",
    "created_at": "2025-05-24T00:00:00Z",
    "updated_at": "2025-05-24T00:00:00Z"
  }
}
```

### Update an Appointment

**Request:**
```
PUT {{base_url}}/api/appointments/1
Content-Type: application/json
Authorization: Bearer {{token}}

{
  "notes": "Updated test appointment",
  "quantity_to_deliver": 150
}
```

**Response:**
```json
{
  "appointment": {
    "id": 1,
    "supplier_id": 1,
    "employee_id": 2,
    "operation_id": 3,
    "product_id": 4,
    "scheduled_start": "2025-06-01T10:00:00Z",
    "scheduled_end": "2025-06-01T11:00:00Z",
    "notes": "Updated test appointment",
    "quantity_to_deliver": 150,
    "status": "pending",
    "created_at": "2025-05-24T00:00:00Z",
    "updated_at": "2025-05-24T00:00:00Z"
  }
}
```

### Update Appointment Status

**Request:**
```
POST {{base_url}}/api/appointments/1/status
Content-Type: application/json
Authorization: Bearer {{token}}

{
  "status": "confirmed",
  "reason": "Confirmed by employee"
}
```

**Response:**
```json
{
  "appointment": {
    "id": 1,
    "supplier_id": 1,
    "employee_id": 2,
    "operation_id": 3,
    "product_id": 4,
    "scheduled_start": "2025-06-01T10:00:00Z",
    "scheduled_end": "2025-06-01T11:00:00Z",
    "notes": "Updated test appointment",
    "quantity_to_deliver": 150,
    "status": "confirmed",
    "created_at": "2025-05-24T00:00:00Z",
    "updated_at": "2025-05-24T00:00:00Z"
  }
}
```

### List Appointments

**Request:**
```
GET {{base_url}}/api/appointments?page=1&limit=10
Authorization: Bearer {{token}}
```

**Response:**
```json
{
  "appointments": [
    {
      "id": 1,
      "supplier_id": 1,
      "employee_id": 2,
      "operation_id": 3,
      "product_id": 4,
      "scheduled_start": "2025-06-01T10:00:00Z",
      "scheduled_end": "2025-06-01T11:00:00Z",
      "notes": "Updated test appointment",
      "quantity_to_deliver": 150,
      "status": "confirmed",
      "created_at": "2025-05-24T00:00:00Z",
      "updated_at": "2025-05-24T00:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10,
  "total_pages": 1
}
```

### Delete an Appointment

**Request:**
```
DELETE {{base_url}}/api/appointments/1
Authorization: Bearer {{token}}
```

**Response:**
```json
{
  "message": "Appointment deleted successfully"
}
```

### Check Appointment Availability

**Request:**
```
POST {{base_url}}/api/appointments/check-availability
Content-Type: application/json
Authorization: Bearer {{token}}

{
  "operation_id": 3,
  "employee_id": 2,
  "scheduled_start": "2025-06-02T10:00:00Z",
  "scheduled_end": "2025-06-02T11:00:00Z"
}
```

**Response:**
```json
{
  "available": true,
  "scheduled_start": "2025-06-02T10:00:00Z",
  "scheduled_end": "2025-06-02T11:00:00Z",
  "operation_id": 3,
  "employee_id": 2
}
```

### Get Appointments by Date Range

**Request:**
```
GET {{base_url}}/api/appointments/by-date-range?start_date=2025-06-01T00:00:00Z&end_date=2025-06-30T23:59:59Z
Authorization: Bearer {{token}}
```

**Response:**
```json
{
  "appointments": [
    {
      "id": 1,
      "supplier_id": 1,
      "employee_id": 2,
      "operation_id": 3,
      "product_id": 4,
      "scheduled_start": "2025-06-01T10:00:00Z",
      "scheduled_end": "2025-06-01T11:00:00Z",
      "notes": "Updated test appointment",
      "quantity_to_deliver": 150,
      "status": "confirmed",
      "created_at": "2025-05-24T00:00:00Z",
      "updated_at": "2025-05-24T00:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10,
  "total_pages": 1,
  "start_date": "2025-06-01T00:00:00Z",
  "end_date": "2025-06-30T23:59:59Z"
}
```

## Recurring Appointments

### Create a Recurring Appointment

**Request:**
```
POST {{base_url}}/api/recurring-appointments
Content-Type: application/json
Authorization: Bearer {{token}}

{
  "supplier_id": 1,
  "employee_id": 2,
  "operation_id": 3,
  "product_id": 4,
  "quantity_to_deliver": 100,
  "notes": "Weekly delivery",
  "pattern": "weekly",
  "start_date": "2025-06-01",
  "end_date": "2025-08-31",
  "start_time_minutes": 600,
  "duration_minutes": 60,
  "week_days": [1, 3, 5],
  "exclusion_dates": ["2025-07-04"]
}
```

**Response:**
```json
{
  "recurring_appointment": {
    "id": 1,
    "supplier_id": 1,
    "employee_id": 2,
    "operation_id": 3,
    "product_id": 4,
    "quantity_to_deliver": 100,
    "notes": "Weekly delivery",
    "pattern": "weekly",
    "start_date": "2025-06-01T00:00:00Z",
    "end_date": "2025-08-31T00:00:00Z",
    "start_time_minutes": 600,
    "duration_minutes": 60,
    "week_days": [1, 3, 5],
    "exclusion_dates": ["2025-07-04T00:00:00Z"],
    "created_at": "2025-05-24T00:00:00Z",
    "updated_at": "2025-05-24T00:00:00Z"
  }
}
```

### Get Recurring Appointment by ID

**Request:**
```
GET {{base_url}}/api/recurring-appointments/1
Authorization: Bearer {{token}}
```

**Response:**
```json
{
  "recurring_appointment": {
    "id": 1,
    "supplier_id": 1,
    "employee_id": 2,
    "operation_id": 3,
    "product_id": 4,
    "quantity_to_deliver": 100,
    "notes": "Weekly delivery",
    "pattern": "weekly",
    "start_date": "2025-06-01T00:00:00Z",
    "end_date": "2025-08-31T00:00:00Z",
    "start_time_minutes": 600,
    "duration_minutes": 60,
    "week_days": [1, 3, 5],
    "exclusion_dates": ["2025-07-04T00:00:00Z"],
    "created_at": "2025-05-24T00:00:00Z",
    "updated_at": "2025-05-24T00:00:00Z"
  }
}
```

### Generate Occurrences for a Recurring Appointment

**Request:**
```
GET {{base_url}}/api/recurring-appointments/1/occurrences
Authorization: Bearer {{token}}
```

**Response:**
```json
{
  "occurrences": [
    "2025-06-02T10:00:00Z",
    "2025-06-04T10:00:00Z",
    "2025-06-06T10:00:00Z",
    "2025-06-09T10:00:00Z",
    "2025-06-11T10:00:00Z",
    "2025-06-13T10:00:00Z"
  ],
  "total": 26
}
```

## Notifications

### Get User Notification Preferences

**Request:**
```
GET {{base_url}}/api/notifications/preferences
Authorization: Bearer {{token}}
```

**Response:**
```json
{
  "preferences": {
    "email_enabled": true,
    "sms_enabled": false,
    "push_enabled": false,
    "event_prefs": {
      "appointment_created": true,
      "appointment_updated": true,
      "appointment_cancelled": true,
      "appointment_confirmed": true,
      "appointment_completed": true,
      "appointment_reminder": true
    },
    "reminder_hours": 24,
    "email": "test@example.com",
    "phone_number": ""
  }
}
```

### Update User Notification Preferences

**Request:**
```
PUT {{base_url}}/api/notifications/preferences
Content-Type: application/json
Authorization: Bearer {{token}}

{
  "email_enabled": true,
  "sms_enabled": true,
  "push_enabled": false,
  "event_prefs": {
    "appointment_created": true,
    "appointment_updated": true,
    "appointment_cancelled": true,
    "appointment_confirmed": true,
    "appointment_completed": true,
    "appointment_reminder": true
  },
  "reminder_hours": 12,
  "phone_number": "+1234567890"
}
```

**Response:**
```json
{
  "preferences": {
    "email_enabled": true,
    "sms_enabled": true,
    "push_enabled": false,
    "event_prefs": {
      "appointment_created": true,
      "appointment_updated": true,
      "appointment_cancelled": true,
      "appointment_confirmed": true,
      "appointment_completed": true,
      "appointment_reminder": true
    },
    "reminder_hours": 12,
    "email": "test@example.com",
    "phone_number": "+1234567890"
  }
}
```

### Get User Notifications

**Request:**
```
GET {{base_url}}/api/notifications?page=1&limit=10
Authorization: Bearer {{token}}
```

**Response:**
```json
{
  "notifications": [
    {
      "id": 1,
      "type": "email",
      "status": "sent",
      "event": "appointment_created",
      "subject": "New Appointment Created",
      "sent_at": "2025-05-24T00:00:00Z",
      "appointment_id": 1
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10,
  "total_pages": 1
}
```

## Calendar Integration

### Get User Calendar Preferences

**Request:**
```
GET {{base_url}}/api/calendar/preferences
Authorization: Bearer {{token}}
```

**Response:**
```json
{
  "preferences": {
    "google_enabled": false,
    "outlook_enabled": false,
    "ical_enabled": true,
    "google_calendar_id": "",
    "google_access_token": ""
  }
}
```

### Update User Calendar Preferences

**Request:**
```
PUT {{base_url}}/api/calendar/preferences
Content-Type: application/json
Authorization: Bearer {{token}}

{
  "google_enabled": true,
  "outlook_enabled": false,
  "ical_enabled": true,
  "google_calendar_id": "primary"
}
```

**Response:**
```json
{
  "preferences": {
    "google_enabled": true,
    "outlook_enabled": false,
    "ical_enabled": true,
    "google_calendar_id": "primary"
  }
}
```

### Generate iCalendar for an Appointment

**Request:**
```
GET {{base_url}}/api/calendar/ical/appointments/1
Authorization: Bearer {{token}}
```

**Response:**
```
BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Scheduling API//Delivery Appointment//EN
CALSCALE:GREGORIAN
METHOD:PUBLISH
BEGIN:VEVENT
UID:appointment-1@scheduling-api.example.com
DTSTAMP:20250524T000000Z
DTSTART:20250601T100000Z
DTEND:20250601T110000Z
SUMMARY:Delivery from Supplier Name
DESCRIPTION:Supplier: Supplier Name\nEmployee: Employee Name\nOperation: Operation Name\nProduct: Product Name\nQuantity: 150\nStatus: confirmed
LOCATION:Operation Name
STATUS:CONFIRMED
URL:https://scheduling-api.example.com/appointments/1
ORGANIZER;CN=Employee Name:mailto:noreply@example.com
END:VEVENT
END:VCALENDAR
```

### Generate Calendar Links for an Appointment

**Request:**
```
GET {{base_url}}/api/calendar/links/appointments/1
Authorization: Bearer {{token}}
```

**Response:**
```json
{
  "google_calendar_link": "https://calendar.google.com/calendar/render?action=TEMPLATE&text=Delivery+from+Supplier+Name&dates=20250601T100000Z/20250601T110000Z&details=Supplier:+Supplier+Name%0AEmployee:+Employee+Name%0AOperation:+Operation+Name%0AProduct:+Product+Name%0AQuantity:+150&location=Operation+Name",
  "outlook_calendar_link": "https://outlook.office.com/calendar/0/deeplink/compose?subject=Delivery+from+Supplier+Name&startdt=2025-06-01T10:00:00Z&enddt=2025-06-01T11:00:00Z&body=Supplier:+Supplier+Name%0AEmployee:+Employee+Name%0AOperation:+Operation+Name%0AProduct:+Product+Name%0AQuantity:+150&location=Operation+Name",
  "ical_download_link": "{{base_url}}/api/calendar/ical/appointments/1"
}
```

### Sync Appointment to External Calendar

**Request:**
```
POST {{base_url}}/api/calendar/sync
Content-Type: application/json
Authorization: Bearer {{token}}

{
  "appointment_id": 1,
  "provider": "google"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Appointment synced to Google Calendar",
  "external_event_id": "abc123xyz456"
}
```

## Testing the API Health

### Health Check

**Request:**
```
GET {{base_url}}/health
```

**Response:**
```json
{
  "status": "UP",
  "time": "2025-05-24T00:00:00Z",
  "mode": "debug",
  "version": "1.0.0"
}
```

### Readiness Check

**Request:**
```
GET {{base_url}}/ready
```

**Response:**
```json
{
  "status": "UP",
  "database": "connected"
}
```

## Using the Postman Collection

1. **Authentication Flow**: Start by registering a user, then logging in to get the JWT token.
2. **Create Test Data**: Create suppliers, employees, operations, and products before creating appointments.
3. **Test Appointments**: Create, update, and manage appointments.
4. **Test Recurring Appointments**: Create recurring patterns and view occurrences.
5. **Test Notifications**: Update notification preferences and check notifications.
6. **Test Calendar Integration**: Generate calendar links and sync appointments.

## Troubleshooting

### Common Issues

1. **401 Unauthorized**: Check if your token is valid and not expired.
2. **403 Forbidden**: Check if you have the necessary permissions for the operation.
3. **400 Bad Request**: Check your request payload for missing or invalid fields.
4. **500 Internal Server Error**: Check the server logs for details.

### Refreshing Tokens

If you receive a 401 Unauthorized error with a message indicating that your token has expired, use the refresh token endpoint to obtain a new token:

```
POST {{base_url}}/api/auth/refresh
Content-Type: application/json

{
  "refresh_token": "{{refresh_token}}"
}
```

This will update your token environment variable automatically.

