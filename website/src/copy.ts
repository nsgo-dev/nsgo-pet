// The site's words. Facts come from the project's README, CHANGELOG and
// docs; keep both languages saying the same thing. The home page shows the
// app; the engine is for developers, in the docs.
import type { StarlightIcon } from '@astrojs/starlight/types';
import type { Lang } from './site';

/** A link that leaves the site, with a Starlight icon or 'nsgo' for NSGo's logo. */
export type Link = { label: string; href: string; icon: StarlightIcon | 'nsgo'; note?: string; avatar?: ImageMetadata };

export type Platform = {
	id: 'macos' | 'linux' | 'windows';
	name: string;
	status: 'runs' | 'later';
	text: string;
};

const community = (t: { nsgoNote: string; telegram: string; awesomeNote: string }): Link[] => [
	{ label: 'NSGo', note: t.nsgoNote, href: 'https://nsgo.dev', icon: 'nsgo' },
	{ label: t.telegram, note: '@nsgo_dev', href: 'https://t.me/nsgo_dev', icon: 'telegram' },
	{ label: 'Awesome NSGo', note: t.awesomeNote, href: 'https://awesomensgo.dev/', icon: 'star' },
];

/** Where readers report problems and share ideas. */
export const feedback = 'https://github.com/nsgo-dev/nsgo-pet/issues';

const en = {
	htmlLang: 'en',
	title: 'nsgo-pet: a desktop pet for macOS and Linux',
	description: 'A desktop pet for macOS and Linux built with NSGo: it floats above your windows, greets you, reacts to your clicks and naps at night.',
	nav: { label: 'Main', menu: 'Menu', docs: 'Docs', pets: 'Pets', github: 'GitHub', community: 'Community', feedback: 'Feedback' },
	community: community({ nsgoNote: 'The UI engine nsgo-pet is built on', telegram: 'NSGo Telegram', awesomeNote: 'Apps, libraries and resources' }),
	theme: { name: 'Theme', light: 'Light', dark: 'Dark', auto: 'Auto' },
	language: { name: 'Language', auto: 'Auto' },
	headline: ['A pet that lives', 'on your desktop.'],
	lead: 'nsgo-pet floats above your windows on macOS and Linux. It greets you, reacts to your clicks, walks along the bottom of the screen and naps at night.',
	cta: { download: 'Download for free', start: 'Get started' },
	meta: ['macOS', 'Linux', '11 pets'],
	buddy: 'NSGo-chan: click her, or drag her around',
	pets: {
		eyebrow: 'Pets',
		title: 'Eleven pets, one click apart',
		text: 'Switch pets in the settings the tray icon opens, and nsgo-pet remembers your choice. Each pet keeps its own license.',
		names: { 'nsgo-chan': 'NSGo-chan', bigbluefish: 'Big Blue Fish' } as Record<string, string>,
		clips: (n: number) => `${n} clips`,
		nonCommercial: 'Non-commercial',
	},
	features: {
		eyebrow: 'The app',
		title: 'Company that stays out of the way',
		items: [
			{ icon: 'layers', title: 'Above everything', text: 'Floats in a transparent window above other windows, on every Space and over full-screen apps, and stays off the Dock and the taskbar.' },
			{ icon: 'mouse-pointer-click', title: 'Clicks pass through', text: 'Takes the pointer only on the pet’s body, so clicks elsewhere reach the window below.' },
			{ icon: 'calendar-clock', title: 'Knows what time it is', text: 'Plays events chosen by time of day, season, meals, festivals and whether an editor or a terminal is in front.' },
			{ icon: 'footprints', title: 'Alive under the pointer', text: 'Reacts to clicks, dangles while dragged and walks along the bottom of the screen.' },
			{ icon: 'paw', title: 'Eleven pets', text: 'Switch between them in the settings the tray icon opens; your choice is remembered.' },
			{ icon: 'package-down', title: 'Updates itself', text: 'Installs updates from the repository’s releases, checked daily and from the settings.' },
		],
	},
	platforms: {
		eyebrow: 'Platforms',
		title: 'On macOS and Linux today',
		status: { runs: 'Runs', later: 'Not yet' },
		items: [
			{ id: 'macos', name: 'macOS', status: 'runs', text: 'Plays the HEVC clips with alpha.' },
			{ id: 'linux', name: 'Linux', status: 'runs', text: 'Plays the WebM clips. Floats on X11; on Wayland it cannot float or move itself.' },
			{ id: 'windows', name: 'Windows', status: 'later', text: 'Waits on NSGo’s Windows player, which refuses clips with alpha.' },
		] satisfies Platform[],
	},
	end: {
		site: 'NSGo website',
		title: 'Build with NSGo',
		text: 'nsgo-pet is an NSGo app: its transparent window, tray, menus and clips are NSGo’s, from one Go codebase. NSGo is a cross-platform native UI engine with a Go backend.',
	},
	footer: {
		copyright: '© 2026 nsgo-dev',
		license: { label: 'MIT License', href: 'https://github.com/nsgo-dev/nsgo-pet/blob/main/LICENSE' },
		links: [
			{ label: 'Docs', href: 'docs/' },
			{ label: 'Feedback', href: 'https://github.com/nsgo-dev/nsgo-pet/issues' },
			{ label: 'Changelog', href: 'https://github.com/nsgo-dev/nsgo-pet/releases' },
			{ label: 'GitHub', href: 'https://github.com/nsgo-dev/nsgo-pet' },
		],
	},
};

