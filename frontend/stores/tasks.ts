import { defineStore } from 'pinia'
import { ref, computed, readonly } from 'vue'
import { useApi } from '../composables/useApi'
import type { Task, Category, CreateTaskRequest, UpdateTaskRequest } from '../types'

export const useTasksStore = defineStore('tasks', () => {
  // State
  const tasks = ref<Task[]>([])
  const categories = ref<Category[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const activeTasks = computed(() => 
    tasks.value.filter(task => !task.deleted_at && !task.completed)
  )
  
  const completedTasks = computed(() => 
    tasks.value.filter(task => !task.deleted_at && task.completed)
  )
  
  const deletedTasks = computed(() => 
    tasks.value.filter(task => !!task.deleted_at)
  )

  const tasksByCategory = computed(() => {
    const grouped: Record<string, Task[]> = {}
    activeTasks.value.forEach(task => {
      if (!grouped[task.category]) {
        grouped[task.category] = []
      }
      grouped[task.category].push(task)
    })
    return grouped
  })

  // Actions
  const { apiCall } = useApi()

  const fetchTasks = async () => {
    try {
      isLoading.value = true
      error.value = null
      
      const tasksData = await apiCall<Task[]>('/tasks')
      tasks.value = tasksData
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch tasks'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const fetchCategories = async () => {
    try {
      error.value = null
      
      const categoriesData = await apiCall<Category[]>('/categories')
      categories.value = categoriesData
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch categories'
      throw err
    }
  }

  const createTask = async (taskData: CreateTaskRequest) => {
    try {
      isLoading.value = true
      error.value = null
      
      console.log('📝 Creating task:', taskData)
      const newTask = await apiCall<Task>('/tasks', {
        method: 'POST',
        body: taskData,
      })
      
      console.log('✅ Task created:', newTask)
      // Add to local tasks array
      const currentTasks = [...tasks.value, newTask]
      tasks.value = currentTasks
      await fetchCategories() // Refresh categories
      console.log('🔄 Task added to list, total tasks:', tasks.value.length)
    } catch (err: any) {
      console.error('❌ Task creation failed:', err)
      error.value = err.message || 'Failed to create task'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const updateTask = async (taskId: string, updates: UpdateTaskRequest) => {
    try {
      isLoading.value = true
      error.value = null
      
      const updatedTask = await apiCall<Task>(`/tasks/${taskId}`, {
        method: 'PUT',
        body: updates,
      })
      
      const index = tasks.value.findIndex(task => task.id === taskId)
      if (index !== -1) {
        tasks.value[index] = updatedTask
      }
      
      await fetchCategories() // Refresh categories if category changed
    } catch (err: any) {
      error.value = err.message || 'Failed to update task'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const toggleTaskCompletion = async (taskId: string) => {
    const task = tasks.value.find(t => t.id === taskId)
    if (!task) return
    
    await updateTask(taskId, { completed: !task.completed })
  }

  const deleteTask = async (taskId: string) => {
    try {
      isLoading.value = true
      error.value = null
      
      await apiCall(`/tasks/${taskId}`, {
        method: 'DELETE',
      })
      
      // Mark task as deleted in local state
      const task = tasks.value.find(t => t.id === taskId)
      if (task) {
        task.deleted_at = new Date().toISOString()
      }
      
      await fetchCategories() // Refresh categories
    } catch (err: any) {
      error.value = err.message || 'Failed to delete task'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const restoreTask = async (taskId: string) => {
    try {
      isLoading.value = true
      error.value = null
      
      const restoredTask = await apiCall<Task>(`/tasks/${taskId}/restore`, {
        method: 'POST',
      })
      
      const index = tasks.value.findIndex(task => task.id === taskId)
      if (index !== -1) {
        tasks.value[index] = restoredTask
      }
      
      await fetchCategories() // Refresh categories
    } catch (err: any) {
      error.value = err.message || 'Failed to restore task'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const renameCategory = async (oldName: string, newName: string) => {
    try {
      isLoading.value = true
      error.value = null
      
      await apiCall(`/categories/${encodeURIComponent(oldName)}`, {
        method: 'PUT',
        body: { name: newName },
      })
      
      // Update local state
      tasks.value.forEach(task => {
        if (task.category === oldName) {
          task.category = newName
        }
      })
      
      await fetchCategories() // Refresh categories
    } catch (err: any) {
      error.value = err.message || 'Failed to rename category'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const deleteCategory = async (categoryName: string) => {
    try {
      isLoading.value = true
      error.value = null
      
      await apiCall(`/categories/${encodeURIComponent(categoryName)}`, {
        method: 'DELETE',
      })
      
      // Remove tasks from this category locally
      tasks.value = tasks.value.filter(task => task.category !== categoryName)
      
      await fetchCategories() // Refresh categories
    } catch (err: any) {
      error.value = err.message || 'Failed to delete category'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const clearError = () => {
    error.value = null
  }

  // Helper methods for filtering
  const getTasksByCategory = (categoryName: string) => {
    if (categoryName === 'All') {
      return tasks.value
    }
    return tasks.value.filter(task => task.category === categoryName)
  }

  const getTasksWithFilters = (options: {
    category?: string
    showCompleted?: boolean
    showDeleted?: boolean
  } = {}) => {
    let filteredTasks = [...tasks.value]

    // Filter by category
    if (options.category && options.category !== 'All') {
      filteredTasks = filteredTasks.filter(task => task.category === options.category)
    }

    // Filter by completion status
    if (options.showCompleted === false) {
      filteredTasks = filteredTasks.filter(task => !task.completed)
    }

    // Filter by deleted status
    if (options.showDeleted === false) {
      filteredTasks = filteredTasks.filter(task => !task.deleted_at)
    }

    return filteredTasks
  }

  const getTaskCounts = () => {
    const active = tasks.value.filter(t => !t.deleted_at && !t.completed).length
    const completed = tasks.value.filter(t => !t.deleted_at && t.completed).length
    const deleted = tasks.value.filter(t => !!t.deleted_at).length
    
    return { active, completed, deleted }
  }

  // Additional computed properties for dashboard compatibility
  const totalTasks = computed(() => tasks.value.filter(t => !t.deleted_at).length)

  // Alias methods for backward compatibility
  const loadTasks = fetchTasks
  const loadCategories = fetchCategories

  return {
    // State
    tasks: readonly(tasks),
    categories: readonly(categories),
    isLoading: readonly(isLoading),
    error: readonly(error),
    
    // Getters
    activeTasks,
    completedTasks,
    deletedTasks,
    tasksByCategory,
    totalTasks,
    
    // Actions
    fetchTasks,
    fetchCategories,
    loadTasks, // Alias for fetchTasks
    loadCategories, // Alias for fetchCategories
    createTask,
    updateTask,
    toggleTaskCompletion,
    deleteTask,
    restoreTask,
    renameCategory,
    deleteCategory,
    clearError,
    
    // Helper methods
    getTasksByCategory,
    getTasksWithFilters,
    getTaskCounts,
  }
})