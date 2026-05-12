// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  srcDir: 'app',
  css: ['~/assets/css/main.css'],
  modules: [
    '@nuxt/eslint',
    '@nuxt/ui',
    '@pinia/nuxt'
  ],
  routeRules: {
    '/': { ssr: false }
  },
  future: {
    compatibilityVersion: 4
  },

  devtools: {
    enabled: true
  },
  
  devServer: {
    port: 9245,
    host: '127.0.0.1',
    loadingScreen: false
  },

  compatibilityDate: '2025-01-15',

  eslint: {
    config: {
      stylistic: {
        commaDangle: 'never',
        braceStyle: '1tbs'
      }
    }
  }
})
