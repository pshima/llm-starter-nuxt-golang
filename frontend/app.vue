<template>
  <div id="app">
    <NuxtRouteAnnouncer />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from './stores/auth'

// Global app setup
useHead({
  title: 'Task Tracker',
  meta: [
    { name: 'description', content: 'A simple and intuitive task management system' },
    { name: 'viewport', content: 'width=device-width, initial-scale=1' }
  ]
})

// Initialize auth on app start
const authStore = useAuthStore()

// Initialize auth state on app start
// Note: Session restoration is now handled by the auth middleware
// to avoid race conditions with route protection
onMounted(async () => {
  // Only restore session if not already authenticated and not on auth pages
  const route = useRoute()
  const isAuthPage = ['/login', '/register', '/'].includes(route.path)
  
  if (!authStore.isAuthenticated && !isAuthPage) {
    try {
      await authStore.getCurrentUser()
    } catch (error) {
      // User not authenticated, that's okay
      console.log('User not authenticated on app initialization')
    }
  }
})
</script>
