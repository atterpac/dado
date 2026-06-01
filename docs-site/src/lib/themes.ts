import { readdir } from 'node:fs/promises';
import { resolve } from 'node:path';

// Base URL for component preview images. Override with DADO_IMAGE_BASE env var
// when deploying — the Go server serves images from a different path.
// e.g. DADO_IMAGE_BASE=https://assets.example.com/dado/images
export const IMAGE_BASE: string =
	import.meta.env.DADO_IMAGE_BASE ?? '/images/components';

// Returns the sorted list of theme names by reading the images directory.
// Called at build time — the images must be present in public/images/components/
// (local dev) or served from IMAGE_BASE (production).
export async function getAvailableThemes(): Promise<string[]> {
	const dir = resolve('public/images/components');
	try {
		const entries = await readdir(dir, { withFileTypes: true });
		return entries
			.filter(e => e.isDirectory())
			.map(e => e.name)
			.sort();
	} catch {
		// Images not yet copied — return the full known theme list as fallback.
		return [
			'catppuccin-frappe',
			'catppuccin-latte',
			'catppuccin-macchiato',
			'catppuccin-mocha',
			'dracula',
			'dracula-light',
			'everforest-dark',
			'everforest-light',
			'github-dark',
			'github-light',
			'gruvbox-dark',
			'gruvbox-light',
			'kanagawa',
			'monokai',
			'nord',
			'onedark',
			'onelight',
			'rosepine',
			'rosepine-dawn',
			'rosepine-moon',
			'solarized-dark',
			'solarized-light',
			'tokyonight-day',
			'tokyonight-moon',
			'tokyonight-night',
			'tokyonight-storm',
		];
	}
}
