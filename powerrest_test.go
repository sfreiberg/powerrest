package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
)

// setupDB replaces the global db with a sqlmock for the duration of a test.
func setupDB(t *testing.T) sqlmock.Sqlmock {
	t.Helper()
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	old := db
	db = mockDB
	t.Cleanup(func() {
		db = old
		mockDB.Close()
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
	return mock
}

func call(h http.HandlerFunc, method, target, body string, vars map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if vars != nil {
		req = mux.SetURLVars(req, vars)
	}
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func TestLoadConfigFile(t *testing.T) {
	dir := t.TempDir()

	full := filepath.Join(dir, "full.conf")
	os.WriteFile(full, []byte("db-type = postgres\ndb-connection = root@/dns\nlisten-address = :9090\n"), 0o600)
	c, err := loadConfigFile(full)
	if err != nil {
		t.Fatal(err)
	}
	if c.DbType != "postgres" || c.DbConn != "root@/dns" || c.ListenAddr != ":9090" {
		t.Errorf("config = %+v", c)
	}

	empty := filepath.Join(dir, "empty.conf")
	os.WriteFile(empty, nil, 0o600)
	c, err = loadConfigFile(empty)
	if err != nil {
		t.Fatal(err)
	}
	if c.DbType != "mysql" || c.DbConn != "" || c.ListenAddr != ":8080" {
		t.Errorf("defaults = %+v", c)
	}

	if _, err := loadConfigFile(filepath.Join(dir, "missing.conf")); err == nil {
		t.Error("expected an error for a missing file")
	}
}

func TestDomainIndex(t *testing.T) {
	mock := setupDB(t)
	mock.ExpectQuery("select id, name, type from domains").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "type"}).
			AddRow(1, "example.com", "NATIVE").
			AddRow(2, "example.org", "MASTER"))

	rec := call(domainIndex, "GET", "/v1/domains", "", nil)
	if rec.Code != 200 {
		t.Fatalf("code = %d", rec.Code)
	}
	var got []Domain
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[1].Name != "example.org" || got[1].Type != "MASTER" {
		t.Fatalf("domains = %+v", got)
	}
}

func TestDomainShow(t *testing.T) {
	mock := setupDB(t)
	q := "SELECT id, name, type FROM domains WHERE id = ?"
	mock.ExpectQuery(q).WithArgs(7).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "type"}).AddRow(7, "example.com", "NATIVE"))
	mock.ExpectQuery(q).WithArgs(8).WillReturnError(sql.ErrNoRows)

	rec := call(domainShow, "GET", "/v1/domains/7", "", map[string]string{"id": "7"})
	if rec.Code != 200 {
		t.Fatalf("code = %d", rec.Code)
	}
	var d Domain
	if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	if d.Id != 7 || d.Name != "example.com" {
		t.Fatalf("domain = %+v", d)
	}

	if rec := call(domainShow, "GET", "/v1/domains/8", "", map[string]string{"id": "8"}); rec.Code != 404 {
		t.Fatalf("missing domain: code = %d, want 404", rec.Code)
	}
}

func TestDomainCreate(t *testing.T) {
	mock := setupDB(t)
	mock.ExpectExec("INSERT INTO domains (id, name, type) VALUES (?, ?, ?)").
		WithArgs(0, "example.com", "NATIVE").
		WillReturnResult(sqlmock.NewResult(1, 1))

	rec := call(domainCreate, "POST", "/v1/domains", `{"name":"example.com"}`, nil)
	if rec.Code != 200 {
		t.Fatalf("code = %d, body = %s", rec.Code, rec.Body)
	}
}

func TestDomainUpdate(t *testing.T) {
	mock := setupDB(t)
	mock.ExpectQuery("SELECT id, name, type FROM domains WHERE id = ?").WithArgs(3).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "type"}).AddRow(3, "old.com", "NATIVE"))
	mock.ExpectExec("UPDATE domains SET name=? WHERE id=?").WithArgs("new.com", 3).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rec := call(domainUpdate, "POST", "/v1/domains/3", `{"name":"new.com"}`, map[string]string{"id": "3"})
	if rec.Code != 200 {
		t.Fatalf("code = %d, body = %s", rec.Code, rec.Body)
	}
}

func TestDomainDeleteRemovesRecords(t *testing.T) {
	mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM records WHERE domain_id = ?").WithArgs(5).WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec("DELETE FROM domains WHERE id = ?").WithArgs(5).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	rec := call(domainDelete, "DELETE", "/v1/domains/5", "", map[string]string{"id": "5"})
	if rec.Code != 200 {
		t.Fatalf("code = %d, body = %s", rec.Code, rec.Body)
	}
}

