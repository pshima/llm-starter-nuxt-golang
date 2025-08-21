import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { navigateTo } from '#app'
import type { User, LoginRequest, RegisterRequest } from '../types'
import { useApi } from '../composables/useApi'

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const isAuthenticated = computed(() => !!user.value)

  // Actions
  const { apiCall } = useApi()

  const login = async (credentials: LoginRequest) => {
    try {
      isLoading.value = true
      error.value = null
      
      const userData = await apiCall<User>('/auth/login', {
        method: 'POST',
        body: credentials,
      })
      
      user.value = userData
      await navigateTo('/dashboard')
    } catch (err: any) {
      error.value = err.message || 'Login failed'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const register = async (userData: RegisterRequest) => {
    try {
      isLoading.value = true
      error.value = null
      
      console.log('📝 Starting registration...')
      const newUser = await apiCall<User>('/auth/register', {
        method: 'POST',
        body: userData,
      })
      
      console.log('✅ Registration API success:', newUser)
      user.value = newUser
      console.log('👤 User set:', user.value)
      console.log('🔐 Is authenticated:', isAuthenticated.value)
      console.log('🏠 Navigating to dashboard...')
      await navigateTo('/dashboard')
    } catch (err: any) {
      console.error('❌ Registration failed:', err)
      error.value = err.message || 'Registration failed'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const logout = async () => {
    try {
      isLoading.value = true
      error.value = null
      
      await apiCall('/auth/logout', {
        method: 'POST',
      })
      
      user.value = null
      await navigateTo('/login')
    } catch (err: any) {
      // Even if logout fails on server, clear local state
      user.value = null
      error.value = err.message || 'Logout failed'
      await navigateTo('/login')
    } finally {
      isLoading.value = false
    }
  }

  const getCurrentUser = async () => {
    try {
      isLoading.value = true
      error.value = null
      
      const userData = await apiCall<User>('/auth/me')
      user.value = userData
    } catch (err: any) {
      user.value = null
      error.value = err.message || 'Failed to get current user'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const clearError = () => {
    error.value = null
  }

  return {
    // State
    user: readonly(user),
    isLoading: readonly(isLoading),
    error: readonly(error),
    
    // Getters
    isAuthenticated,
    
    // Actions
    login,
    register,
    logout,
    getCurrentUser,
    clearError,
  }
})