import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import TasksPage from '../../pages/tasks.vue'
import type { Task, Category } from '../../types'

// Mock Pinia store
const mockTasksStore = {
  tasks: [],
  categories: [],
  isLoading: false,
  error: null,
  activeTasks: [],
  completedTasks: [],
  deletedTasks: [],
  fetchTasks: vi.fn(),
  fetchCategories: vi.fn(),
  createTask: vi.fn(),
  updateTask: vi.fn(),
  toggleTaskCompletion: vi.fn(),
  deleteTask: vi.fn(),
  restoreTask: vi.fn(),
  clearError: vi.fn(),
}

const mockAuthStore = {
  user: { id: 'user1', display_name: 'Test User' },
  isAuthenticated: true,
}

vi.mock('../../stores/tasks', () => ({
  useTasksStore: () => mockTasksStore,
}))

vi.mock('../../stores/auth', () => ({
  useAuthStore: () => mockAuthStore,
}))

// Mock components
vi.mock('../../components/TaskFilters.vue', () => ({
  default: {
    name: 'TaskFilters',
    template: '<div data-testid="task-filters" @category-change="$emit(\'category-change\', $event)" @toggle-completed="$emit(\'toggle-completed\', $event)" @toggle-deleted="$emit(\'toggle-deleted\', $event)">Filters</div>',
    props: ['categories', 'selectedCategory', 'showCompleted', 'showDeleted'],
    emits: ['category-change', 'toggle-completed', 'toggle-deleted'],
  },
}))

vi.mock('../../components/TaskList.vue', () => ({
  default: {
    name: 'TaskList',
    template: '<div data-testid="task-list" @toggle-completion="$emit(\'toggle-completion\', $event)" @edit-task="$emit(\'edit-task\', $event)" @delete-task="$emit(\'delete-task\', $event)" @restore-task="$emit(\'restore-task\', $event)">Task List</div>',
    props: ['tasks', 'loading'],
    emits: ['toggle-completion', 'edit-task', 'delete-task', 'restore-task'],
  },
}))

vi.mock('../../components/TaskModal.vue', () => ({
  default: {
    name: 'TaskModal',
    template: '<div data-testid="task-modal" @save="$emit(\'save\', $event)" @close="$emit(\'close\')">Task Modal</div>',
    props: ['isOpen', 'task', 'categories', 'loading'],
    emits: ['save', 'close'],
  },
}))

