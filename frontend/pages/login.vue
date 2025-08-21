<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
      <div>
        <h1 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
          Login
        </h1>
        <p class="mt-2 text-center text-sm text-gray-600">
          Or
          <NuxtLink
            to="/register"
            class="font-medium text-indigo-600 hover:text-indigo-500 ml-1"
          >
            create a new account
          </NuxtLink>
        </p>
      </div>
      
      <form class="mt-8 space-y-6 bg-white shadow rounded-lg p-6" @submit.prevent="handleSubmit">
        <!-- Display server errors -->
        <div v-if="authStore.error" class="rounded-md bg-red-50 p-4">
          <div class="text-sm text-red-600">
            {{ authStore.error }}
          </div>
        </div>

        <div class="space-y-4">
          <!-- Email field -->
          <div>
            <label for="email" class="block text-sm font-medium text-gray-700">
              Email address
            </label>
            <input
              id="email"
              v-model="form.email"
              name="email"
              type="email"
              autocomplete="email"
              required
              class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
              :class="{ 'border-red-500': errors.email }"
              placeholder="Enter your email"
              @input="clearErrors"
            />
            <p v-if="errors.email" class="mt-1 text-sm text-red-600">
              {{ errors.email }}
            </p>
          </div>

          <!-- Password field -->
          <div>
            <label for="password" class="block text-sm font-medium text-gray-700">
              Password
            </label>
            <input
              id="password"
              v-model="form.password"
              name="password"
              type="password"
              autocomplete="current-password"
              required
              class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
              :class="{ 'border-red-500': errors.password }"
              placeholder="Enter your password"
              @input="clearErrors"
            />
            <p v-if="errors.password" class="mt-1 text-sm text-red-600">
              {{ errors.password }}
            </p>
          </div>
        </div>

        <div>
          <button
            type="submit"
            :disabled="authStore.isLoading"
            class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span v-if="authStore.isLoading">Signing in...</span>
            <span v-else>Sign in</span>
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import type { LoginRequest } from '../types'
import { useAuthStore } from '../stores/auth'

// Set up guest middleware to redirect authenticated users
definePageMeta({
  middleware: 'guest',
  layout: 'auth'
})

// Reactive form data
const form = reactive<LoginRequest>({
  email: '',
  password: ''
})

// Form validation errors
const errors = reactive({
  email: '',
  password: ''
})

// Auth store
const authStore = useAuthStore()

// Clear errors when user starts typing
const clearErrors = () => {
  errors.email = ''
  errors.password = ''
  authStore.clearError()
}

// Email validation function
const validateEmail = (email: string): boolean => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

// Form validation
const validateForm = (): boolean => {
  let isValid = true

  // Clear previous errors
  clearErrors()

  // Email validation
  if (!form.email.trim()) {
    errors.email = 'Email is required'
    isValid = false
  } else if (!validateEmail(form.email)) {
    errors.email = 'Please enter a valid email address'
    isValid = false
  }

  // Password validation
  if (!form.password) {
    errors.password = 'Password is required'
    isValid = false
  }

  return isValid
}

// Handle form submission
const handleSubmit = async () => {
  // Prevent submission during loading
  if (authStore.isLoading) {
    return
  }

  // Validate form
  if (!validateForm()) {
    return
  }

  try {
    await authStore.login(form)
    // Navigation is handled in the store
  } catch (error) {
    // Error handling is managed by the store
    console.error('Login failed:', error)
  }
}

// Set page title
useHead({
  title: 'Login - Task Tracker'
})
</script>