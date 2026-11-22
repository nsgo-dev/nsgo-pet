// The card a page shows when shared on X, Telegram and the like: Starlight
// gives each page its title, description and URL, and this adds the image
// scripts/share-cards.ts draws for the page's language, names the site in
// docs titles and marks the home pages as the site itself.
import { defineRouteMiddleware } from '@astrojs/starlight/route-data';
import { copy } from './copy';
import { project, type Lang } from './site';

type Head = App.Locals['starlightRoute']['head'];

export const onRequest = defineRouteMiddleware((context) => {
	const route = context.locals.starlightRoute;
	const lang: Lang = route.lang.startsWith('zh') ? 'zh' : 'en';
	const t = copy[lang];
	const base = import.meta.env.BASE_URL.replace(/\/$/, '');
	const image = new URL(`${base}/share/${lang}.png`, context.site).href;
	const alt = lang === 'zh' ? `${project.name}：${t.headline.join('')}` : `${project.name}: ${t.headline.join(' ')}`;
	const home = route.entry.data.title === t.title;

	const set = (key: 'property' | 'name', name: string, content: string) => {
		const meta = route.head.find((h) => h.tag === 'meta' && h.attrs?.[key] === name);
		if (meta) meta.attrs = { ...meta.attrs, content };
		else route.head.push({ tag: 'meta', attrs: { [key]: name, content } } as Head[number]);
	};
	if (!home) set('property', 'og:title', `${route.entry.data.title} | ${project.name}`);
	set('property', 'og:type', home ? 'website' : 'article');
	set('property', 'og:locale', lang === 'zh' ? 'zh_CN' : 'en_US');
	set('property', 'og:image', image);
	set('property', 'og:image:type', 'image/png');
	set('property', 'og:image:width', '1200');
	set('property', 'og:image:height', '630');
	set('property', 'og:image:alt', alt);
	set('name', 'twitter:card', 'summary_large_image');
	set('name', 'twitter:image', image);
	set('name', 'twitter:image:alt', alt);
});
