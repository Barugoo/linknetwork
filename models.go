package main

import "time"

type Link struct {
	ID         int64     `json:"id"`
	URL        *string   `json:"url"`       // unique
	ShortURL   *string   `json:"short_url"` // unique
	UserID     int64     `json:"user_id"`   // unique
	ClickCount int64     `json:"click_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
