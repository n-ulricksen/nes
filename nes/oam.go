package nes

import "fmt"

type objectAttributeMemory []oamSprite

// oamSprite represents one entry, or sprite, in the Object Attribute memory.
type oamSprite struct {
	y         byte // Y position of the sprite
	id        byte // pattern memory ID
	attribute byte // flag specifying rendering attributes
	x         byte // X position of the sprite
}

func (oam objectAttributeMemory) read(addr byte) byte {
	spriteIdx := int(addr) / 4
	propIdx := int(addr) % 4

	sprite := oam[spriteIdx]

	var data byte
	switch propIdx {
	case 0:
		data = sprite.y
	case 1:
		data = sprite.id
	case 2:
		data = sprite.attribute
	case 3:
		data = sprite.x
	}

	return data
}

func (oam objectAttributeMemory) write(addr byte, data byte) {
	spriteIdx := int(addr) / 4
	propIdx := int(addr) % 4

	sprite := oam[spriteIdx]

	switch propIdx {
	case 0:
		sprite.y = data
	case 1:
		sprite.id = data
	case 2:
		sprite.attribute = data
	case 3:
		sprite.x = data
	}
}

func (oam objectAttributeMemory) clear() {
	for i := range oam {
		oam[i].y = 0xFF
		oam[i].id = 0xFF
		oam[i].attribute = 0xFF
		oam[i].x = 0xFF
	}
}

func copyOamEntry(to, from *oamSprite) {
	fmt.Println(*to, *from)
	to.y = from.y
	to.id = from.id
	to.attribute = from.attribute
	to.x = from.x
}