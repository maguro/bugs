package hooks_test

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"m4o.io/bugs/gorm/hooks"
)

func TestInsertEntryDb(t *testing.T) {
	sqlDb, mock, db, err := connections()
	assert.NoError(t, err)
	defer sqlDb.Close()

	var edb hooks.EntryDb
	edb.ParentPK = 1

	e := &edb.Entry
	e.EntryPK = 123
	e.Links = map[string]int64{"one": 1, "two": 2, "three": 3}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO \"foo_entries\" \\(.*\\) VALUES \\(.*\\)").
		WithArgs(123, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO \"foo_entry_links\" \\(.*\\) VALUES \\(.*\\).*").
		WithArgs(123, "one", 1, 123, "three", 3, 123, "two", 2).
		WillReturnResult(sqlmock.NewResult(4, 3))
	mock.ExpectCommit()

	result := db.Create(&edb)
	assert.NoError(t, result.Error)

	// ensure edb is cleaned up after save
	assert.Equal(t, hooks.EntryDb{
		ParentPK: 1,
		Entry: hooks.Entry{
			EntryPK: 123,
			Links:   map[string]int64{"one": 1, "two": 2, "three": 3},
		}}, edb)
}

func TestSelectEntryDb(t *testing.T) {
	sqlDb, mock, db, err := connections()
	assert.NoError(t, err)
	defer sqlDb.Close()

	var edb hooks.EntryDb

	mock.ExpectQuery("SELECT \\* FROM \"foo_entries\" WHERE .*").
		WithArgs(123).
		WillReturnRows(
			sqlmock.NewRows([]string{"foo_entry_pk", "foo_parent_pk"}).
				AddRow(123, 1))
	mock.ExpectQuery("SELECT \\* FROM \"foo_entry_links\" WHERE .*").
		WithArgs(123).
		WillReturnRows(
			sqlmock.NewRows([]string{"foo_entry_pk", "foo_key", "foo_link"}).
				AddRow(123, "one", 1).
				AddRow(123, "two", 2).
				AddRow(123, "three", 3))

	result := db.Preload("LinksDb").Find(&edb, 123)
	assert.NoError(t, result.Error)
	assert.Equal(t, hooks.EntryDb{
		Entry: hooks.Entry{
			EntryPK: 123,
			Links:   map[string]int64{"one": 1, "two": 2, "three": 3},
		},
		ParentPK: 1,
	}, edb)
}

func TestInsertParentDb(t *testing.T) {
	sqlDb, mock, db, err := connections()
	assert.NoError(t, err)
	defer sqlDb.Close()

	var pdb hooks.ParentDb
	pdb.ParentPK = 1

	p := &pdb.Parent
	p.Entries = []*hooks.Entry{{
		EntryPK: 123,
		Links:   map[string]int64{"one": 1, "two": 2, "three": 3},
	}}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO \"foo_parents\" \\(.*\\) VALUES \\(.*\\)").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO \"foo_entries\" \\(.*\\) VALUES \\(.*\\)").
		WithArgs(123, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO \"foo_entry_links\" \\(.*\\) VALUES \\(.*\\).*").
		WithArgs(123, "one", 1, 123, "three", 3, 123, "two", 2).
		WillReturnResult(sqlmock.NewResult(4, 3))
	mock.ExpectCommit()

	result := db.Create(&pdb)
	assert.NoError(t, result.Error)

	// ensure pdb is cleaned up after save
	assert.Equal(t, hooks.ParentDb{
		Parent: hooks.Parent{ParentPK: 1,
			Entries: []*hooks.Entry{{
				EntryPK: 123,
				Links:   map[string]int64{"one": 1, "two": 2, "three": 3},
			}}},
	}, pdb)
}

func TestSelectParentDb(t *testing.T) {
	sqlDb, mock, db, err := connections()
	assert.NoError(t, err)
	defer sqlDb.Close()

	var pdb hooks.ParentDb

	mock.ExpectQuery("SELECT \\* FROM \"foo_parents\" WHERE .*").
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"foo_parent_pk"}).
				AddRow(1))
	mock.ExpectQuery("SELECT \\* FROM \"foo_entries\" WHERE .*").
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"foo_entry_pk", "foo_parent_pk"}).
				AddRow(123, 1))
	mock.ExpectQuery("SELECT \\* FROM \"foo_entry_links\" WHERE .*").
		WithArgs(123).
		WillReturnRows(
			sqlmock.NewRows([]string{"foo_entry_pk", "foo_key", "foo_link"}).
				AddRow(123, "one", 1).
				AddRow(123, "two", 2).
				AddRow(123, "three", 3))

	result := db.Preload("EntriesDb.LinksDb").Find(&pdb, 1)
	assert.NoError(t, result.Error)
	assert.Equal(t, hooks.ParentDb{
		Parent: hooks.Parent{
			ParentPK: 1,
			Entries: []*hooks.Entry{{
				EntryPK: 123,
				Links:   map[string]int64{"one": 1, "two": 2, "three": 3},
			}}},
	}, pdb)
}

func connections() (*sql.DB, sqlmock.Sqlmock, *gorm.DB, error) {
	sqlDb, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, nil, err
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDb,
	}), &gorm.Config{})

	return sqlDb, mock, db, err
}
