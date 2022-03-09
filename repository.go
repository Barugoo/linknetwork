package main

import "database/sql"

func InsertLink(db *sql.DB, userID int64, link string) error {
	_, err := db.Exec("INSERT INTO links (userID, link) VALUES (?, ?)", userID, link)
	return err
}

func GetLinkByUser(db *sql.DB, userID int64) (*string, error) {
	row := db.QueryRow("SELECT link FROM links WHERE userID = ?", userID)

	var res string
	if err := row.Scan(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func UpdateLinkByUser(db *sql.DB, userID int64, newLink string) error {
	_, err := db.Exec("UPDATE links SET link = ? WHERE userID = ?", newLink, userID)
	return err
}

func DeleteLinkByUser(db *sql.DB, userID int64) error {
	_, err := db.Exec("DELETE FROM links WHERE userID = ?", userID)
	return err
}

func GetLinkCount(db *sql.DB) (int64, error) {
	row := db.QueryRow("SELECT COUNT() FROM links")

	var res int64
	if err := row.Scan(&res); err != nil {
		return 0, err
	}
	return res, nil
}

func ListAllLinks(db *sql.DB) (map[string]string, error) {
	rows, err := db.Query("SELECT userID, link FROM links")
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)

	var userID, link string
	for rows.Next() {
		rows.Scan(&userID, &link)
		m[userID] = link
	}

	return m, nil
}

func LinkCount(db *sql.DB) (int64, error) {
	row := db.QueryRow("SELECT count() as count FROM links WHERE link is not '0'")

	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func InitDB(db *sql.DB) {
	db.Exec(`CREATE TABLE links(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		userID TEXT,
		link TEXT
	  );`)
}