const zh: typeof en = {
	htmlLang: 'zh-CN',
	title: 'nsgo-pet：macOS 和 Linux 的桌面宠物',
	description: '用 NSGo 构建的 macOS 和 Linux 桌面宠物：浮在窗口之上，向你打招呼、回应你的点击，夜里还会打盹。',
	nav: { label: '主导航', menu: '菜单', docs: '文档', pets: '宠物', github: 'GitHub', community: '社区', feedback: '反馈' },
	community: community({ nsgoNote: 'nsgo-pet 所用的 UI 引擎', telegram: 'NSGo Telegram 群组', awesomeNote: '应用、库和资源' }),
	theme: { name: '主题', light: '浅色', dark: '深色', auto: '自动' },
	language: { name: '语言', auto: '自动' },
	headline: ['住在你桌面上的', '小宠物。'],
	lead: 'nsgo-pet 浮在 macOS 和 Linux 的窗口之上。它会向你打招呼、回应你的点击、沿着屏幕底部走动，夜里还会打盹。',
	cta: { download: '免费下载', start: '入门指南' },
	meta: ['macOS', 'Linux', '11 只宠物'],
	buddy: 'NSGo 酱：点点她，或者拖着她走',
	pets: {
		eyebrow: '宠物',
		title: '十一只宠物，一键切换',
		text: '在托盘图标打开的设置里切换宠物，nsgo-pet 会记住你的选择。每只宠物都保留自己的许可证。',
		names: { 'nsgo-chan': 'NSGo 酱', bigbluefish: '大蓝鱼', 'openpets-halloween': 'OpenPets（万圣节）' },
		clips: (n: number) => `${n} 段动画`,
		nonCommercial: '禁止商用',
	},
	features: {
		eyebrow: '应用',
		title: '陪着你，却不挡着你',
		items: [
			{ icon: 'layers', title: '浮在最上层', text: '在透明窗口里浮于其他窗口之上，出现在每个桌面空间和全屏应用上，不占程序坞和任务栏。' },
			{ icon: 'mouse-pointer-click', title: '点击可穿透', text: '只有宠物身体部分接收指针，点在别处会落到下面的窗口。' },
			{ icon: 'calendar-clock', title: '知道现在几点', text: '按时段、季节、饭点、节日，以及前台是不是编辑器或终端来挑选要播放的事件。' },
			{ icon: 'footprints', title: '指针下活灵活现', text: '点击有反应，拖动时悬空晃动，还会沿着屏幕底部走动。' },
			{ icon: 'paw', title: '十一只宠物', text: '在托盘图标打开的设置里切换，并记住你的选择。' },
			{ icon: 'package-down', title: '自动更新', text: '从存储库的发行版安装更新，每天检查一次，也可在设置里检查。' },
		],
	},
	platforms: {
		eyebrow: '平台',
		title: '现已支持 macOS 和 Linux',
		status: { runs: '可用', later: '暂不支持' },
		items: [
			{ id: 'macos', name: 'macOS', status: 'runs', text: '播放带透明通道的 HEVC 动画。' },
			{ id: 'linux', name: 'Linux', status: 'runs', text: '播放 WebM 动画。在 X11 上可以浮于顶层；在 Wayland 上无法浮于顶层，也无法自己移动。' },
			{ id: 'windows', name: 'Windows', status: 'later', text: '等待 NSGo 的 Windows 播放器支持带透明通道的视频。' },
		],
	},
	end: {
		site: 'NSGo 官网',
		title: '用 NSGo 构建',
		text: 'nsgo-pet 是一个 NSGo 应用：透明窗口、托盘、菜单和动画播放都来自 NSGo，只用一套 Go 代码。NSGo 是以 Go 为后端的跨平台原生 UI 引擎。',
	},
	footer: {
		copyright: '© 2026 nsgo-dev',
		license: { label: 'MIT 许可证', href: 'https://github.com/nsgo-dev/nsgo-pet/blob/main/LICENSE' },
		links: [
			{ label: '文档', href: 'zh/docs/' },
			{ label: '反馈', href: 'https://github.com/nsgo-dev/nsgo-pet/issues' },
			{ label: '更新日志', href: 'https://github.com/nsgo-dev/nsgo-pet/releases' },
			{ label: 'GitHub', href: 'https://github.com/nsgo-dev/nsgo-pet' },
		],
	},
};

export const copy = { en, zh } satisfies Record<Lang, typeof en>;

export const docsHome = { en: 'docs/', zh: 'zh/docs/' } satisfies Record<Lang, string>;
export const homePath = { en: '', zh: 'zh/' } satisfies Record<Lang, string>;
