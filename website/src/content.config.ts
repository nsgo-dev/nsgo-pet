import { defineCollection } from 'astro:content';
import { z } from 'astro/zod';
import { glob } from 'astro/loaders';
import { docsSchema } from '@astrojs/starlight/schema';

// The docs live with the project, in docs/public/<language>/, one folder per
// language; each is served under that language's docs path, so en/packs.md
// becomes docs/packs/ and zh/packs.md zh/docs/packs/.
const prefix: Record<string, string> = { en: 'docs', zh: 'zh/docs' };

export const collections = {
	docs: defineCollection({
		loader: glob({
			base: '../docs/public',
			pattern: '{en,zh}/**/[^_]*.{md,mdx}',
			generateId: ({ entry }) => {
				const [lang, ...rest] = entry.replace(/\.mdx?$/, '').split('/');
				const page = rest.join('/').replace(/(^|\/)index$/, '');
				return [prefix[lang], page].filter(Boolean).join('/');
			},
		}),
		// home: the home page, which lays out its own sections.
		schema: docsSchema({ extend: z.object({ home: z.boolean().optional() }) }),
	}),
};
