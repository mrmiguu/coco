package main

import "strings"
import "./coco"

type App struct {
	Logo Logo
}

func NewApp() App {
	return App{
		NewLogo(),
	}
}

func (a App) OnNameClick() {
	a.Logo.cur = (a.Logo.cur + 1) % len(a.Logo.Name)
	name := strings.ToLower(a.Logo.Name)

	head := name[:a.Logo.cur]
	letter := name[a.Logo.cur]
	tail := name[a.Logo.cur+1:]

	a.Logo.Name = head + strings.ToUpper(string(letter)) + tail

	coco.Set(a)
}

func (a App) OnIconClick() {
	runes := []rune(a.Logo.Name)
	for i, j := 0, len(a.Logo.Name)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]

		if i == a.Logo.cur {
			a.Logo.cur = j
		} else if j == a.Logo.cur {
			a.Logo.cur = i
		}
	}
	a.Logo.Name = string(runes)

	coco.Set(a)
}
