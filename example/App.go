package main

type App struct {
	Name string
	Count
}

func NewApp() App {
	return App{
		"My Coco App",
		NewCount(),
	}
}

func (a App) OnNameClick() {
	println("clicked Name")
}
