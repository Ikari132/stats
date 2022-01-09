package main

type Log struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Count     int    `json:"count"`
	Product   string `json:"product"`
	CreatedAt int64  `json:"created_at" gorm:"autoCreateTime"`
}
type Product struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Name      string `json:"name"`
	Count     int    `json:"count"`
	CreatedAt int64  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt int64  `json:"updated_at" gorm:"autoUpdateTime"`
}
