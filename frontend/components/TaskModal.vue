<template>
  <TransitionRoot :show="isOpen" as="template">
    <Dialog @close="$emit('close')" class="relative z-50">
      <!-- Backdrop -->
      <TransitionChild
        enter="ease-out duration-300"
        enter-from="opacity-0"
        enter-to="opacity-100"
        leave="ease-in duration-200"
        leave-from="opacity-100"
        leave-to="opacity-0"
      >
        <div data-testid="modal-backdrop" class="fixed inset-0 bg-gray-900 bg-opacity-50 transition-opacity" />
      </TransitionChild>

      <div class="fixed inset-0 z-10 overflow-y-auto">
        <div class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
          <TransitionChild
            enter="ease-out duration-300"
            enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
            enter-to="opacity-100 translate-y-0 sm:scale-100"
            leave="ease-in duration-200"
            leave-from="opacity-100 translate-y-0 sm:scale-100"
            leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
          >
            <DialogPanel data-testid="modal-panel" class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6">
              <!-- Modal Header -->
              <div class="mb-4">
                <DialogTitle class="text-lg font-semibold leading-6 text-gray-900">
                  {{ task ? 'Edit Task' : 'Create Task' }}
                </DialogTitle>
              </div>

              <!-- Form -->
              <form @submit.prevent="handleSubmit" @keydown.escape="$emit('close')">
                <!-- Description Field -->
                <div class="mb-4">
                  <label for="description" class="block text-sm font-medium text-gray-700 mb-2">
                    Description
                  </label>
                  <textarea
                    id="description"
                    ref="descriptionRef"
                    v-model="form.description"
                    data-testid="task-description"
                    rows="3"
                    placeholder="Enter task description..."
                    @keydown.enter.prevent="handleEnterKey"
                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                    :class="errors.description ? 'ring-red-500 focus:ring-red-500' : ''"
                  />
                  <p
                    v-if="errors.description"
                    data-testid="description-error"
                    class="mt-1 text-sm text-red-600"
                  >
                    {{ errors.description }}
                  </p>
                </div>

                <!-- Category Field -->
                <div class="mb-6">
                  <label class="block text-sm font-medium text-gray-700 mb-2">
                    Category
                  </label>
                  <Listbox v-model="form.category" data-testid="task-category">
                    <div class="relative">
                      <ListboxButton
                        class="relative cursor-pointer rounded-lg bg-white py-2 pl-3 pr-10 text-left shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm w-full"
                        :class="errors.category ? 'ring-red-500 focus:ring-red-500' : ''"
                      >
                        <span class="block truncate">
                          {{ form.category || 'Select a category' }}
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
                          <!-- Existing Categories -->
                          <ListboxOption
                            v-for="category in categories"
                            :key="category.name"
                            :value="category.name"
                            class="cursor-pointer select-none relative py-2 pl-10 pr-4 hover:bg-indigo-600 hover:text-white"
                            :class="form.category === category.name ? 'bg-indigo-600 text-white' : 'text-gray-900'"
                          >
                            <span class="block truncate">
                              {{ category.name }}
                              <span class="text-xs opacity-75">({{ category.task_count }})</span>
                            </span>
                            <span v-if="form.category === category.name" class="absolute inset-y-0 left-0 flex items-center pl-3 text-current">
                              <CheckIcon class="h-5 w-5" aria-hidden="true" />
                            </span>
                          </ListboxOption>

                          <!-- Create New Category Option -->
                          <ListboxOption
                            value=""
                            class="cursor-pointer select-none relative py-2 pl-10 pr-4 hover:bg-indigo-600 hover:text-white text-gray-900 border-t border-gray-200"
                          >
                            <span class="block truncate font-medium">
                              + Create new category
                            </span>
                            <PlusIcon class="absolute inset-y-0 left-0 flex items-center pl-3 h-5 w-5" />
                          </ListboxOption>
                        </ListboxOptions>
                      </transition>
                    </div>
                  </Listbox>

                  <!-- New Category Input (shown when creating new category) -->
                  <div v-if="form.category === ''" class="mt-2">
                    <input
                      v-model="newCategoryName"
                      data-testid="new-category-input"
                      type="text"
                      placeholder="Enter new category name..."
                      class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                    />
                  </div>
                  
                  <p
                    v-if="errors.category"
                    data-testid="category-error"
                    class="mt-1 text-sm text-red-600"
                  >
                    {{ errors.category }}
                  </p>
                </div>

                <!-- Action Buttons -->
                <div class="flex gap-3 justify-end">
                  <button
                    type="button"
                    data-testid="cancel-button"
                    @click="$emit('close')"
                    class="inline-flex justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus-visible:outline-offset-0"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    data-testid="save-button"
                    :disabled="loading"
                    class="inline-flex justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {{ loading ? 'Saving...' : 'Save' }}
                  </button>
                </div>
              </form>
            </DialogPanel>
          </TransitionChild>
        </div>
      </div>
    </Dialog>
  </TransitionRoot>
