// Command nsgo-pet is a desktop pet on NSGo whose pet can be switched: each pet in resources/pets lives in a transparent window above the other windows and plays its clips by its own rules, in Big Blue Fish's format. It is a host of the pet engine, package pet, which ships no pets.
package main

import (
	"log"

	"github.com/nsgo-dev/nsgo"
	"github.com/nsgo-dev/nsgo-pet/pet"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/desktop"
	"github.com/nsgo-dev/nsgo/updater"
)

func main() {
	packs, err := LoadPacks()
	if err != nil {
		log.Fatal(err)
	}
	nsgo.Run(App(packs), nsgo.WindowSize(packs[0].Width, packs[0].Height), nsgo.WindowTransparent(), nsgo.WindowNoShadow(), nsgo.SkipTaskbar(), nsgo.HideDock())
}

// petKey is the storage key of the pet the user chose last.
const petKey = "pet"

// App is the pet's window, showing the pet the user chose last, the first of packs at first.
func App(packs []*pet.Pack) func(cx *core.Scope) core.Node {
	return func(cx *core.Scope) core.Node {
		find := func(id string) *pet.Pack {
			for _, p := range packs {
				if p.ID == id {
					return p
				}
			}
			return packs[0]
		}
		saved, _ := cx.Storage().GetItem(petKey)
		chosen := core.NewSignal(cx, find(saved).ID)
		var current *pet.Pet
		choose := func(id string) {
			if id == chosen.GetUntracked() {
				return
			}
			if err := cx.Storage().SetItem(petKey, id); err != nil {
				log.Printf("saving the pet: %v", err)
			}
			chosen.Set(id)
		}
		menu := func() []core.MenuItem {
			items := []core.MenuItem{
				{Label: "切换桌宠", Items: petMenu(packs, chosen.GetUntracked(), choose)},
				// Sparkle's windows on macOS, an alert that installs the signed package on Linux.
				{Label: "检查更新…", OnSelect: func() { updater.Check(cx) }},
				{Separator: true},
			}
			if current != nil {
				items = append(items, petActions(cx, current)...)
			}
			return items
		}
		desktop.NewTray(cx, desktop.TrayOptions{Title: "🐾", Tooltip: "nsgo-pet", Menu: menu})
		return core.Dynamic(func(cx *core.Scope) core.Node {
			p := pet.New(cx, find(chosen.Get()), pet.Options{Window: true, Menu: menu})
			current = p
			cx.Post(p.Start)
			return p.View()
		})
	}
}

// petMenu lists packs to choose from, those of a group in a submenu of its name after the ungrouped ones, each in the order packs has them.
func petMenu(packs []*pet.Pack, chosen string, choose func(id string)) []core.MenuItem {
	var top []core.MenuItem
	groups := map[string]int{}
	var grouped []core.MenuItem
	for _, pk := range packs {
		item := core.MenuItem{Label: pk.Name, Radio: true, Checked: pk.ID == chosen, OnSelect: func() { choose(pk.ID) }}
		if pk.Group == "" {
			top = append(top, item)
			continue
		}
		i, ok := groups[pk.Group]
		if !ok {
			i = len(grouped)
			groups[pk.Group] = i
			grouped = append(grouped, core.MenuItem{Label: pk.Group})
		}
		grouped[i].Items = append(grouped[i].Items, item)
	}
	return append(top, grouped...)
}

// petActions are the menu's items for the pet that shows.
func petActions(cx *core.Scope, p *pet.Pet) []core.MenuItem {
	return []core.MenuItem{
		{Label: "来点动作", OnSelect: p.Act},
		{Label: "回到右下角", OnSelect: p.PlaceAtCorner},
		{Separator: true},
		{Label: "退出", OnSelect: func() { p.Leave(cx.Quit) }},
	}
}
