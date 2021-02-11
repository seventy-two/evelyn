package database

import (
	"encoding/json"
	"log"

	"github.com/bwmarrin/discordgo"
	badger "github.com/dgraph-io/badger/v2"
)

// Database holds a badget db, duh
type Database struct {
	db *badger.DB
}

// NewDatabase returns a new or existing db depending on if a db exists at dbPath or not
func NewDatabase(dbPath string) *Database {
	db, err := badger.Open(badger.DefaultOptions(dbPath))
	if err != nil {
		log.Fatal(err)
	}

	return &Database{db: db}
}

// Cunt is a Discord user from Evelyn's perspective
type Cunt struct {
	ID     string
	Member discordgo.Member
	Info   *CuntInfo
}

// CuntInfo contains all stored information for a cunt
type CuntInfo struct {
	Location  string
	Watchlist []string
}

// SetCunt will store or update a Cunt
func (d *Database) SetCunt(c *Cunt) error {
	if c.ID == "261097001301704704" {
		c.Info = &CuntInfo{}
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err := d.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(c.ID), b)
		return err
	}); err != nil {
		return err
	}
	return nil
}

// GetCunt will get a Cunt
func (d *Database) GetCunt(id string) (*Cunt, error) {
	var v []byte
	if err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
		if err != nil {
			return err
		}

		v, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	c := &Cunt{}
	if err := json.Unmarshal(v, c); err != nil {
		return nil, err
	}

	return c, nil
}

// UpdateCuntInfo will update CuntInfo for a given Cunt if existing, or create a new Cunt if not
func (d *Database) UpdateCuntInfo(id string, c *CuntInfo) error {
	cunt, err := d.GetCunt(id)
	if err != nil || cunt == nil {
		err := d.NewCunt(id, c)
		return err
	}
	newInfo := mergeCuntInfo(c, cunt.Info)
	cunt.Info = newInfo
	if err := d.SetCunt(cunt); err != nil {
		return err
	}
	return nil
}

// RemoveCuntInfo will remove select CuntInfo
func (d *Database) RemoveCuntInfo(id string, c *CuntInfo) error {
	cunt, err := d.GetCunt(id)
	if err != nil || cunt == nil {
		err := d.NewCunt(id, c)
		return err
	}
	newInfo := inverseMerge(c, cunt.Info)
	cunt.Info = newInfo
	if err := d.SetCunt(cunt); err != nil {
		return err
	}
	return nil
}

// NewCunt creates a new Cunt with the given CuntInfo
func (d *Database) NewCunt(id string, c *CuntInfo) error {
	cunt := &Cunt{
		ID:   id,
		Info: c,
	}
	if err := d.SetCunt(cunt); err != nil {
		return err
	}
	return nil
}

func mergeCuntInfo(c1, c2 *CuntInfo) *CuntInfo {
	if c1.Location != "" {
		c2.Location = c1.Location
	}
	if len(c1.Watchlist) > 0 {
		for _, stock1 := range c1.Watchlist {
			hasStock := false
			for _, stock2 := range c2.Watchlist {
				if stock2 == stock1 {
					hasStock = true
				}
			}
			if !hasStock {
				c2.Watchlist = append(c2.Watchlist, stock1)
			}
		}
	}
	return c2
}

func inverseMerge(c1, c2 *CuntInfo) *CuntInfo {
	newWL := c2.Watchlist
	if len(c1.Watchlist) > 0 {
		for _, entry := range c1.Watchlist {
			for i := range c2.Watchlist {
				if c2.Watchlist[i] == entry {
					newWL = append(c2.Watchlist[:i], c2.Watchlist[i+1:]...)
				}
			}
		}
	}
	c2.Watchlist = newWL
	return c2
}
