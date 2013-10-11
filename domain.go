package main

import (
	"log"
	"time"
)

type Domain struct {
	Id        int        `json:"id"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func NewDomain(name string) *Domain {
	return &Domain{Name: name, Type: "NATIVE"}
}

func FindDomain(id int) (*Domain, error) {
	row := db.QueryRow("SELECT id, name, type FROM domains WHERE id = ?", id)

	d := &Domain{}
	err := row.Scan(
		&d.Id,
		&d.Name,
		&d.Type,
	)

	return d, err
}

func AllDomains() []*Domain {
	rows, err := db.Query("select id, name, type from domains")
	if err != nil {
		log.Fatal(err)
	}

	domains := []*Domain{}

	for rows.Next() {
		d := &Domain{}
		err := rows.Scan(
			&d.Id,
			&d.Name,
			&d.Type,
		)
		if err != nil {
			log.Fatal(err)
		}

		domains = append(domains, d)
	}

	return domains
}

func (d *Domain) Create() error {
	sql := "INSERT INTO domains (id, name, type) VALUES (?, ?, ?)"

	_, err := db.Exec(
		sql,
		d.Id,
		d.Name,
		d.Type,
	)
	return err
}

func (d *Domain) Update() error {
	sql := "UPDATE domains SET name=? WHERE id=?"

	_, err := db.Exec(sql, d.Name, d.Id)

	return err
}

func (d *Domain) Delete() error {
	sql := "DELETE FROM domains WHERE id = ?"
	recSql := "DELETE FROM records WHERE domain_id = ?"

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Delete records that belong to this domain and rollback on err
	_, err = tx.Exec(recSql, d.Id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Delete domain and rollback on err
	_, err = tx.Exec(sql, d.Id)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
