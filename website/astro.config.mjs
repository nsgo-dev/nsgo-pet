import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// GitHub Pages serves the site under the repository's name; a deploy may
// pass the address it was given in SITE and BASE, and a local build uses the
// same path so that links are checked under a subpath too.
const site = process.env.SITE ?? 'https://nsgo-dev.github.io';
const base = process.env.BASE ?? '/nsgo-pet';

export default defineConfig({
	site,
	base,
	trailingSlash: 'always',
	devToolbar: { enabled: false },
	integrations: [
		starlight({
			title: 'nsgo-pet',
			description: 'A desktop pet for macOS and Linux built with NSGo, and the pet engine behind it.',
			logo: { src: './src/assets/logo.png' },
			favicon: '/favicon.png',
			defaultLocale: 'root',
			locales: {
				root: { label: 'English', lang: 'en' },
				zh: { label: '简体中文', lang: 'zh-CN' },
			},
			social: [
				{ icon: 'github', label: 'GitHub', href: 'https://github.com/nsgo-dev/nsgo-pet' },
			],
			// Every docs page in its order; src/routeData.ts groups them by
			// their front matter as each page renders.
			sidebar: [{ autogenerate: { directory: 'docs' } }],
			routeMiddleware: ['./src/routeData.ts', './src/share.ts'],
			// Code as the home page shows it: the same themes in a soft frame.
			// Its colors are set in docs.css, from the site's own.
			expressiveCode: {
				themes: ['github-dark-dimmed', 'github-light'],
				styleOverrides: {
					borderRadius: '1rem',
					codeFontSize: '0.8125rem',
					codeLineHeight: '1.7',
					codePaddingBlock: '1rem',
					codePaddingInline: '1.25rem',
					frames: { shadowColor: 'transparent' },
				},
			},
			customCss: ['./src/styles/tokens.css', './src/styles/shell.css', './src/styles/docs.css'],
			components: {
				Head: './src/components/starlight/Head.astro',
				Header: './src/components/shell/SiteHeader.astro',
				PageFrame: './src/components/starlight/PageFrame.astro',
				TwoColumnContent: './src/components/starlight/TwoColumnContent.astro',
				MobileMenuToggle: './src/components/starlight/Empty.astro',
				ContentPanel: './src/components/starlight/ContentPanel.astro',
				MarkdownContent: './src/components/starlight/MarkdownContent.astro',
				PageTitle: './src/components/starlight/PageTitle.astro',
				Footer: './src/components/starlight/Footer.astro',
				EditLink: './src/components/starlight/EditLink.astro',
			},
			head: [
				{ tag: 'link', attrs: { rel: 'apple-touch-icon', href: `${base}/apple-touch-icon.png` } },
			],
		}),
	],
});
