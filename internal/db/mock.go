package db

import (
	"URLShortner/pkg"
	"fmt"
)

type MockDatabase struct {
	db     map[string]*pkg.URLShortened
	getter *KeyGetter
}

func CreateURLMockDB(opts KeyGetterOptions) URLDatabase {
	db := &MockDatabase{
		db: make(map[string]*pkg.URLShortened),
	}
	db.getter = CreateDefaultKeyGetter(opts, db)

	return db
}

func (m *MockDatabase) Avaliable(key string) (bool, error) {
	_, ok := m.db[key]
	return !ok, nil
}

func (m *MockDatabase) Create(url pkg.URLShortened) error {
	if ok, _ := m.Avaliable(url.Key); !ok {
		return fmt.Errorf("mock db error, key %q already exist", url.Key)
	}
	m.db[url.Key] = &url
	return nil
}

func (m *MockDatabase) Get(key string) (*pkg.URLShortened, error) {
	val, ok := m.db[key]
	if !ok {
		return nil, fmt.Errorf("mock db error, key %q doenst exit", key)
	}
	return val, nil
}

func (m *MockDatabase) GetFreeKey() (string, error) {
	return m.getter.GetFreeKey()
}

func (m *MockDatabase) Delete(key string) error {
	if ok, _ := m.Avaliable(key); ok {
		return fmt.Errorf("mock db error, key %q doesnt exist", key)
	}
	delete(m.db, key)

	return nil
}

// DeleteWebhook implements URLDatabase.
func (m *MockDatabase) DeleteWebhook(key string) error {
	if ok, _ := m.Avaliable(key); ok {
		return fmt.Errorf("mock db error, key %q doesnt exist", key)
	}
	m.db[key].WebHook = nil

	return nil
}
