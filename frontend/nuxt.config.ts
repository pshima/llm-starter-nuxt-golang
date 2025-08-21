// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  
  // Modules
  modules: [
    '@nuxtjs/tailwindcss',
    '@pinia/nuxt'
  ],
  
  // CSS files
  css: ['~/assets/css/tailwind.css'],
  
  // Configure for SPA mode
  ssr: false,
  
  // Development server configuration
  devServer: {
    port: 3000,
    host: '0.0.0.0'
  },
  
  // TypeScript configuration
  typescript: {
    typeCheck: false // Disable for development, auto-imports work but vue-tsc has issues
  },
  
  // Auto-imports configuration
  imports: {
    dirs: ['stores']
  },
  
  // Runtime config for API configuration
  runtimeConfig: {
    public: {
      // Use environment variable for API base URL, fallback to localhost for development
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080/api/v1'
    }
  }
})
