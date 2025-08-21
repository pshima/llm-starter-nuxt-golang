<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Main Content -->
    <main class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <!-- Header Section -->
      <div class="mb-8">
        <div class="flex items-center justify-between">
          <div>
            <h1 class="text-3xl font-bold text-gray-900">Tasks</h1>
            <div class="mt-1 flex items-center gap-4 text-sm text-gray-600">
              <span>{{ filteredTasks.filter(t => !t.completed && !t.deleted_at).length }} active</span>
              <span v-if="showCompleted">{{ filteredTasks.filter(t => t.completed && !t.deleted_at).length }} completed</span>
              <span v-if="showDeleted">{{ filteredTasks.filter(t => !!t.deleted_at).length }} deleted</span>
            </div>
          </div>
          <button
            @click="openCreateModal"
            data-testid="new-task-button"
            class="inline-flex items-center justify-center rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 transition-colors"
          >
            <PlusIcon class="h-5 w-5 mr-2" />
            New Task
          </button>
        </div>

        <!-- Error Message -->
        <div
          v-if="tasksStore.error"
          data-testid="error-message"
          class="mt-4 rounded-md bg-red-50 p-4"
        >
          <div class="flex">
            <div class="flex-shrink-0">
              <ExclamationTriangleIcon class="h-5 w-5 text-red-400" />
            </div>
            <div class="ml-3">
              <h3 class="text-sm font-medium text-red-800">Error</h3>
              <p class="mt-1 text-sm text-red-700">{{ tasksStore.error }}</p>
            </div>
            <div class="ml-auto pl-3">
              <div class="-mx-1.5 -my-1.5">
                <button
                  @click="tasksStore.clearError"
                  data-testid="dismiss-error"
                  class="inline-flex rounded-md bg-red-50 p-1.5 text-red-500 hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-red-600 focus:ring-offset-2 focus:ring-offset-red-50"
                >
                  <XMarkIcon class="h-5 w-5" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Filters -->
      <TaskFilters
        :categories="tasksStore.categories"
        :selected-category="selectedCategory"
        :show-completed="showCompleted"
        :show-deleted="showDeleted"
        @category-change="handleCategoryChange"
        @toggle-completed="handleToggleCompleted"
        @toggle-deleted="handleToggleDeleted"
      />

      <!-- Task List -->
      <TaskList
        :tasks="filteredTasks"
        :loading="tasksStore.isLoading"
        @toggle-completion="handleToggleCompletion"
        @edit-task="handleEditTask"
        @delete-task="handleDeleteTask"
        @restore-task="handleRestoreTask"
      />

      <!-- Task Modal -->
      <TaskModal
        :is-open="isModalOpen"
        :task="selectedTask"
        :categories="tasksStore.categories"
        :loading="tasksStore.isLoading"
        @save="handleSaveTask"
        @close="closeModal"
      />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { PlusIcon, ExclamationTriangleIcon, XMarkIcon } from '@heroicons/vue/24/outline'
import TaskFilters from '../components/TaskFilters.vue'
import TaskList from '../components/TaskList.vue'
import TaskModal from '../components/TaskModal.vue'
import { useTasksStore } from '../stores/tasks'
import type { Task } from '../types'

// Set up auth middleware to protect this route
definePageMeta({
  middleware: 'auth'
})

// Set page title
useHead({
  title: 'Tasks - Task Tracker'
})

// Stores
const tasksStore = useTasksStore()

// State
const selectedCategory = ref('All')
const showCompleted = ref(true)
const showDeleted = ref(false)
const isModalOpen = ref(false)
const selectedTask = ref<Task | null>(null)

// Computed
const filteredTasks = computed(() => {
  let tasks = [...tasksStore.tasks]

  // Filter by category
  if (selectedCategory.value !== 'All') {
    tasks = tasks.filter(task => task.category === selectedCategory.value)
  }

  // Filter by completion status
  if (!showCompleted.value) {
    tasks = tasks.filter(task => !task.completed)
  }

  // Filter by deleted status
  if (!showDeleted.value) {
    tasks = tasks.filter(task => !task.deleted_at)
  } else {
    // When showing deleted tasks, include both deleted and non-deleted
    // (the user can choose to hide completed tasks separately)
  }

  return tasks
})

// Event handlers
const handleCategoryChange = (category: string) => {
  selectedCategory.value = category
}

const handleToggleCompleted = (show: boolean) => {
  showCompleted.value = show
}

const handleToggleDeleted = (show: boolean) => {
  showDeleted.value = show
}

const openCreateModal = () => {
  selectedTask.value = null
  isModalOpen.value = true
}

const handleEditTask = (taskId: string) => {
  const task = tasksStore.tasks.find(t => t.id === taskId)
  if (task) {
    selectedTask.value = task
    isModalOpen.value = true
  }
}

const closeModal = () => {
  isModalOpen.value = false
  selectedTask.value = null
}

const handleSaveTask = async (data: { description: string; category: string }) => {
  console.log('🔧 handleSaveTask called with:', data)
  try {
    if (selectedTask.value) {
      // Editing existing task
      console.log('📝 Editing existing task:', selectedTask.value.id)
      await tasksStore.updateTask(selectedTask.value.id, data)
    } else {
      // Creating new task
      console.log('📝 Creating new task')
      await tasksStore.createTask(data)
    }
    // Close modal after successful save
    closeModal()
  } catch (error) {
    console.error('❌ Save task error:', error)
    // Error is handled by the store
  }
}

const handleToggleCompletion = async (taskId: string) => {
  try {
    await tasksStore.toggleTaskCompletion(taskId)
  } catch (error) {
    // Error is handled by the store
  }
}

const handleDeleteTask = async (taskId: string) => {
  try {
    await tasksStore.deleteTask(taskId)
  } catch (error) {
    // Error is handled by the store
  }
}

const handleRestoreTask = async (taskId: string) => {
  try {
    await tasksStore.restoreTask(taskId)
  } catch (error) {
    // Error is handled by the store
  }
}

// Lifecycle
onMounted(async () => {
  try {
    await Promise.all([
      tasksStore.fetchTasks(),
      tasksStore.fetchCategories()
    ])
  } catch (error) {
    // Error is handled by the store
  }
})
</script>