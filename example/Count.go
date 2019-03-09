package main

import "github.com/mrmiguu/coco"

type Count struct {
	Up
	Cocos []string
	Down
}

func NewCount() Count {
	return Count{
		Cocos: []string{"🥥"},
	}
}

func (c Count) OnUpClick() {
	println("clicked Up")
	c.Cocos = append(c.Cocos, "🥥")
	coco.Set(c)
}

func (c Count) OnCocoClick() {
	println("clicked Coco")
}

func (c Count) OnDownClick() {
	println("clicked Down")
	c.Cocos = c.Cocos[1:]
	coco.Set(c)
}
