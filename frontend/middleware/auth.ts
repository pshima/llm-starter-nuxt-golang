/**
 * Auth middleware to protect routes that require authentication
 * Redirects unauthenticated users to the login page
 */

export default defineNuxtRouteMiddleware(async (to) => {
  const authStore = useAuthStore()
  
  // If user is already authenticated, continue
  if (authStore.isAuthenticated) {
    return
  }
  
  // If not authenticated, try to restore session from cookie
  // This handles page refreshes and direct navigation to protected routes
  try {
    await authStore.getCurrentUser()
    
    // If session restoration was successful, user is now authenticated
    if (authStore.isAuthenticated) {
      return
    }
  } catch (error) {
    // Session restoration failed, user is not authenticated
    console.log('Session restoration failed:', error)
  }
  
  // User is not authenticated, redirect to login
  return navigateTo('/login')
})