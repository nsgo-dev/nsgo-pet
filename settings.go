package main

import (
	"strings"

	"github.com/nsgo-dev/nsgo-pet/pet"
	"github.com/nsgo-dev/nsgo/base"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/desktop"
	"github.com/nsgo-dev/nsgo/updater"
)

// Prefs are what the settings show and change, the app's state.
type Prefs struct {
	Packs    []*pet.Pack
	Chosen   func() string
	Choose   func(id string)
	Stats    func() bool
	SetStats func(on bool)
}

// petSection is pets listed under a group's name, or under none.
type petSection struct {
	Group string
	Packs []*pet.Pack
}

// petSections lists packs by group, the ungrouped first and each group after in the order packs has them.
func petSections(packs []*pet.Pack) []petSection {
	sections := []petSection{{}}
	groups := map[string]int{}
	for _, pk := range packs {
		i, ok := groups[pk.Group]
		if !ok && pk.Group != "" {
			i = len(sections)
			groups[pk.Group] = i
			sections = append(sections, petSection{Group: pk.Group})
		}
		sections[i].Packs = append(sections[i].Packs, pk)
	}
	if len(sections[0].Packs) == 0 {
		sections = sections[1:]
	}
	return sections
}

// settingsPages are the settings' pages by title, the pets first.
var settingsPages = []string{"宠物", "通用"}

// titleBarHeight is the height of the settings' own title bar, which holds the pages' tabs.
const titleBarHeight = 52

// Settings is the settings window: a title bar of its own with a tab for each page, the pets in a grid, each choosing itself as pressed, then checking for updates and the usage statistics.
func Settings(cx *core.Scope, prefs Prefs) core.Node {
	desktop.SetTitleBar(cx, desktop.TitleBarHidden, titleBarHeight)
	page := core.NewSignal(cx, settingsPages[0])

	var tabs []core.Node
	for _, title := range settingsPages {
		tabs = append(tabs, base.Button{Label: title, TestID: "page-" + title,
			Selected: func() bool { return page.Get() == title },
			OnPress:  func(core.PressEvent) { page.Set(title) },
			Class:    "h-7 px-3 justify-center rounded-md hover:bg-black/5 selected:bg-black/10 dark:hover:bg-white/5 dark:selected:bg-white/10",
			Children: []core.Node{core.Text{Class: "text-sm text-neutral-900 dark:text-neutral-100", Content: title}}})
	}
	bar := core.View{Class: "h-[52px] shrink-0 flex-row items-center justify-center gap-1 border-b border-black/10 dark:border-white/10",
		Desktop: core.DesktopViewProps{WindowDrag: core.WindowDragOn},
		StyleFunc: func() core.Style {
			// Clear of the window's controls, alike on both sides to keep the tabs centered.
			e := desktop.TitleBar(cx)
			side := max(e.Left, e.Right)
			return core.Style{Padding: core.Edges{Left: side, Right: side}}
		},
		Children: tabs}

	return core.View{Class: "flex-1 bg-white dark:bg-neutral-900", Children: []core.Node{
		bar,
		core.Dynamic(func(cx *core.Scope) core.Node {
			if page.Get() == settingsPages[0] {
				return petGrid(cx, prefs)
			}
			return general(cx, prefs)
		}),
	}}
}

// petGrid shows every pet idling, by group, the chosen one outlined; a press chooses a pet.
func petGrid(cx *core.Scope, prefs Prefs) core.Node {
	os := cx.Env().OS()
	var sections []core.Node
	for _, s := range petSections(prefs.Packs) {
		var cards []core.Node
		for _, pk := range s.Packs {
			cards = append(cards, base.Button{Label: pk.Name, TestID: "pet-" + pk.ID,
				Selected: func() bool { return prefs.Chosen() == pk.ID },
				OnPress:  func(core.PressEvent) { prefs.Choose(pk.ID) },
				Class: "w-[128px] p-2 gap-1 items-center rounded-lg border-2 border-transparent bg-black/[0.04] dark:bg-white/[0.05] " +
					"hover:bg-black/[0.07] dark:hover:bg-white/[0.08] selected:border-blue-500",
				Children: []core.Node{
					core.Video{Source: idleClip(pk, os), Repeat: true, Muted: true, ResizeMode: core.Contain, Label: pk.Name, TestID: "clip-" + pk.ID, Class: "size-24"},
					core.Text{Class: "text-sm text-neutral-900 dark:text-neutral-100", Content: pk.Name, MaxLines: 1},
				}})
		}
		var children []core.Node
		if s.Group != "" {
			children = append(children, core.Text{Class: "text-xs font-medium text-neutral-500", Content: s.Group})
		}
		children = append(children, core.View{Class: "flex-row flex-wrap gap-3", Children: cards})
		sections = append(sections, core.View{Class: "gap-2", Children: children})
	}
	return core.Scroll{Class: "flex-1", Children: []core.Node{core.View{Class: "p-4 gap-4", Children: sections}}}
}

// idleClip is pk's idle clip in this platform's form, as the engine plays it: HEVC with alpha on macOS, WebM elsewhere.
func idleClip(pk *pet.Pack, os core.OS) string {
	name := strings.TrimSuffix(pk.Policy.Find(pk.Clips.Idle).File, ".webm")
	if os == core.OSMacOS {
		return pk.Base + "/mov/" + name + ".mov"
	}
	return pk.Base + "/webm/" + name + ".webm"
}

// general is checking for updates and the usage statistics.
func general(cx *core.Scope, prefs Prefs) core.Node {
	row := "flex-row items-center justify-between gap-3 min-h-11 px-3"
	return core.View{Class: "flex-1 p-4", Children: []core.Node{
		core.View{Class: "rounded-lg bg-black/[0.04] dark:bg-white/[0.05]", Children: []core.Node{
			core.View{Class: row + " border-b border-black/10 dark:border-white/10", Children: []core.Node{
				core.Text{Class: "text-sm text-neutral-900 dark:text-neutral-100", Content: "检查更新"},
				base.Button{TestID: "update", Label: "检查", OnPress: func(core.PressEvent) { updater.Check(cx) },
					Class:    "px-3 h-7 justify-center rounded-md bg-blue-500 active:bg-blue-600",
					Children: []core.Node{core.Text{Class: "text-sm text-white", Content: "检查"}}},
			}},
			core.View{Class: row, Children: []core.Node{
				core.Text{Class: "text-sm text-neutral-900 dark:text-neutral-100", Content: "发送匿名使用统计"},
				base.Switch{TestID: "stats", Label: "发送匿名使用统计", Checked: prefs.Stats, OnChange: prefs.SetStats,
					TrackClass: "flex-row w-10 h-6 rounded-full bg-neutral-300 dark:bg-neutral-600 checked:bg-blue-500 checked:justify-end",
					ThumbClass: "size-5 m-0.5 rounded-full bg-white"},
			}},
		}},
	}}
}
