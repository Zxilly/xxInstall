package main

import (
	"encoding/json"
	"os"
)

type persist struct {
	configURL string
}


func (p *persist) save() error {
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}
	err = os.WriteFile("persist.json", s, 0644)
	return err
}

func (p *persist) load() error {
	s, err := os.ReadFile("persist.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(s, p)
	return err
}