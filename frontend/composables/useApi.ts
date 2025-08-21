import type { ApiResponse, ApiError } from '../types'

// API configuration composable
export const useApi = () => {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase as string

  // Generic API call function with proper error handling
  const apiCall = async <T>(
    endpoint: string,
    options: any = {}
  ): Promise<T> => {
    try {
      const response = await $fetch<any>(`${baseURL}${endpoint}`, {
        credentials: 'include', // Important for session cookies
        ...options,
      })

      // Handle API-level errors
      if (response.error) {
        throw new Error(response.error)
      }

      // Check if response is wrapped (has data property) or direct
      if (response.data !== undefined) {
        return response.data as T
      } else {
        // Direct response (like auth endpoints)
        return response as T
      }
    } catch (error: any) {
      // Handle network errors and API errors
      if (error.data?.error) {
        throw new Error(error.data.error)
      }
      throw error
    }
  }

  return {
    apiCall,
    baseURL,
  }
}