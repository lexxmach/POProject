package db

import (
	"URLShortner/pkg"
)

type URLDatabase interface {
	Create(pkg.URLShortened) error
	Get(string) (*pkg.URLShortened, error)
	Avaliable(string) (bool, error)
	DeleteWebhook(string) error
	Delete(string) error

	GetFreeKey() (string, error)
}
