import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import vueDevTools from 'vite-plugin-vue-devtools'

import fs from 'node:fs'
import path from 'node:path'

// process.env.VITE_APP_VERSION = require("./package.json").version;

// https://vite.dev/config/
export default defineConfig({

  plugins: [
    vue(),
    vueJsx(),
    vueDevTools(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },

  server: {
    // https: {
    //   key: fs.readFileSync(path.resolve('.cert/dev.key')),
    //   cert: fs.readFileSync(path.resolve('.cert/dev.crt')),
    // },
    proxy: {
      '/api/v1': {
        target: 'https://localhost:6443',
        changeOrigin: true,

        // dev-only: allows Vite's Node proxy to talk to your
        // self-signed backend cert
        secure: false,
      },
    },
  },

})
