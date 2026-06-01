// @ts-check
import { defineConfig, envField } from 'astro/config';
import starlight from '@astrojs/starlight';
import remarkGfm from 'remark-gfm';

export default defineConfig({
	// Canonical production URL. Starlight renders nav links / canonical tags
	// absolute against this; if unset it leaks the dev host
	// (http://localhost:8080), which trips browser "local network access"
	// prompts on the deployed site. Keep it hardcoded to the prod domain.
	site: 'https://dado.atterpac.dev',
	markdown: {
		remarkPlugins: [remarkGfm],
	},
	env: {
		schema: {
			// Override to point at your VPS image server in production.
			// e.g. DADO_IMAGE_BASE=https://assets.example.com/dado/images
			DADO_IMAGE_BASE: envField.string({
				context: 'client',
				access: 'public',
				optional: true,
				default: '/images/components',
			}),
		},
	},
	integrations: [
		starlight({
			customCss: ['./src/styles/custom.css'],
			favicon: '/favicon.svg',
			title: 'dado',
			description: 'Terminal UI components for Go',
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/atterpac/dado' }],
			sidebar: [
				{
					label: 'Getting Started',
					items: [
						{ label: 'Introduction', slug: 'guides/introduction' },
						{ label: 'Installation', slug: 'guides/installation' },
						{ label: 'Quick Start', slug: 'guides/quick-start' },
					],
				},
				{
					label: 'Concepts',
					items: [
						{ label: 'Lifecycle', slug: 'concepts/lifecycle' },
						{ label: 'Views & Navigation', slug: 'concepts/navigation' },
						{ label: 'Themes', slug: 'concepts/themes' },
						{ label: 'Key Bindings', slug: 'concepts/key-bindings' },
						{ label: 'Events', slug: 'concepts/events' },
						{ label: 'Data Binding', slug: 'concepts/data-binding' },
					],
				},
				{
					label: 'Components',
					items: [{ autogenerate: { directory: 'components' } }],
				},
			],
		}),
	],
});
