package main

import (
	"encoding/json"
	"os"
)

type persist struct {
	ConfigURL string `json:"config_url"`
}

func (p *persist) save() error {
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}
	err = os.WriteFile(PersistFile, s, 0644)
	return err
}

func (p *persist) load() error {
	s, err := os.ReadFile(PersistFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(s, p)
	return err
}
