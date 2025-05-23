package models

import (
    "time"
    "errors"
)

// Operation represents a company location or branch
type Operation struct {
    ID              uint      `json:"id" gorm:"primaryKey"`
    Name            string    `json:"name" gorm:"not null"`
    Code            string    `json:"code" gorm:"uniqueIndex;not null"`
    Address         string    `json:"address" gorm:"not null"`
    City            string    `json:"city" gorm:"not null"`
    State           string    `json:"state" gorm:"not null"`
    ZipCode         string    `json:"zip_code" gorm:"not null"`
    Country         string    `json:"country" gorm:"not null;default:'Brazil'"`
    Phone           string    `json:"phone"`
    Email           string    `json:"email"`
    ManagerID       uint      `json:"manager_id" gorm:"not null"`
    Manager         Employee  `json:"manager" gorm:"foreignKey:ManagerID"`
    OpeningTime     string    `json:"opening_time" gorm:"not null;default:'08:00'"`
    ClosingTime     string    `json:"closing_time" gorm:"not null;default:'18:00'"`
    Active          bool      `json:"active" gorm:"default:true"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

// Validate performs validation on the operation
func (o *Operation) Validate() error {
    if o.Name == "" {
        return errors.New("name is required")
    }
    if o.Code == "" {
        return errors.New("code is required")
    }
    if o.Address == "" {
        return errors.New("address is required")
    }
    if o.City == "" {
        return errors.New("city is required")
    }
    if o.State == "" {
        return errors.New("state is required")
    }
    if o.ZipCode == "" {
        return errors.New("zip code is required")
    }
    if o.ManagerID == 0 {
        return errors.New("manager is required")
    }
    return nil
}

// BeforeCreate is called by GORM before creating a new record
func (o *Operation) BeforeCreate() error {
    return o.Validate()
}

// BeforeUpdate is called by GORM before updating a record
func (o *Operation) BeforeUpdate() error {
    return o.Validate()
}

