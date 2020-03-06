package main

import (
	"github.com/golang/protobuf/proto"
)

func (app *application) getCache(key string, pb proto.Message) error {
	b, err := app.cache.Get(key)
	if err != nil {
		return err
	}
	return proto.Unmarshal(b, pb)
}

func (app *application) setCache(key string, pb proto.Message) error {
	b, err := proto.Marshal(pb)
	if err != nil {
		return err
	}
	return app.cache.Set(key, b)
}
