package models

import "time"

// Product represents the structure of a product.
type Product struct {
	//ProductID               *int      `json:"product_id"`
	ProductName             string    `json:"product_name"`
	ProductDescription      string    `json:"product_description"`
	ProductImages           []string  `json:"product_images"`
	ProductPrice            *int      `json:"product_price"`
	CompressedProductImages []string  `json:"compressed_product_images"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
	UserID                  *int      `json:"user_id"`
}

type User struct {
	ID        *int      `json:"id"`
	Name      string    `json:"name"`
	Mobile    string    `json:"mobile"`
	Latitude  *float64  `json:"latitude"`
	Longitude *float64  `json:"longitude"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Message represents the structure of a message that is send to MessageQueue
type Message struct {
	ProductID string  `json:"product_id"`
	Product   Product `json:"product"`
}
