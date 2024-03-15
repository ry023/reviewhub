package reviewhub

import (
	"os"

	"github.com/go-playground/validator"
	"github.com/go-yaml/yaml"
)

type Config struct {
	Retrievers []RetrieverConfig
	Notifiers  []NotifierConfig
	Users      []User
}

type MetaData any

type NotifierConfig struct {
	Name     string
	Type     string
	MetaData MetaData
}

type RetrieverConfig struct {
	Name     string
	Type     string
	MetaData MetaData
}

func NewConfig(filepath string) (*Config, error) {
	b, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var config *Config
	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, err
	}

	return config, nil
}

func ParseMetaData[T any](raw MetaData) (*T, error) {
	b, err := yaml.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var m T
	err = yaml.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}

	err = validator.New().Struct(m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
