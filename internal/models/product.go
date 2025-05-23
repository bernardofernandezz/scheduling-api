package models

import (
	"time"
)

// Product represents a product that can be delivered
type Product struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description" gorm:"type:text"`
	SKU         string    `json:"sku" gorm:"uniqueIndex"`
	Category    string    `json:"category"`
	Price       float64   `json:"price" gorm:"type:decimal(10,2)"`
	SupplierID  uint      `json:"supplier_id" gorm:"not null"`
	Supplier    Supplier  `json:"supplier" gorm:"foreignKey:SupplierID"`
	Active      bool      `json:"active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

