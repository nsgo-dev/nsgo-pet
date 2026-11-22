import type { Lang } from './site';

/** The site's language of a Starlight route's BCP-47 tag. */
export function langOf(tag: string): Lang {
	return tag.startsWith('zh') ? 'zh' : 'en';
}

/** The same page's path in the other language: Chinese lives under zh/. */
export function otherLanguagePath(pathname: string): string {
	const base = import.meta.env.BASE_URL.replace(/\/$/, '');
	const path = pathname.slice(base.length).replace(/^\//, '');
	return `${base}/${path.startsWith('zh/') ? path.slice(3) : `zh/${path}`}`;
}
