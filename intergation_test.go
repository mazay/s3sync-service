package main

import (
	"github.com/bxcodec/faker"
)

func FakeConfig() (Config, error) {
	a := Config{}
	err := faker.FakeData(&a)
	if err != nil {
		return a, err
	}
	return a, nil
}
