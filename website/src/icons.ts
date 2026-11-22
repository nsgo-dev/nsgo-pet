// The Kubai Icons the site uses, by name. The server renders their markup
// and the browser imports the same ones to animate them, so only these are
// bundled.
import {
	AppWindow,
	CalendarClock,
	Download,
	Footprints,
	Globe,
	Languages,
	Laptop,
	Layers,
	Monitor,
	Moon,
	MousePointerClick,
	PackageDown,
	Paw,
	Sun,
	Terminal,
	type KubaiIconData,
} from 'kubai-icons';

export const icons = {
	'app-window': AppWindow,
	'calendar-clock': CalendarClock,
	download: Download,
	footprints: Footprints,
	globe: Globe,
	languages: Languages,
	laptop: Laptop,
	layers: Layers,
	monitor: Monitor,
	moon: Moon,
	'mouse-pointer-click': MousePointerClick,
	'package-down': PackageDown,
	paw: Paw,
	sun: Sun,
	terminal: Terminal,
} satisfies Record<string, KubaiIconData>;

export type IconName = keyof typeof icons;
