package main

import (
	"log"
	"time"
)

type Record struct {
	Id        int        `json:"id"`
	DomainId  int        `json:"domain_id"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	Content   string     `json:"content"`
	Ttl       int        `json:"ttl"`
	Priority  *int       `json:"prio"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func FindRecord(id int) (*Record, error) {
	row := db.QueryRow("SELECT id, domain_id, name, type, content, ttl, prio FROM records WHERE id = ?", id)

	r := &Record{}
	err := row.Scan(
		&r.Id,
		&r.DomainId,
		&r.Name,
		&r.Type,
		&r.Content,
		&r.Ttl,
		&r.Priority,
	)

	return r, err
}

func AllRecords() []*Record {
	rows, err := db.Query("SELECT id, domain_id, name, type, content, ttl, prio FROM records")
	if err != nil {
		log.Fatal(err)
	}

	records := []*Record{}

	for rows.Next() {
		r := &Record{}
		err := rows.Scan(
			&r.Id,
			&r.DomainId,
			&r.Name,
			&r.Type,
			&r.Content,
			&r.Ttl,
			&r.Priority,
		)
		if err != nil {
			log.Fatal(err)
		}

		records = append(records, r)
	}

	return records
}

func (r *Record) Create() error {
	sql := "INSERT INTO records (id, domain_id, name, type, content, ttl, prio) VALUES (?, ?, ?, ?, ?, ?, ?)"

	_, err := db.Exec(
		sql,
		r.Id,
		r.DomainId,
		r.Name,
		r.Type,
		r.Content,
		r.Ttl,
		r.Priority,
	)
	return err
}

func (r *Record) Update() error {
	sql := "UPDATE records SET name=?, type=?, content=?, ttl=?, prio=? WHERE id=?"

	_, err := db.Exec(sql, r.Name, r.Type, r.Content, r.Ttl, r.Priority, r.Id)

	return err
}

func (r *Record) Delete() error {
	sql := "DELETE FROM records WHERE id = ?"

	_, err := db.Exec(sql, r.Id)

	return err
}
