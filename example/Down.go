package main

import "github.com/mrmiguu/coco"

type Down struct {
	Globals
}

func (d Down) OnDownClick() {

	if len(d.Cocos) > 0 {
		println("Cocos reduced.")
		d.Cocos = d.Cocos[1:]
	}

	coco.Set(d)
}
