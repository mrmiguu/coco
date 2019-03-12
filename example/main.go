package main

import "github.com/mrmiguu/coco"

func main() {
	coco.Globals(NewGlobals())
	coco.Render(NewApp())
}
