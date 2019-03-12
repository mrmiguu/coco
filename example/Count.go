package main

import "strconv"

type Count struct {
	Globals
	Up
	Down
}

func NewCount() Count {
	return Count{}
}

func (c Count) OnCocosClick() {
	println("clicked " + strconv.Itoa(len(c.Cocos)) + " cocos.")
}
