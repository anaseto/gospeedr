// +build sdl

package main

import (
	"image"
	"image/color"
	"io/ioutil"
	"log"

	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/tiles"
)

const TTF = true

func getTileDrawer() (*TileDrawer, error) {
	t := &TileDrawer{}
	ttf := gomono.TTF
	var err error
	if OptTTF != "" {
		ttf, err = ioutil.ReadFile(OptTTF)
		if err != nil {
			log.Println("font file: %s", err)
		}
	}
	// We get a monospace font TTF.
	font, err := opentype.Parse(ttf)
	if err != nil {
		return nil, err
	}
	// We retrieve a font face.
	face, err := opentype.NewFace(font, &opentype.FaceOptions{
		Size: 24,
		DPI:  72,
	})
	if err != nil {
		return nil, err
	}
	// We create a new drawer for tiles using the previous face. Note that
	// if more than one face is wanted (such as an italic or bold variant),
	// you would have to create drawers for thoses faces too, and then use
	// the relevant one accordingly in the GetImage method.
	t.drawer, err = tiles.NewDrawer(face)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// Tile implements TileManager.
type TileDrawer struct {
	drawer *tiles.Drawer
}

func (t *TileDrawer) GetImage(c gruid.Cell) *image.RGBA {
	// we use some selenized colors
	fg := image.NewUniform(color.RGBA{0xad, 0xbc, 0xbc, 255})
	bg := image.NewUniform(color.RGBA{0x18, 0x49, 0x56, 255})
	switch c.Style.Fg {
	case ColorYellow:
		fg = image.NewUniform(color.RGBA{0xdb, 0xb3, 0x2d, 255})
	case ColorMagenta:
		fg = image.NewUniform(color.RGBA{0xf2, 0x75, 0xbe, 255})
	case ColorCyan:
		fg = image.NewUniform(color.RGBA{0x41, 0xc7, 0xb9, 255})
	case ColorGreen:
		fg = image.NewUniform(color.RGBA{0x75, 0xb9, 0x38, 255})
	case ColorRed:
		fg = image.NewUniform(color.RGBA{0xfa, 0x57, 0x50, 255})
	}
	return t.drawer.Draw(c.Rune, fg, bg)
}

func (t *TileDrawer) TileSize() gruid.Point {
	return t.drawer.Size()
}
