package main

import "github.com/charmbracelet/bubbles/key"

type keybindings struct {
	ForceQuit key.Binding
	Quit      key.Binding
	Execute   key.Binding
	PrevExpr  key.Binding
	NextExpr  key.Binding
}

var keys = keybindings{
	ForceQuit: key.NewBinding(key.WithKeys("ctrl+c")),
	Quit:      key.NewBinding(key.WithKeys("esc")),
	Execute:   key.NewBinding(key.WithKeys("enter")),
	PrevExpr:  key.NewBinding(key.WithKeys("up", "tab")),
	NextExpr:  key.NewBinding(key.WithKeys("down", "shift+tab")),
}
