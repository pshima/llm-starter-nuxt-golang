/**
 * Guest middleware to redirect authenticated users away from auth pages
 * Redirects authenticated users to the dashboard
 */

export default defineNuxtRouteMiddleware((to) => {
  const authStore = useAuthStore()
  
  // If user is already authenticated, redirect to dashboard
  if (authStore.isAuthenticated) {
    return navigateTo('/dashboard')
  }
})