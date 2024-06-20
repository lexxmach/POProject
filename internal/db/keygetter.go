package db

import (
	"fmt"
	"math/rand"
)

type KeyGetterOptions struct {
	MinLength uint16
	MaxLength uint16
	Retries   uint16

	RuneSet []rune
}

type KeyGetter struct {
	minLength uint16
	maxLength uint16
	retries   uint16

	runeSet        []rune
	checkAvaliable func(string) (bool, error)
}

func CreateDefaultKeyGetter(opts KeyGetterOptions, db URLDatabase) *KeyGetter {
	return &KeyGetter{
		minLength: opts.MinLength,
		maxLength: opts.MaxLength,
		retries:   opts.Retries,

		runeSet:        opts.RuneSet,
		checkAvaliable: db.Avaliable,
	}
}

func (k *KeyGetter) GetFreeKey() (string, error) {
	var lastError error
	for len := k.minLength; len <= k.maxLength; len++ {
		for i := uint16(0); i < k.retries; i++ {
			keyCheck := randString(k.runeSet, len)

			var avaliable bool
			if avaliable, lastError = k.checkAvaliable(keyCheck); avaliable && lastError == nil {
				return keyCheck, nil
			}
		}
	}
	return "", fmt.Errorf("failed to get free key, last error: %w", lastError)
}

func randString(letters []rune, length uint16) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