describe('Tasks Page', () => {
  const mockTasks: Task[] = [
    {
      id: '1',
      user_id: 'user1',
      description: 'Active task',
      category: 'Work',
      completed: false,
      created_at: '2025-08-21T10:00:00Z',
      updated_at: '2025-08-21T10:00:00Z',
    },
    {
      id: '2',
      user_id: 'user1',
      description: 'Completed task',
      category: 'Personal',
      completed: true,
      created_at: '2025-08-21T09:00:00Z',
      updated_at: '2025-08-21T10:30:00Z',
    },
    {
      id: '3',
      user_id: 'user1',
      description: 'Deleted task',
      category: 'Work',
      completed: false,
      created_at: '2025-08-21T08:00:00Z',
      updated_at: '2025-08-21T09:30:00Z',
      deleted_at: '2025-08-21T11:00:00Z',
    },
  ]

  const mockCategories: Category[] = [
    { name: 'Work', task_count: 2 },
    { name: 'Personal', task_count: 1 },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    
    // Reset store state
    mockTasksStore.tasks = [...mockTasks]
    mockTasksStore.categories = [...mockCategories]
    mockTasksStore.isLoading = false
    mockTasksStore.error = null
    mockTasksStore.activeTasks = mockTasks.filter(t => !t.deleted_at && !t.completed)
    mockTasksStore.completedTasks = mockTasks.filter(t => !t.deleted_at && t.completed)
    mockTasksStore.deletedTasks = mockTasks.filter(t => !!t.deleted_at)
  })

  it('renders page with all components', () => {
    const wrapper = mount(TasksPage)

    expect(wrapper.find('h1').text()).toBe('Tasks')
    expect(wrapper.findComponent({ name: 'TaskFilters' }).exists()).toBe(true)
    expect(wrapper.findComponent({ name: 'TaskList' }).exists()).toBe(true)
    expect(wrapper.find('[data-testid="new-task-button"]').exists()).toBe(true)
  })

  it('fetches tasks and categories on mount', () => {
    mount(TasksPage)

    expect(mockTasksStore.fetchTasks).toHaveBeenCalled()
    expect(mockTasksStore.fetchCategories).toHaveBeenCalled()
  })

  it('shows task counts in header', () => {
    const wrapper = mount(TasksPage)

    expect(wrapper.text()).toContain('1 active') // 1 active task
    expect(wrapper.text()).toContain('1 completed') // 1 completed task
    expect(wrapper.text()).toContain('1 deleted') // 1 deleted task
  })

  it('opens modal when New Task button is clicked', async () => {
    const wrapper = mount(TasksPage)

    const newTaskButton = wrapper.find('[data-testid="new-task-button"]')
    await newTaskButton.trigger('click')

    const modal = wrapper.findComponent({ name: 'TaskModal' })
    expect(modal.props('isOpen')).toBe(true)
    expect(modal.props('task')).toBe(null) // Creating new task
  })

  it('opens modal for editing when edit-task event is received', async () => {
    const wrapper = mount(TasksPage)

    const taskList = wrapper.findComponent({ name: 'TaskList' })
    await taskList.vm.$emit('edit-task', '1')

    const modal = wrapper.findComponent({ name: 'TaskModal' })
    expect(modal.props('isOpen')).toBe(true)
    expect(modal.props('task')).toEqual(mockTasks[0])
  })

  it('filters tasks by category', async () => {
    const wrapper = mount(TasksPage)

    const taskFilters = wrapper.findComponent({ name: 'TaskFilters' })
    await taskFilters.vm.$emit('category-change', 'Work')

    // Should filter tasks to only show Work category
    const taskList = wrapper.findComponent({ name: 'TaskList' })
    const filteredTasks = taskList.props('tasks')
    expect(filteredTasks.every(task => task.category === 'Work')).toBe(true)
  })

  it('toggles completed tasks visibility', async () => {
    const wrapper = mount(TasksPage)

    const taskFilters = wrapper.findComponent({ name: 'TaskFilters' })
    await taskFilters.vm.$emit('toggle-completed', false)

    // Should hide completed tasks
    const taskList = wrapper.findComponent({ name: 'TaskList' })
    const visibleTasks = taskList.props('tasks')
    expect(visibleTasks.every(task => !task.completed)).toBe(true)
  })

  it('toggles deleted tasks visibility', async () => {
    const wrapper = mount(TasksPage)

    const taskFilters = wrapper.findComponent({ name: 'TaskFilters' })
    await taskFilters.vm.$emit('toggle-deleted', true)

    // Should show deleted tasks
    const taskList = wrapper.findComponent({ name: 'TaskList' })
    const visibleTasks = taskList.props('tasks')
    expect(visibleTasks.some(task => !!task.deleted_at)).toBe(true)
  })

  it('creates new task when modal save is triggered', async () => {
    const wrapper = mount(TasksPage)

    // Open modal
    const newTaskButton = wrapper.find('[data-testid="new-task-button"]')
    await newTaskButton.trigger('click')

    // Save task
    const modal = wrapper.findComponent({ name: 'TaskModal' })
    const taskData = { description: 'New task', category: 'Work' }
    await modal.vm.$emit('save', taskData)

    expect(mockTasksStore.createTask).toHaveBeenCalledWith(taskData)
  })

  it('updates existing task when modal save is triggered for editing', async () => {
    const wrapper = mount(TasksPage)

    // Open edit modal
    const taskList = wrapper.findComponent({ name: 'TaskList' })
    await taskList.vm.$emit('edit-task', '1')

    // Save changes
    const modal = wrapper.findComponent({ name: 'TaskModal' })
    const updates = { description: 'Updated task', category: 'Personal' }
    await modal.vm.$emit('save', updates)

    expect(mockTasksStore.updateTask).toHaveBeenCalledWith('1', updates)
  })

  it('toggles task completion', async () => {
    const wrapper = mount(TasksPage)

    const taskList = wrapper.findComponent({ name: 'TaskList' })
    await taskList.vm.$emit('toggle-completion', '1')

    expect(mockTasksStore.toggleTaskCompletion).toHaveBeenCalledWith('1')
  })

  it('deletes task', async () => {
    const wrapper = mount(TasksPage)

    const taskList = wrapper.findComponent({ name: 'TaskList' })
    await taskList.vm.$emit('delete-task', '1')

    expect(mockTasksStore.deleteTask).toHaveBeenCalledWith('1')
  })

  it('restores deleted task', async () => {
    const wrapper = mount(TasksPage)

    const taskList = wrapper.findComponent({ name: 'TaskList' })
    await taskList.vm.$emit('restore-task', '3')

    expect(mockTasksStore.restoreTask).toHaveBeenCalledWith('3')
  })

  it('closes modal when close event is received', async () => {
    const wrapper = mount(TasksPage)

    // Open modal
    const newTaskButton = wrapper.find('[data-testid="new-task-button"]')
    await newTaskButton.trigger('click')

    // Close modal
    const modal = wrapper.findComponent({ name: 'TaskModal' })
    await modal.vm.$emit('close')

    expect(modal.props('isOpen')).toBe(false)
  })

  it('passes correct props to TaskFilters component', () => {
    const wrapper = mount(TasksPage)

    const taskFilters = wrapper.findComponent({ name: 'TaskFilters' })
    expect(taskFilters.props('categories')).toEqual(mockCategories)
    expect(taskFilters.props('selectedCategory')).toBe('All')
    expect(taskFilters.props('showCompleted')).toBe(true)
    expect(taskFilters.props('showDeleted')).toBe(false)
  })

  it('passes correct props to TaskList component', () => {
    const wrapper = mount(TasksPage)

    const taskList = wrapper.findComponent({ name: 'TaskList' })
    expect(taskList.props('loading')).toBe(false)
    expect(Array.isArray(taskList.props('tasks'))).toBe(true)
  })

  it('passes correct props to TaskModal component', async () => {
    const wrapper = mount(TasksPage)

    const modal = wrapper.findComponent({ name: 'TaskModal' })
    expect(modal.props('categories')).toEqual(mockCategories)
    expect(modal.props('loading')).toBe(false)
    expect(modal.props('isOpen')).toBe(false)
    expect(modal.props('task')).toBe(null)
  })

  it('shows error message when store has error', async () => {
    mockTasksStore.error = 'Failed to load tasks'
    const wrapper = mount(TasksPage)

    expect(wrapper.find('[data-testid="error-message"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Failed to load tasks')
  })

  it('clears error when user dismisses it', async () => {
    mockTasksStore.error = 'Test error'
    const wrapper = mount(TasksPage)

    const dismissButton = wrapper.find('[data-testid="dismiss-error"]')
    await dismissButton.trigger('click')

    expect(mockTasksStore.clearError).toHaveBeenCalled()
  })

  it('shows loading state in task list', () => {
    mockTasksStore.isLoading = true
    const wrapper = mount(TasksPage)

    const taskList = wrapper.findComponent({ name: 'TaskList' })
    expect(taskList.props('loading')).toBe(true)
  })

  it('applies correct filtering logic for "All" category', async () => {
    const wrapper = mount(TasksPage)

    const taskFilters = wrapper.findComponent({ name: 'TaskFilters' })
    await taskFilters.vm.$emit('category-change', 'All')

    const taskList = wrapper.findComponent({ name: 'TaskList' })
    const visibleTasks = taskList.props('tasks')
    
    // Should show all non-deleted tasks by default
    expect(visibleTasks.length).toBeGreaterThan(0)
    expect(visibleTasks.every(task => !task.deleted_at)).toBe(true)
  })

  it('updates task counts when filters change', async () => {
    const wrapper = mount(TasksPage)

    // Hide completed tasks
    const taskFilters = wrapper.findComponent({ name: 'TaskFilters' })
    await taskFilters.vm.$emit('toggle-completed', false)

    // Task counts should update to reflect filtered view
    expect(wrapper.text()).toContain('1 active')
    expect(wrapper.text()).not.toContain('1 completed')
  })
})