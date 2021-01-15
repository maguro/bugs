package hooks

import (
	"fmt"
	"sort"

	"gorm.io/gorm"
)

type Parent struct {
	ParentPK uint64   `gorm:"primary_key;autoIncrement:false;column:foo_parent_pk;type:INT8;"`
	Entries  []*Entry `gorm:"-"`
}

type ParentDb struct {
	Parent
	EntriesDb []*EntryDb `gorm:"foreignKey:ParentPK;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// TableName overrides the table name used by ParentDb to `foo_parents`.
func (ParentDb) TableName() string {
	return "foo_parents"
}

func (tdb *ParentDb) AfterFind(tx *gorm.DB) (err error) {
	if tx.Error == nil {
		for _, edb := range tdb.EntriesDb {
			edb.ParentPK = tdb.ParentPK
			tdb.Entries = append(tdb.Entries, &edb.Entry)
		}
		tdb.EntriesDb = nil
	}
	return
}

func (tdb *ParentDb) BeforeSave(_ *gorm.DB) (err error) {
	for _, e := range tdb.Entries {
		tdb.EntriesDb = append(tdb.EntriesDb, &EntryDb{Entry: *e, ParentPK: tdb.ParentPK})
	}
	return
}

func (tdb *ParentDb) AfterSave(_ *gorm.DB) (err error) {
	tdb.EntriesDb = nil
	return
}

// Entry represents an entry.
type Entry struct {
	EntryPK uint64           `gorm:"primary_key;autoIncrement:false;column:foo_entry_pk;type:INT8;"`
	Links   map[string]int64 `gorm:"-"`
}

// EntryDb holds one row from the foo_entries table.
type EntryDb struct {
	Entry
	ParentPK uint64   `gorm:"association_foreignkey:ParentPK;column:foo_parent_pk;type:INT8;"`
	LinksDb  []LinkDb `gorm:"foreignKey:EntryPK;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// TableName overrides the table name used by EntryDb to `foo_entries`.
func (EntryDb) TableName() string {
	return "foo_entries"
}

func (edb *EntryDb) AfterFind(tx *gorm.DB) (err error) {
	if tx.Error == nil {
		edb.Links = make(map[string]int64)
		for _, l := range edb.LinksDb {
			if l.Key == "" {
				return fmt.Errorf("empty key for entry links")
			}
			edb.Links[l.Key] = l.Link
		}
		edb.LinksDb = nil
	}
	return
}

func (edb *EntryDb) BeforeSave(_ *gorm.DB) (err error) {
	var sorted []string
	for key := range edb.Links {
		sorted = append(sorted, key)
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	for _, key := range sorted {
		edb.LinksDb = append(edb.LinksDb, LinkDb{EntryPK: edb.EntryPK, Key: key, Link: edb.Links[key]})
	}
	return
}

func (edb *EntryDb) AfterSave(_ *gorm.DB) (err error) {
	edb.LinksDb = nil
	return
}

// LinkDb holds one row from the foo_entry_links table.
type LinkDb struct {
	EntryPK uint64 `gorm:"primary_key;autoIncrement:false;association_foreignkey:EntryPK;column:foo_entry_pk;type:INT8;"`
	Key     string `gorm:"primary_key;column:foo_key;type:VARCHAR;size:64;"`
	Link    int64  `gorm:"column:foo_link;type:INT4;"`
}

// TableName overrides the table name used by LinkDb to `foo_entry_links`.
func (LinkDb) TableName() string {
	return "foo_entry_links"
}
