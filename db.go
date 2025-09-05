package main

import(
	"log"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func DB_Init() *sql.DB {
	var err error

	db, err = sql.Open("sqlite3", "site.sqlite3")
	if err != nil {
		log.Fatal("SQL:", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS contacts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL);`)

	if err != nil {
		log.Fatal("SQL:", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS count (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		count INTEGER);`)

	if err != nil {
		log.Fatal("SQL:", err)
	}

	DB_IncCount()

	return db
}

func DB_IncCount() int {
	row := db.QueryRow("UPDATE count SET count = count+1 WHERE id = 1 RETURNING count;")

	var count int

	err := row.Scan(&count)

	if err == sql.ErrNoRows {
		_, err = db.Exec("INSERT INTO count (count) VALUES (1);")
		count = 1
	}

	if err != nil {
		log.Fatal("SQL:", err)
	}

	return count
}

func DB_GetCount() int {
	row := db.QueryRow("SELECT count FROM count WHERE id = 1;")

	var count int

	err := row.Scan(&count)

	if err != nil {
		return 42069
	}
	return count
}

func DB_GetOneContact(id int) ContactEntry {
	var c ContactEntry
	row := db.QueryRow("SELECT * FROM contacts WHERE id = ?;", id)
	err := row.Scan(&c.Id, &c.Name, &c.Email)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func DB_GetAllContacts() []ContactEntry {
	var contacts []ContactEntry
	var con ContactEntry

	rows, err := db.Query("SELECT id, name, email FROM contacts;")
	defer rows.Close()

	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		err = rows.Scan(&con.Id, &con.Name, &con.Email)
		if err != nil {
			log.Fatal(err)
		}
		contacts = append(contacts, con)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	return contacts
}

func DB_ContactEmailExists(c ContactEntry) bool {
	row := db.QueryRow("SELECT email FROM contacts WHERE email = ?;", c.Email)

	var temp string
	err := row.Scan(&temp)

	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		log.Fatal(row.Err())
	}
	return true
}

func DB_AddContact(c ContactEntry) {
	_, err := db.Exec("INSERT INTO contacts (name, email) VALUES (?, ?)", c.Name, c.Email)
	if err != nil {
		log.Fatal(err)
	}
}

func DB_DeleteContact(id int) {
		_, err := db.Exec("DELETE FROM contacts WHERE id = ?", id)
	if err != nil {
		log.Fatal(err)
	}
}

func DB_UpdateContact(c ContactEntry) {
	_, err := db.Exec("UPDATE contacts SET name = ?, email = ? WHERE id = ?", c.Name, c.Email, c.Id)
		if err != nil {
		log.Fatal(err)
	}
}
