import { defineCollection } from 'astro:content';
import { z } from 'astro/zod';
import { docsLoader } from '@astrojs/starlight/loaders';
import { docsSchema } from '@astrojs/starlight/schema';

export const collections = {
	docs: defineCollection({
		loader: docsLoader(),
		// group: the page's sidebar group, which astro.config.mjs reads.
		// home: the home page, which lays out its own sections.
		schema: docsSchema({ extend: z.object({ group: z.string().optional(), home: z.boolean().optional() }) }),
	}),
};
