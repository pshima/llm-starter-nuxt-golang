<template>
  <div data-testid="filters-container" class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 bg-white rounded-lg shadow-sm p-4 mb-6">
    <!-- Category Filter -->
    <div class="flex items-center gap-3">
      <label class="text-sm font-medium text-gray-700">Category:</label>
      <Listbox
        :model-value="selectedCategory"
        @update:model-value="$emit('category-change', $event)"
        data-testid="category-filter"
      >
        <div class="relative">
          <ListboxButton
            class="relative cursor-pointer rounded-lg bg-white py-2 pl-3 pr-10 text-left shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 sm:text-sm min-w-[200px]"
          >
            <span class="block truncate">
              {{ selectedCategory === 'All' ? 'All Categories' : selectedCategory }}
            </span>
            <span class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
              <ChevronUpDownIcon class="h-5 w-5 text-gray-400" aria-hidden="true" />
            </span>
          </ListboxButton>

          <transition
            leave-active-class="transition duration-100 ease-in"
            leave-from-class="opacity-100"
            leave-to-class="opacity-0"
          >
            <ListboxOptions class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
              <ListboxOption
                value="All"
                data-testid="category-option"
                class="cursor-pointer select-none relative py-2 pl-10 pr-4 hover:bg-indigo-600 hover:text-white"
                :class="selectedCategory === 'All' ? 'bg-indigo-600 text-white' : 'text-gray-900'"
              >
                <span class="block truncate font-medium">All Categories</span>
                <span v-if="selectedCategory === 'All'" class="absolute inset-y-0 left-0 flex items-center pl-3 text-current">
                  <CheckIcon class="h-5 w-5" aria-hidden="true" />
                </span>
              </ListboxOption>
              
              <ListboxOption
                v-for="category in categories"
                :key="category.name"
                :value="category.name"
                data-testid="category-option"
                class="cursor-pointer select-none relative py-2 pl-10 pr-4 hover:bg-indigo-600 hover:text-white"
                :class="selectedCategory === category.name ? 'bg-indigo-600 text-white' : 'text-gray-900'"
              >
                <span class="block truncate">
                  {{ category.name }}
                  <span class="text-xs opacity-75">({{ category.task_count }})</span>
                </span>
                <span v-if="selectedCategory === category.name" class="absolute inset-y-0 left-0 flex items-center pl-3 text-current">
                  <CheckIcon class="h-5 w-5" aria-hidden="true" />
                </span>
              </ListboxOption>
            </ListboxOptions>
          </transition>
        </div>
      </Listbox>
    </div>

    <!-- Toggle Controls -->
    <div class="flex items-center gap-6">
      <!-- Show Completed Tasks Toggle -->
      <SwitchGroup>
        <div class="flex items-center">
          <SwitchLabel class="text-sm text-gray-700 mr-3">Show completed tasks</SwitchLabel>
          <Switch
            :model-value="showCompleted"
            @update:model-value="$emit('toggle-completed', $event)"
            data-testid="show-completed-toggle"
            :class="[
              showCompleted ? 'bg-indigo-600' : 'bg-gray-200',
              'relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2'
            ]"
          >
            <span
              :class="[
                showCompleted ? 'translate-x-5' : 'translate-x-0',
                'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out'
              ]"
            />
          </Switch>
        </div>
      </SwitchGroup>

      <!-- Show Deleted Tasks Toggle -->
      <SwitchGroup>
        <div class="flex items-center">
          <SwitchLabel class="text-sm text-gray-700 mr-3">Show deleted tasks</SwitchLabel>
          <Switch
            :model-value="showDeleted"
            @update:model-value="$emit('toggle-deleted', $event)"
            data-testid="show-deleted-toggle"
            :class="[
              showDeleted ? 'bg-indigo-600' : 'bg-gray-200',
              'relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2'
            ]"
          >
            <span
              :class="[
                showDeleted ? 'translate-x-5' : 'translate-x-0',
                'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out'
              ]"
            />
          </Switch>
        </div>
      </SwitchGroup>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  Listbox,
  ListboxButton,
  ListboxOptions,
  ListboxOption,
  Switch,
  SwitchGroup,
  SwitchLabel,
} from '@headlessui/vue'
import { ChevronUpDownIcon, CheckIcon } from '@heroicons/vue/24/outline'
import type { Category } from '~/types'

interface Props {
  categories: Category[]
  selectedCategory: string
  showCompleted: boolean
  showDeleted: boolean
}

defineProps<Props>()

defineEmits<{
  'category-change': [category: string]
  'toggle-completed': [show: boolean]
  'toggle-deleted': [show: boolean]
}>()
</script>