<template>
  <div>
    <!-- Loading State -->
    <div
      v-if="loading"
      data-testid="loading-state"
      class="flex flex-col items-center justify-center py-12 text-center"
    >
      <div
        data-testid="loading-spinner"
        class="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600 mb-4"
      ></div>
      <p class="text-gray-600 text-lg">Loading tasks...</p>
    </div>

    <!-- Empty State -->
    <div
      v-else-if="!tasks || tasks.length === 0"
      data-testid="empty-state"
      class="flex flex-col items-center justify-center py-12 text-center"
    >
      <div class="mx-auto max-w-md">
        <svg
          class="mx-auto h-16 w-16 text-gray-400 mb-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke-width="1.5"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M9 12h3.75M9 15h3.75M9 18h3.75m3-9v12c0 1.1-.9 2-2 2H3c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2h15c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2h-1M3 6h18v4H3V6z"
          />
        </svg>
        <h3 class="text-lg font-medium text-gray-900 mb-2">No tasks found</h3>
        <p class="text-gray-600">
          Create your first task to get started with organizing your work.
        </p>
      </div>
    </div>

    <!-- Tasks Grid -->
    <div
      v-else
      data-testid="tasks-container"
      class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4"
    >
      <TaskItem
        v-for="task in tasks"
        :key="task.id"
        :task="task"
        @toggle-completion="$emit('toggle-completion', $event)"
        @edit-task="$emit('edit-task', $event)"
        @delete-task="$emit('delete-task', $event)"
        @restore-task="$emit('restore-task', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import TaskItem from './TaskItem.vue'
import type { Task } from '~/types'

interface Props {
  tasks: Task[]
  loading: boolean
}

defineProps<Props>()

defineEmits<{
  'toggle-completion': [taskId: string]
  'edit-task': [taskId: string]
  'delete-task': [taskId: string]
  'restore-task': [taskId: string]
}>()
</script>