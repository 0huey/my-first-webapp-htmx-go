package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/argon2"
)

const (
	SQL_TRUE  int = 1
	SQL_FALSE int = 0
)

var DB_ErrorUnknownUserOrPass = errors.New("Unknown username or password")
var DB_ErrorAccountLocked = errors.New("Account locked")
var DB_ErrorLoginSessionExpired = errors.New("Login session expired")

var db *sql.DB

func DB_Init() *sql.DB {
	var err error

	db, err = sql.Open("sqlite3", "site.sqlite3")
	if err != nil {
		log.Panic("SQL:", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS contacts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL);`)

	if err != nil {
		log.Panic("SQL:", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS count (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		count INTEGER NOT NULL);`)

	if err != nil {
		log.Panic("SQL:", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_auth (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		salt TEXT NOT NULL,
		hash TEXT NOT NULL,
		enabled INTEGER NOT NULL);`)

	if err != nil {
		log.Panic("SQL:", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_login_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL,
		expires INT NOT NULL);`)

	if err != nil {
		log.Panic("SQL:", err)
	}

	row := db.QueryRow("SELECT username FROM user_auth WHERE username = 'admin';")
	var name string
	err = row.Scan(&name)
	if err == sql.ErrNoRows {
		DB_UserRegister("admin", "admin")
	} else if err != nil {
		log.Panic(err)
	}

	DB_IncCount()

	return db
}

func wildcardTerm(term string) string {
	return "%" + term + "%"
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
		log.Panic(err)
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
	if err == sql.ErrNoRows {
		return c
	} else if err != nil {
		log.Panic(err)
	}
	return c
}

func DB_GetAllContacts() []ContactEntry {
	var contacts []ContactEntry
	var con ContactEntry

	rows, err := db.Query("SELECT id, name, email FROM contacts ORDER BY LOWER(name);")
	defer rows.Close()

	if err != nil {
		log.Panic(err)
	}

	for rows.Next() {
		err = rows.Scan(&con.Id, &con.Name, &con.Email)
		if err != nil {
			log.Panic(err)
		}
		contacts = append(contacts, con)
	}

	if err = rows.Err(); err != nil {
		log.Panic(err)
	}

	return contacts
}

func DB_GetAllContactsSearch(term string) []ContactEntry {
	var contacts []ContactEntry
	var con ContactEntry

	term = wildcardTerm(term)

	rows, err := db.Query(`SELECT id, name, email FROM contacts
		where name LIKE ? OR email LIKE ?
		ORDER BY LOWER(name);`, term, term)
	defer rows.Close()

	if err != nil {
		log.Panic(err)
	}

	for rows.Next() {
		err = rows.Scan(&con.Id, &con.Name, &con.Email)
		if err != nil {
			log.Panic(err)
		}
		contacts = append(contacts, con)
	}

	if err = rows.Err(); err != nil {
		log.Panic(err)
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
		log.Panic(err)
	}
	return true
}

func DB_AddContact(c ContactEntry) ContactEntry {
	row := db.QueryRow("INSERT INTO contacts (name, email) VALUES (?, ?) RETURNING id;", c.Name, c.Email)
	var id int
	err := row.Scan(&id)
	if err != nil {
		log.Panic(err)
	}
	c.Id = id
	return c
}

func DB_DeleteContact(id int) {
	_, err := db.Exec("DELETE FROM contacts WHERE id = ?;", id)
	if err != nil {
		log.Panic(err)
	}
}

func DB_UpdateContact(c ContactEntry) {
	_, err := db.Exec("UPDATE contacts SET name = ?, email = ? WHERE id = ?;", c.Name, c.Email, c.Id)
	if err != nil {
		log.Panic(err)
	}
}

func DB_UserRegister(name string, password string) {
	salt := rand.Text()
	hash := hashPassword(password, salt)

	_, err := db.Exec("INSERT INTO user_auth (username, salt, hash, enabled) VALUES (?, ?, ?, ?);",
		name, salt, hash, SQL_TRUE)

	if err != nil {
		log.Panic(err)
	}
}

func DB_UserGetNameOfId(id int) (string, error) {
	var username string
	row := db.QueryRow("SELECT username FROM user_auth WHERE id = ?", id)

	err := row.Scan(&username)

	if err == sql.ErrNoRows {
		return username, DB_ErrorUnknownUserOrPass
	} else if err != nil {
		log.Panic(err)
	}

	return username, nil
}

func DB_UserLogin(name string, password string) (LoginSession, error) {
	var user_id int
	var salt string
	var hash string
	var enabled int
	var token LoginSession

	row := db.QueryRow("SELECT id, salt, hash, enabled FROM user_auth WHERE username = ?;", name)

	err := row.Scan(&user_id, &salt, &hash, &enabled)

	if err == sql.ErrNoRows {
		return token, DB_ErrorUnknownUserOrPass
	} else if err != nil {
		log.Panic(err)
	}

	hash2 := hashPassword(password, salt)

	if strings.Compare(hash, hash2) != 0 {
		return token, DB_ErrorUnknownUserOrPass
	}

	if enabled == SQL_FALSE {
		return token, DB_ErrorAccountLocked
	}

	token.Username = name
	token.Token = rand.Text() + rand.Text()
	token.Expires = time.Now().UTC().Add(time.Hour * 12)

	_, err = db.Exec("INSERT INTO user_login_tokens (user_id, token, expires) VALUES (?, ?, ?);",
		user_id, token.Token, token.Expires.Unix())

	if err != nil {
		log.Panic(err)
	}

	return token, nil
}

func DB_UserLookupLoginToken(token string) (string, error) {
	var user_id int
	var expires_unix int

	row := db.QueryRow("SELECT user_id, expires FROM user_login_tokens WHERE token = ?;", token)

	err := row.Scan(&user_id, &expires_unix)

	if err == sql.ErrNoRows {
		return "", DB_ErrorLoginSessionExpired

	} else if err != nil {
		log.Panic(err)
	}

	// add expiration

	return DB_UserGetNameOfId(user_id)
}

func hashPassword(pass string, salt string) string {
	digest := argon2.IDKey([]byte(pass), []byte(salt), 8, 64*1024, 4, 32)
	return hex.EncodeToString(digest)
}