</template>

<script setup lang="ts">
import { ref, reactive, watch, nextTick } from 'vue'
import {
  Dialog,
  DialogPanel,
  DialogTitle,
  TransitionChild,
  TransitionRoot,
  Listbox,
  ListboxButton,
  ListboxOptions,
  ListboxOption,
} from '@headlessui/vue'
import { ChevronUpDownIcon, CheckIcon, PlusIcon } from '@heroicons/vue/24/outline'
import type { Task, Category } from '~/types'

interface Props {
  isOpen: boolean
  task?: Task | null
  categories: Category[]
  loading: boolean
}

const props = defineProps<Props>()

const emit = defineEmits<{
  save: [data: { description: string; category: string }]
  close: []
}>()

// Form state
const form = reactive({
  description: '',
  category: '',
})

const newCategoryName = ref('')
const errors = reactive({
  description: '',
  category: '',
})

const descriptionRef = ref<HTMLTextAreaElement>()

const resetForm = () => {
  form.description = ''
  form.category = props.categories.length > 0 ? props.categories[0].name : ''
  newCategoryName.value = ''
}

const clearErrors = () => {
  errors.description = ''
  errors.category = ''
}

// Watch for task changes to populate form
watch(
  () => props.task,
  (newTask) => {
    if (newTask) {
      form.description = newTask.description
      form.category = newTask.category
    } else {
      resetForm()
    }
    clearErrors()
  },
  { immediate: true }
)

// Watch for modal open to focus description field
watch(
  () => props.isOpen,
  (isOpen) => {
    if (isOpen) {
      nextTick(() => {
        if (descriptionRef.value) {
          descriptionRef.value.focus()
        }
      })
      // Set default category if no categories exist or if creating new task
      if (!props.task && props.categories.length > 0) {
        form.category = props.categories[0].name
      }
    }
  }
)

const validateForm = () => {
  let isValid = true
  clearErrors()

  // Validate description
  if (!form.description.trim()) {
    errors.description = 'Description is required'
    isValid = false
  } else if (form.description.length > 500) {
    errors.description = 'Description is too long (maximum 500 characters)'
    isValid = false
  }

  // Validate category  
  const finalCategory = form.category === '' ? newCategoryName.value.trim() : form.category
  if (!finalCategory) {
    // If no category is provided and no categories exist, use 'General' as default
    if (props.categories.length === 0) {
      // This will be handled in handleSubmit by setting 'General' as default
    } else {
      errors.category = 'Category is required'
      isValid = false
    }
  }

  return isValid
}

const handleSubmit = () => {
  console.log('🔧 Modal handleSubmit called')
  console.log('📋 Form description:', form.description)
  console.log('📋 Form category:', form.category) 
  console.log('📂 New category name:', newCategoryName.value)
  console.log('📦 Categories count:', props.categories.length)
  
  if (!validateForm()) {
    console.log('❌ Form validation failed')
    console.log('🚫 Errors:', errors)
    return
  }
  
  console.log('✅ Form validation passed')
  let finalCategory = form.category === '' ? newCategoryName.value.trim() : form.category
  
  // If still no category and no categories exist, use 'General' as default
  if (!finalCategory && props.categories.length === 0) {
    finalCategory = 'General'
  }
  
  console.log('📁 Final category:', finalCategory)
  
  const saveData = {
    description: form.description.trim(),
    category: finalCategory,
  }
  console.log('💾 Emitting save event with:', saveData)

  emit('save', saveData)

  // If creating new task (not editing), clear form and keep modal open
  if (!props.task) {
    resetForm()
    nextTick(() => {
      if (descriptionRef.value) {
        descriptionRef.value.focus()
      }
    })
  } else {
    // If editing, close modal
    emit('close')
  }
}

const handleEnterKey = (event: KeyboardEvent) => {
  if (!event.shiftKey) {
    event.preventDefault()
    handleSubmit()
  }
}
</script>