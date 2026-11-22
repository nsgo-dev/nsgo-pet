// Groups the docs' sidebar by each page's `group` front matter, as each
// page renders, so a page's group shows as soon as it is written, in the
// dev server too. Pages without a group come first, then each group in the
// order of its first page; Chinese pages give their groups' Chinese names.
import { defineRouteMiddleware } from '@astrojs/starlight/route-data';
import { getCollection } from 'astro:content';

type Entries = App.Locals['starlightRoute']['sidebar'];
type Link = Extract<Entries[number], { type: 'link' }>;
type Group = Extract<Entries[number], { type: 'group' }>;

const flatten = (entries: Entries): Link[] => entries.flatMap((e) => (e.type === 'link' ? [e] : flatten(e.entries)));

export const onRequest = defineRouteMiddleware(async (context) => {
	const route = context.locals.starlightRoute;
	if (!route.hasSidebar) return;
	const base = import.meta.env.BASE_URL.replace(/\/$/, '');
	const groups = new Map((await getCollection('docs')).map((e) => [e.id, (e.data as { group?: string }).group]));
	const idOf = (href: string) => decodeURI(href).slice(base.length).replace(/^\/|\/$/g, '');

	const sidebar: Entries = [];
	const byName = new Map<string, Group>();
	for (const link of flatten(route.sidebar)) {
		const name = groups.get(idOf(link.href));
		if (!name) {
			sidebar.push(link);
			continue;
		}
		let group = byName.get(name);
		if (!group) {
			group = { type: 'group', label: name, entries: [], collapsed: false, badge: undefined };
			byName.set(name, group);
			sidebar.push(group);
		}
		group.entries.push(link);
	}
	route.sidebar = sidebar;
});
