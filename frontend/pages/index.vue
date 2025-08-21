<template>
  <div class="text-center">
    <div class="max-w-2xl mx-auto">
      <h1 class="text-4xl font-bold text-gray-900 mb-6">
        Welcome to Task Tracker
      </h1>
      
      <p class="text-lg text-gray-600 mb-8">
        A simple and intuitive task management system to help you stay organized and productive.
      </p>
      
      <div class="space-x-4">
        <NuxtLink 
          to="/login"
          class="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          Sign In
        </NuxtLink>
        
        <NuxtLink 
          to="/register"
          class="inline-flex items-center px-6 py-3 border border-gray-300 text-base font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          Sign Up
        </NuxtLink>
      </div>
      
      <div class="mt-12">
        <h2 class="text-2xl font-bold text-gray-900 mb-6">Features</h2>
        
        <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div class="text-center">
            <div class="w-12 h-12 mx-auto mb-4 bg-blue-100 rounded-lg flex items-center justify-center">
              <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"></path>
              </svg>
            </div>
            <h3 class="text-lg font-medium text-gray-900 mb-2">Task Management</h3>
            <p class="text-gray-600">Create, organize, and track your tasks with ease.</p>
          </div>
          
          <div class="text-center">
            <div class="w-12 h-12 mx-auto mb-4 bg-green-100 rounded-lg flex items-center justify-center">
              <svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path>
              </svg>
            </div>
            <h3 class="text-lg font-medium text-gray-900 mb-2">Categories</h3>
            <p class="text-gray-600">Organize tasks into custom categories for better structure.</p>
          </div>
          
          <div class="text-center">
            <div class="w-12 h-12 mx-auto mb-4 bg-purple-100 rounded-lg flex items-center justify-center">
              <svg class="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
              </svg>
            </div>
            <h3 class="text-lg font-medium text-gray-900 mb-2">Recovery</h3>
            <p class="text-gray-600">Accidentally deleted tasks? Recover them within 7 days.</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { watchEffect } from 'vue'
import { navigateTo } from '#app'
import { useAuthStore } from '../stores/auth'

// Set page meta
definePageMeta({
  layout: 'auth'
})

useHead({
  title: 'Task Tracker - Simple Task Management',
  meta: [
    { name: 'description', content: 'A simple and intuitive task management system to help you stay organized and productive.' }
  ]
})

// Redirect to dashboard if already authenticated
const authStore = useAuthStore()

watchEffect(() => {
  if (authStore.isAuthenticated) {
    navigateTo('/dashboard')
  }
})
</script>