package db

import (
	"URLShortner/pkg"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type GormDatabase struct {
	db     *gorm.DB
	getter *KeyGetter
}

func CreateURLGormDB(dialector gorm.Dialector, opts KeyGetterOptions) (URLDatabase, error) {
	gormDB, err := gorm.Open(dialector)
	if err != nil {
		return nil, fmt.Errorf("failed to setup gorm: %w", err)
	}

	gormDB.AutoMigrate(&pkg.URLShortened{})

	db := &GormDatabase{
		db: gormDB,
	}
	db.getter = CreateDefaultKeyGetter(opts, db)

	return db, nil
}

func (g *GormDatabase) Create(url pkg.URLShortened) error {
	tx := g.db.Create(&url)
	if tx.Error != nil {
		return fmt.Errorf("failed to create url: %w", tx.Error)
	}

	return nil
}

func (g *GormDatabase) Get(key string) (*pkg.URLShortened, error) {
	url := &pkg.URLShortened{
		Key: key,
	}

	tx := g.db.First(url)
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to get url: %w", tx.Error)
	}

	return url, nil
}

func (g *GormDatabase) GetFreeKey() (string, error) {
	return g.getter.GetFreeKey()
}

func (g *GormDatabase) Avaliable(key string) (bool, error) {
	url := &pkg.URLShortened{
		Key: key,
	}

	tx := g.db.First(url)
	if tx.Error != nil && errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return true, nil
	} else if tx.Error != nil {
		return false, tx.Error
	}

	return false, nil
}

func (g *GormDatabase) Delete(key string) error {
	tx := g.db.Delete(&pkg.URLShortened{
		Key: key,
	})
	if tx.Error != nil {
		return fmt.Errorf("failed to delete key %q: %w", key, tx.Error)
	}

	return nil
}

func (g *GormDatabase) DeleteWebhook(key string) error {
	shortened, err := g.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get key %q: %w", key, err)
	}
	shortened.WebHook = nil

	tx := g.db.Save(shortened)
	if tx.Error != nil {
		return fmt.Errorf("failed to update key %q: %w", key, err)
	}

	return nil
}
