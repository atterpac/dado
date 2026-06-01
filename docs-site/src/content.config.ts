import { defineCollection, z } from 'astro:content';
import { docsLoader } from '@astrojs/starlight/loaders';
import { docsSchema } from '@astrojs/starlight/schema';

const methodDoc = z.object({
	name: z.string(),
	signature: z.string(),
	doc: z.string(),
});

const fieldDoc = z.object({
	name: z.string(),
	type: z.string(),
	doc: z.string(),
});

const exampleDoc = z.object({
	name: z.string(),
	doc: z.string(),
	code: z.string(),
});

const typeDoc = z.object({
	name: z.string(),
	doc: z.string(),
	kind: z.enum(['struct', 'enum', '']).optional(),
	fields: z.array(fieldDoc).optional(),
	values: z.array(z.object({ name: z.string(), doc: z.string() })).optional(),
});

export const collections = {
	docs: defineCollection({ loader: docsLoader(), schema: docsSchema() }),

	api: defineCollection({
		type: 'data',
		schema: z.object({
			name: z.string(),
			slug: z.string(),
			category: z.enum(['component', 'event', 'base', 'data']),
			doc: z.string(),
			constructor: methodDoc.optional(),
			methods: z.array(methodDoc).nullable(),
			types: z.array(typeDoc).optional(),
			examples: z.array(exampleDoc).optional(),
		}),
	}),
};
