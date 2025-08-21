import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import TaskList from '../../components/TaskList.vue'
import TaskItem from '../../components/TaskItem.vue'
import type { Task } from '../../types'

// Mock TaskItem component
vi.mock('../../components/TaskItem.vue', () => ({
  default: {
    name: 'TaskItem',
    template: '<div data-testid="task-item" @toggle-completion="$emit(\'toggle-completion\', $event)" @edit-task="$emit(\'edit-task\', $event)" @delete-task="$emit(\'delete-task\', $event)" @restore-task="$emit(\'restore-task\', $event)">{{ task.description }}</div>',
    props: ['task'],
    emits: ['toggle-completion', 'edit-task', 'delete-task', 'restore-task'],
  },
}))

describe('TaskList', () => {
  const mockTasks: Task[] = [
    {
      id: '1',
      user_id: 'user1',
      description: 'First task',
      category: 'Work',
      completed: false,
      created_at: '2025-08-21T10:00:00Z',
      updated_at: '2025-08-21T10:00:00Z',
    },
    {
      id: '2',
      user_id: 'user1',
      description: 'Second task',
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

  const defaultProps = {
    tasks: mockTasks,
    loading: false,
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders task items using TaskItem components', () => {
    const wrapper = mount(TaskList, {
      props: defaultProps,
    })

    const taskItems = wrapper.findAllComponents(TaskItem)
    expect(taskItems).toHaveLength(3)
    
    // Check that each task is passed to TaskItem
    expect(taskItems[0].props('task')).toEqual(mockTasks[0])
    expect(taskItems[1].props('task')).toEqual(mockTasks[1])
    expect(taskItems[2].props('task')).toEqual(mockTasks[2])
  })

  it('displays empty state when no tasks', () => {
    const wrapper = mount(TaskList, {
      props: {
        tasks: [],
        loading: false,
      },
    })

    const emptyState = wrapper.find('[data-testid="empty-state"]')
    expect(emptyState.exists()).toBe(true)
    expect(emptyState.text()).toContain('No tasks found')
    
    const taskItems = wrapper.findAllComponents(TaskItem)
    expect(taskItems).toHaveLength(0)
  })

  it('displays loading state while fetching', () => {
    const wrapper = mount(TaskList, {
      props: {
        tasks: [],
        loading: true,
      },
    })

    const loadingState = wrapper.find('[data-testid="loading-state"]')
    expect(loadingState.exists()).toBe(true)
    expect(loadingState.text()).toContain('Loading tasks...')
    
    const emptyState = wrapper.find('[data-testid="empty-state"]')
    expect(emptyState.exists()).toBe(false)
  })

  it('uses responsive grid layout', () => {
    const wrapper = mount(TaskList, {
      props: defaultProps,
    })

    const container = wrapper.find('[data-testid="tasks-container"]')
    expect(container.exists()).toBe(true)
    expect(container.classes()).toContain('grid')
    expect(container.classes()).toContain('gap-4')
    expect(container.classes()).toContain('sm:grid-cols-2')
    expect(container.classes()).toContain('lg:grid-cols-3')
  })

  it('forwards toggle-completion event from TaskItem', async () => {
    const wrapper = mount(TaskList, {
      props: defaultProps,
    })

    const firstTaskItem = wrapper.findAllComponents(TaskItem)[0]
    await firstTaskItem.vm.$emit('toggle-completion', '1')

    expect(wrapper.emitted('toggle-completion')).toBeTruthy()
    expect(wrapper.emitted('toggle-completion')![0]).toEqual(['1'])
  })

  it('forwards edit-task event from TaskItem', async () => {
    const wrapper = mount(TaskList, {
      props: defaultProps,
    })

    const firstTaskItem = wrapper.findAllComponents(TaskItem)[0]
    await firstTaskItem.vm.$emit('edit-task', '1')

    expect(wrapper.emitted('edit-task')).toBeTruthy()
    expect(wrapper.emitted('edit-task')![0]).toEqual(['1'])
  })

  it('forwards delete-task event from TaskItem', async () => {
    const wrapper = mount(TaskList, {
      props: defaultProps,
    })

    const firstTaskItem = wrapper.findAllComponents(TaskItem)[0]
    await firstTaskItem.vm.$emit('delete-task', '1')

    expect(wrapper.emitted('delete-task')).toBeTruthy()
    expect(wrapper.emitted('delete-task')![0]).toEqual(['1'])
  })

  it('forwards restore-task event from TaskItem', async () => {
    const wrapper = mount(TaskList, {
      props: defaultProps,
    })

    const thirdTaskItem = wrapper.findAllComponents(TaskItem)[2]
    await thirdTaskItem.vm.$emit('restore-task', '3')

    expect(wrapper.emitted('restore-task')).toBeTruthy()
    expect(wrapper.emitted('restore-task')![0]).toEqual(['3'])
  })

  it('handles single task display', () => {
    const wrapper = mount(TaskList, {
      props: {
        tasks: [mockTasks[0]],
        loading: false,
      },
    })

    const taskItems = wrapper.findAllComponents(TaskItem)
    expect(taskItems).toHaveLength(1)
    expect(taskItems[0].props('task')).toEqual(mockTasks[0])
  })

  it('displays tasks in correct order (as provided)', () => {
    const wrapper = mount(TaskList, {
      props: defaultProps,
    })

    const taskItems = wrapper.findAllComponents(TaskItem)
    expect(taskItems[0].props('task').id).toBe('1')
    expect(taskItems[1].props('task').id).toBe('2')
    expect(taskItems[2].props('task').id).toBe('3')
  })

  it('shows loading state instead of tasks when loading', () => {
    const wrapper = mount(TaskList, {
      props: {
        tasks: mockTasks,
        loading: true,
      },
    })

    const loadingState = wrapper.find('[data-testid="loading-state"]')
    expect(loadingState.exists()).toBe(true)
    
    const tasksContainer = wrapper.find('[data-testid="tasks-container"]')
    expect(tasksContainer.exists()).toBe(false)
  })

  it('has proper container styling', () => {
    const wrapper = mount(TaskList, {
      props: defaultProps,
    })

    const container = wrapper.find('[data-testid="tasks-container"]')
    expect(container.classes()).toContain('grid')
    expect(container.classes()).toContain('gap-4')
  })

  it('shows appropriate empty state message', () => {
    const wrapper = mount(TaskList, {
      props: {
        tasks: [],
        loading: false,
      },
    })

    const emptyState = wrapper.find('[data-testid="empty-state"]')
    expect(emptyState.text()).toContain('No tasks found')
    expect(emptyState.text()).toContain('Create your first task to get started')
  })

  it('shows loading spinner in loading state', () => {
    const wrapper = mount(TaskList, {
      props: {
        tasks: [],
        loading: true,
      },
    })

    const loadingSpinner = wrapper.find('[data-testid="loading-spinner"]')
    expect(loadingSpinner.exists()).toBe(true)
    expect(loadingSpinner.classes()).toContain('animate-spin')
  })

  it('applies correct spacing and layout classes', () => {
    const wrapper = mount(TaskList, {
      props: defaultProps,
    })

    const container = wrapper.find('[data-testid="tasks-container"]')
    // Should have responsive grid with proper breakpoints
    expect(container.classes()).toEqual(expect.arrayContaining([
      'grid',
      'gap-4',
      'sm:grid-cols-2',
      'lg:grid-cols-3',
    ]))
  })

  it('handles edge case with undefined tasks prop gracefully', () => {
    const wrapper = mount(TaskList, {
      props: {
        tasks: undefined as any,
        loading: false,
      },
    })

    // Should show empty state instead of crashing
    const emptyState = wrapper.find('[data-testid="empty-state"]')
    expect(emptyState.exists()).toBe(true)
  })
})