package main

import (
	"database/sql"
	"time"
)

type Repository interface {
	CreateLink(link *Link) error
	GetLinkByUserID(userID int64) (*Link, error)
	GetLinkByShortURL(shortUR string) (*Link, error)
	UpdateLink(link *Link) error
	DeleteLinkByUserID(userID int64) error
	GetLinkCount() (int64, error)
	ListAllLinks(limit int) ([]*Link, error) // userID -> link
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db}
}

func (rep repository) CreateLink(link *Link) error {
	now := time.Now()
	_, err := rep.db.Exec("INSERT INTO links (user_id, url, short_url, click_count, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		link.UserID,
		link.URL,
		link.ShortURL,
		0,
		now,
		now)
	return err
}

func (rep repository) GetLinkByUserID(userID int64) (*Link, error) {
	row := rep.db.QueryRow("SELECT id, user_id, url, short_url, click_count, created_at, updated_at FROM links WHERE user_id = ?", userID)

	var res Link
	if err := row.Scan(
		&res.ID,
		&res.UserID,
		&res.URL,
		&res.ShortURL,
		&res.ClickCount,
		&res.CreatedAt,
		&res.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &res, nil
}

func (rep repository) GetLinkByShortURL(shortURL string) (*Link, error) {
	row := rep.db.QueryRow("SELECT id, user_id, url, short_url, click_count, created_at, updated_at FROM links WHERE short_url = ?", shortURL)

	var res Link
	if err := row.Scan(
		&res.ID,
		&res.UserID,
		&res.URL,
		&res.ShortURL,
		&res.ClickCount,
		&res.CreatedAt,
		&res.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &res, nil
}

func (rep repository) UpdateLink(link *Link) error {
	_, err := rep.db.Exec("UPDATE links SET user_id = ?, url = ?, short_url = ?, click_count = ?, updated_at = ? WHERE id = ?",
		link.UserID,
		link.URL,
		link.ShortURL,
		link.ClickCount,
		time.Now(),
	)
	return err
}

func (rep repository) DeleteLinkByUserID(userID int64) error {
	_, err := rep.db.Exec("DELETE FROM links WHERE user_id = ?", userID)
	return err
}

func (rep repository) ListAllLinks(limit int) ([]*Link, error) {
	rows, err := rep.db.Query("SELECT id, user_id, url, short_url, click_count, created_at, updated_at FROM links WHERE url IS NOT NULL ORDER BY ASC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}

	m := make([]*Link, 0)

	for rows.Next() {
		var res Link
		rows.Scan(
			&res.ID,
			&res.UserID,
			&res.URL,
			&res.ShortURL,
			&res.ClickCount,
			&res.CreatedAt,
			&res.UpdatedAt,
		)
		m = append(m, &res)
	}
	return m, nil
}

func (rep repository) GetLinkCount() (int64, error) {
	row := rep.db.QueryRow("SELECT count() as count FROM links WHERE link is not '0'")

	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
