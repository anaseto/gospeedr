// +build sdl

package main

import (
	"log"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid-sdl"
)

var driver gruid.Driver

func initDriver() {
	t, err := getTileDrawer()
	if err != nil {
		log.Fatalf("could not initialize font drawing: %v", err)
	}
	dr := sdl.NewDriver(sdl.Config{
		TileManager: t,
	})
	//dr.SetScale(2.0, 2.0)
	dr.PreventQuit()
	driver = dr
}
