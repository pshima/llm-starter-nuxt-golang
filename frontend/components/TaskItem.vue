<template>
  <div
    data-testid="task-container"
    :class="[
      'bg-white rounded-lg shadow-sm border transition-all duration-200 hover:shadow-md',
      task.deleted_at ? 'border-red-200 bg-red-50' : 'border-gray-200 hover:border-gray-300'
    ]"
  >
    <div class="p-4">
      <!-- Task Header with Checkbox and Category -->
      <div class="flex items-start justify-between mb-3">
        <div class="flex items-center gap-3">
          <input
            :id="`task-${task.id}`"
            type="checkbox"
            :checked="task.completed"
            @change="handleToggleCompletion"
            @click.stop="handleToggleCompletion"
            data-testid="task-checkbox"
            class="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded cursor-pointer"
          />
          <span
            data-testid="task-category"
            class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800"
          >
            {{ task.category }}
          </span>
        </div>

        <!-- Action Buttons -->
        <div class="flex items-center gap-2">
          <button
            v-if="task.deleted_at"
            @click.stop="$emit('restore-task', task.id)"
            data-testid="restore-button"
            :aria-label="`Restore task: ${task.description}`"
            class="p-1.5 text-green-600 hover:text-green-800 hover:bg-green-100 rounded-full transition-colors"
          >
            <ArrowUturnLeftIcon class="h-4 w-4" />
          </button>
          <button
            v-else
            @click.stop="$emit('delete-task', task.id)"
            data-testid="delete-button"
            :aria-label="`Delete task: ${task.description}`"
            class="p-1.5 text-red-600 hover:text-red-800 hover:bg-red-100 rounded-full transition-colors"
          >
            <TrashIcon class="h-4 w-4" />
          </button>
        </div>
      </div>

      <!-- Task Content (Clickable for editing) -->
      <div
        @click="$emit('edit-task', task.id)"
        data-testid="task-content"
        class="cursor-pointer group"
      >
        <!-- Task Description -->
        <p
          data-testid="task-description"
          :class="[
            'text-gray-900 text-sm font-medium group-hover:text-indigo-600 transition-colors',
            task.completed ? 'line-through text-gray-500' : ''
          ]"
        >
          {{ task.description }}
        </p>

        <!-- Task Metadata -->
        <div class="mt-2 flex items-center justify-between text-xs text-gray-500">
          <span>Created {{ formatDate(task.created_at) }}</span>
          <span v-if="task.updated_at !== task.created_at">
            Updated {{ formatDate(task.updated_at) }}
          </span>
        </div>
        
        <!-- Deleted Indicator -->
        <div v-if="task.deleted_at" class="mt-2 flex items-center text-xs text-red-600">
          <ExclamationTriangleIcon class="h-3 w-3 mr-1" />
          Deleted {{ formatDate(task.deleted_at) }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { TrashIcon, ArrowUturnLeftIcon, ExclamationTriangleIcon } from '@heroicons/vue/24/outline'
import type { Task } from '~/types'

interface Props {
  task: Task
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'toggle-completion': [taskId: string]
  'edit-task': [taskId: string]
  'delete-task': [taskId: string]
  'restore-task': [taskId: string]
}>()

const handleToggleCompletion = (event: Event) => {
  const target = event.target as HTMLInputElement
  // Emit the event - the parent will handle the actual API call
  emit('toggle-completion', props.task.id)
}

const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  const now = new Date()
  const diffInHours = (now.getTime() - date.getTime()) / (1000 * 60 * 60)
  
  if (diffInHours < 24) {
    return date.toLocaleTimeString('en-US', { 
      hour: 'numeric', 
      minute: '2-digit',
      hour12: true 
    })
  } else if (diffInHours < 24 * 7) {
    return date.toLocaleDateString('en-US', { 
      weekday: 'short',
      month: 'short',
      day: 'numeric'
    })
  } else {
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined
    })
  }
}
</script>