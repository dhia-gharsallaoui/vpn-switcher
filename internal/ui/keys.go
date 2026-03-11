package ui

import "charm.land/bubbles/v2/key"

// KeyMap defines all application key bindings.
type KeyMap struct {
	Quit       key.Binding
	Help       key.Binding
	TabNext    key.Binding
	TabPrev    key.Binding
	Connect    key.Binding
	Disconnect key.Binding
	Refresh    key.Binding
	AddRule    key.Binding
	DeleteRule key.Binding
	ToggleRule key.Binding
	Up         key.Binding
	Down       key.Binding
	Confirm    key.Binding
	Cancel     key.Binding
}

// DefaultKeyMap returns the default key bindings.
var DefaultKeyMap = KeyMap{
	Quit:       key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	TabNext:    key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next tab")),
	TabPrev:    key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("S-tab", "prev tab")),
	Connect:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "connect")),
	Disconnect: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "disconnect")),
	Refresh:    key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	AddRule:    key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add rule")),
	DeleteRule: key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "delete rule")),
	ToggleRule: key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
	Up:         key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("up/k", "up")),
	Down:       key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("down/j", "down")),
	Confirm:    key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "yes")),
	Cancel:     key.NewBinding(key.WithKeys("n", "esc"), key.WithHelp("n/esc", "no")),
}

// HelpBindings returns key bindings formatted for the help view.
func (km KeyMap) HelpBindings() []key.Binding {
	return []key.Binding{
		km.Up, km.Down, km.Connect, km.Disconnect,
		km.Refresh, km.TabNext, km.Help, km.Quit,
	}
}

// RoutingHelpBindings returns key bindings for the routing tab.
func (km KeyMap) RoutingHelpBindings() []key.Binding {
	return []key.Binding{
		km.Up, km.Down, km.AddRule, km.DeleteRule,
		km.ToggleRule, km.TabNext, km.Help, km.Quit,
	}
}
