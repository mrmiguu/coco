package main

type Logo struct {
	Icon string
	Name string
	cur  int
}

func NewLogo() Logo {
	return Logo{
		Icon: "🥥",
		Name: "Coco",
	}
}

// func (o Logo) OnLogoClick() {
// 	println("OnLogoClick!")

// 	o.cur = (o.cur + 1) % len(o.Name)
// 	o.Name = strings.ToLower(o.Name)

// 	// coco.Set(o)
// }

// func (o Logo) OnCocoClick() {
// 	println("OnCocoClick!")
// }