func TestDomainDeleteRollsBack(t *testing.T) {
	mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM records WHERE domain_id = ?").WithArgs(5).WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec("DELETE FROM domains WHERE id = ?").WithArgs(5).WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	rec := call(domainDelete, "DELETE", "/v1/domains/5", "", map[string]string{"id": "5"})
	if rec.Code != 500 {
		t.Fatalf("code = %d, want 500", rec.Code)
	}
}

func TestRecordCRUD(t *testing.T) {
	mock := setupDB(t)
	cols := []string{"id", "domain_id", "name", "type", "content", "ttl", "prio"}
	mock.ExpectQuery("SELECT id, domain_id, name, type, content, ttl, prio FROM records WHERE id = ?").WithArgs(9).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(9, 1, "www.example.com", "A", "192.0.2.1", 3600, nil))
	mock.ExpectExec("INSERT INTO records (id, domain_id, name, type, content, ttl, prio) VALUES (?, ?, ?, ?, ?, ?, ?)").
		WithArgs(0, 1, "example.com", "MX", "mail.example.com", 300, 10).
		WillReturnResult(sqlmock.NewResult(10, 1))
	mock.ExpectExec("DELETE FROM records WHERE id = ?").WithArgs(9).WillReturnResult(sqlmock.NewResult(0, 1))

	rec := call(recordShow, "GET", "/v1/records/9", "", map[string]string{"id": "9"})
	if rec.Code != 200 {
		t.Fatalf("show: code = %d", rec.Code)
	}
	var r Record
	if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
		t.Fatal(err)
	}
	if r.Content != "192.0.2.1" || r.Ttl != 3600 || r.Priority != nil {
		t.Fatalf("record = %+v", r)
	}

	body := `{"domain_id":1,"name":"example.com","type":"MX","content":"mail.example.com","ttl":300,"prio":10}`
	if rec := call(recordCreate, "POST", "/v1/records", body, nil); rec.Code != 200 {
		t.Fatalf("create: code = %d, body = %s", rec.Code, rec.Body)
	}

	if rec := call(recordDelete, "DELETE", "/v1/records/9", "", map[string]string{"id": "9"}); rec.Code != 200 {
		t.Fatalf("delete: code = %d", rec.Code)
	}
}

func TestIndexDatabaseErrorReturns500(t *testing.T) {
	// Previously AllDomains/AllRecords called log.Fatal, killing the whole
	// server on any database error.
	mock := setupDB(t)
	mock.ExpectQuery("select id, name, type from domains").WillReturnError(sql.ErrConnDone)
	mock.ExpectQuery("SELECT id, domain_id, name, type, content, ttl, prio FROM records").WillReturnError(sql.ErrConnDone)

	if rec := call(domainIndex, "GET", "/v1/domains", "", nil); rec.Code != 500 {
		t.Fatalf("domains: code = %d, want 500", rec.Code)
	}
	if rec := call(recordIndex, "GET", "/v1/records", "", nil); rec.Code != 500 {
		t.Fatalf("records: code = %d, want 500", rec.Code)
	}
}

func TestIndexScanErrorReturns500(t *testing.T) {
	mock := setupDB(t)
	mock.ExpectQuery("select id, name, type from domains").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "type"}).AddRow("not-a-number", "x", "NATIVE"))

	if rec := call(domainIndex, "GET", "/v1/domains", "", nil); rec.Code != 500 {
		t.Fatalf("code = %d, want 500", rec.Code)
	}
}

func TestUpdateUsesIDFromURL(t *testing.T) {
	mock := setupDB(t)
	mock.ExpectQuery("SELECT id, name, type FROM domains WHERE id = ?").WithArgs(3).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "type"}).AddRow(3, "old.com", "NATIVE"))
	// The body tries to redirect the update to domain 99.
	mock.ExpectExec("UPDATE domains SET name=? WHERE id=?").WithArgs("new.com", 3).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rec := call(domainUpdate, "POST", "/v1/domains/3", `{"id":99,"name":"new.com"}`, map[string]string{"id": "3"})
	if rec.Code != 200 {
		t.Fatalf("code = %d, body = %s", rec.Code, rec.Body)
	}
}

func TestUpdateMissingReturns404(t *testing.T) {
	mock := setupDB(t)
	mock.ExpectQuery("SELECT id, domain_id, name, type, content, ttl, prio FROM records WHERE id = ?").WithArgs(4).
		WillReturnError(sql.ErrNoRows)

	rec := call(recordUpdate, "POST", "/v1/records/4", `{"name":"x"}`, map[string]string{"id": "4"})
	if rec.Code != 404 {
		t.Fatalf("code = %d, want 404", rec.Code)
	}
}

func TestCreateBadJSONReturns400(t *testing.T) {
	setupDB(t)
	if rec := call(domainCreate, "POST", "/v1/domains", `{not json`, nil); rec.Code != 400 {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
}
