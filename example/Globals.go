package main

type Globals struct {
	Cocos []string
}

func NewGlobals() Globals {
	return Globals{
		Cocos: []string{"🥥"},
	}
}
