import { defineConfig } from 'astro/config';
import react from '@astrojs/react';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const root = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  site: 'https://lexwebb.github.io',
  base: '/moto-hud/',
  trailingSlash: 'always',
  output: 'static',
  integrations: [react()],
  vite: {
    resolve: {
      alias: {
        '@design': path.resolve(root, '../design'),
      },
    },
    server: {
      fs: {
        allow: [root, path.resolve(root, '..')],
      },
    },
  },
});
