// Command nsgo-pet is a desktop pet on NSGo whose pet can be switched: each pet in resources/pets lives in a transparent window above the other windows and plays its clips by its own rules, in Big Blue Fish's format. It is a host of the pet engine, package pet, which ships no pets.
package main

import (
	_ "embed"
	"log"

	"github.com/nsgo-dev/nsgo"
	"github.com/nsgo-dev/nsgo-pet/pet"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/desktop"
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

// trayIcon is NSGo-chan's hooded head in silhouette, a template the menu bar colors; assets/Tray.svg is its source.
//
//go:embed assets/Tray.png
var trayIcon []byte

// App is the pet's window, showing the pet the user chose last, the first of packs at first, with the tray icon whose menu opens the settings and quits.
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
		stats := core.NewSignal(cx, statsOn(cx))
		var current *pet.Pet
		prefs := Prefs{
			Packs:  packs,
			Chosen: chosen.Get,
			Choose: func(id string) {
				if id == chosen.GetUntracked() {
					return
				}
				if err := cx.Storage().SetItem(petKey, id); err != nil {
					log.Printf("saving the pet: %v", err)
				}
				chosen.Set(id)
			},
			Stats: stats.Get,
			SetStats: func(on bool) {
				setStatsOn(cx, on)
				stats.Set(on)
			},
		}
		settings := func() {
			desktop.OpenWindow(cx, desktop.WindowOptions{Title: "设置", Width: 620, Height: 500, Key: "settings",
				Constraints: desktop.WindowConstraints{NoResize: true, NoMinimize: true}},
				func(cx *core.Scope) core.Node { return Settings(cx, prefs) })
		}
		quit := func() {
			if current == nil {
				cx.Quit()
				return
			}
			current.Leave(cx.Quit)
		}
		report := statsReporter(cx, stats.GetUntracked)
		desktop.NewTray(cx, desktop.TrayOptions{Icon: trayIcon, Template: true, Tooltip: "nsgo-pet",
			Menu: func() []core.MenuItem { return trayMenu(settings, quit) }})
		return core.Dynamic(func(cx *core.Scope) core.Node {
			pack := find(chosen.Get())
			report(pack)
			p := pet.New(cx, pack, pet.Options{Window: true})
			current = p
			cx.Post(p.Start)
			return p.View()
		})
	}
}

// trayMenu is the tray icon's menu.
func trayMenu(settings, quit func()) []core.MenuItem {
	return []core.MenuItem{
		{Label: "设置", Shortcut: "mod+,", OnSelect: settings},
		{Separator: true},
		{Label: "退出", Shortcut: "mod+q", OnSelect: quit},
	}
}
