package internal

import (
	"encoding/json"
	"os"
)

type ConfigDB struct {
	Mock bool `json:"mock"`

	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Port     uint16 `json:"port"`
}

type ConfigURLSettings struct {
	MinLength uint16 `json:"min_len"`
	MaxLength uint16 `json:"max_len"`

	Retries uint16 `json:"retries"`
	Runes   string `json:"runes"`
}

type Config struct {
	Title   string `json:"title"`
	Version string `json:"version"`

	Port int `json:"port"`

	DB ConfigDB `json:"db"`

	URLShortenedSettins ConfigURLSettings `json:"url_settings"`

	FollowLink string `json:"follow_link"`

	JWTSecret string `json:"jwt_secret"`

	LogLevel string `json:"log_level"`
}

func GetConfig(path string) (*Config, error) {
	plan, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = json.Unmarshal(plan, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func MustGetConfig(path string) *Config {
	return Must(GetConfig(path))
}
