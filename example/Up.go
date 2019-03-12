package main

import "github.com/mrmiguu/coco"

type Up struct {
	Globals
}

func (u Up) OnUpClick() {
	println("Cocos increased.")

	u.Cocos = append(u.Cocos, "🥥")

	coco.Set(u)
}
