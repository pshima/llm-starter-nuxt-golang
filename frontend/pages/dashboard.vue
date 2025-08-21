<template>
  <div class="min-h-screen">
    <!-- Hero Section -->
    <div 
      data-testid="hero-section"
      class="bg-gradient-to-br from-indigo-500 via-purple-600 to-pink-500 px-4 py-16 sm:px-6 lg:px-8"
    >
      <div class="mx-auto max-w-7xl text-center">
        <h1 class="text-4xl font-bold tracking-tight text-white sm:text-6xl">
          Welcome back, {{ displayName }}!
        </h1>
        <p class="mt-6 text-lg leading-8 text-indigo-100">
          Your Task Management Dashboard
        </p>
        <p class="mt-2 text-base text-indigo-200">
          Stay organized and productive with your personal task tracker
        </p>
      </div>
    </div>

    <!-- Quick Stats Section (only if user has tasks) -->
    <div v-if="tasksStore.totalTasks > 0" class="bg-white py-12">
      <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div class="text-center mb-8">
          <h2 class="text-2xl font-bold text-gray-900">Quick Stats</h2>
        </div>
        <div class="grid grid-cols-1 gap-8 sm:grid-cols-3">
          <div class="text-center">
            <div class="text-3xl font-bold text-indigo-600">{{ tasksStore.totalTasks }}</div>
            <div class="text-sm text-gray-600">Total Tasks</div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-green-600">{{ tasksStore.completedTasks.length }}</div>
            <div class="text-sm text-gray-600">Completed Tasks</div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-purple-600">{{ completionPercentage }}%</div>
            <div class="text-sm text-gray-600">Completion Rate</div>
          </div>
        </div>
        <div class="mt-8 text-center">
          <div class="text-lg font-medium text-gray-700">
            Categories: {{ tasksStore.categories.length }}
          </div>
        </div>
      </div>
    </div>

    <!-- Feature Cards -->
    <div class="bg-gray-50 py-16">
      <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div class="text-center mb-12">
          <h2 class="text-3xl font-bold text-gray-900">Task Tracker Features</h2>
          <p class="mt-4 text-lg text-gray-600">
            Everything you need to stay organized and productive
          </p>
        </div>
        
        <div 
          data-testid="features-grid"
          class="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-4"
        >
          <!-- Quick Task Creation -->
          <div class="bg-white rounded-lg shadow-lg p-6 hover:shadow-xl transition-shadow">
            <div class="h-12 w-12 bg-indigo-500 rounded-lg flex items-center justify-center mb-4">
              <svg class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900 mb-2">Quick Task Creation</h3>
            <p class="text-gray-600">Create and manage tasks with ease using our intuitive interface.</p>
          </div>

          <!-- Smart Organization -->
          <div class="bg-white rounded-lg shadow-lg p-6 hover:shadow-xl transition-shadow">
            <div class="h-12 w-12 bg-green-500 rounded-lg flex items-center justify-center mb-4">
              <svg class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9.568 3H5.25A2.25 2.25 0 003 5.25v4.318c0 .597.237 1.17.659 1.591l9.581 9.581c.699.699 1.78.872 2.607.33a18.095 18.095 0 005.223-5.223c.542-.827.369-1.908-.33-2.607L11.16 3.66A2.25 2.25 0 009.568 3z" />
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900 mb-2">Smart Organization</h3>
            <p class="text-gray-600">Organize tasks into custom categories for better workflow management.</p>
          </div>

          <!-- Progress Tracking -->
          <div class="bg-white rounded-lg shadow-lg p-6 hover:shadow-xl transition-shadow">
            <div class="h-12 w-12 bg-purple-500 rounded-lg flex items-center justify-center mb-4">
              <svg class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" />
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900 mb-2">Progress Tracking</h3>
            <p class="text-gray-600">Track your progress and completion rates to stay motivated and focused.</p>
          </div>

          <!-- Safe Deletion -->
          <div class="bg-white rounded-lg shadow-lg p-6 hover:shadow-xl transition-shadow">
            <div class="h-12 w-12 bg-orange-500 rounded-lg flex items-center justify-center mb-4">
              <svg class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.623 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z" />
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900 mb-2">Safe Deletion</h3>
            <p class="text-gray-600">Safely delete tasks with 7-day recovery window. Never lose important work.</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Call to Action -->
    <div class="bg-white py-16">
      <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 text-center">
        <h2 class="text-3xl font-bold text-gray-900 mb-4">Ready to Get Started?</h2>
        <p class="text-lg text-gray-600 mb-8">
          Jump into your task management and start organizing your work today.
        </p>
        <button
          @click="navigateTo('/tasks')"
          data-testid="goto-tasks-button"
          class="inline-flex items-center justify-center rounded-md bg-indigo-600 px-8 py-3 text-base font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 transition-colors"
        >
          View Your Tasks
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'

// Set up auth middleware to protect this route
definePageMeta({
  middleware: 'auth'
})

// Stores
const authStore = useAuthStore()
const tasksStore = useTasksStore()

// Computed properties
const displayName = computed(() => {
  return authStore.user?.displayName || authStore.user?.email || 'User'
})

const completionPercentage = computed(() => {
  if (tasksStore.totalTasks === 0) return 0
  return Math.round((tasksStore.completedTasks.length / tasksStore.totalTasks) * 100)
})

// Load data on mount
onMounted(async () => {
  await Promise.all([
    tasksStore.loadTasks(),
    tasksStore.loadCategories()
  ])
})

// Set page title
useHead({
  title: 'Dashboard - Task Tracker'
})
</script>