import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import TaskItem from '../../components/TaskItem.vue'
import type { Task } from '../../types'

describe('TaskItem', () => {
  const mockTask: Task = {
    id: '1',
    user_id: 'user1',
    description: 'Test task description',
    category: 'Work',
    completed: false,
    created_at: '2025-08-21T10:00:00Z',
    updated_at: '2025-08-21T10:00:00Z',
  }

  const completedTask: Task = {
    ...mockTask,
    id: '2',
    completed: true,
    description: 'Completed task',
  }

  const deletedTask: Task = {
    ...mockTask,
    id: '3',
    deleted_at: '2025-08-21T11:00:00Z',
    description: 'Deleted task',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders task information correctly', () => {
    const wrapper = mount(TaskItem, {
      props: { task: mockTask },
    })

    expect(wrapper.text()).toContain('Test task description')
    expect(wrapper.text()).toContain('Work')
    expect(wrapper.find('[data-testid="task-description"]').text()).toBe('Test task description')
    expect(wrapper.find('[data-testid="task-category"]').text()).toBe('Work')
  })

  it('displays completion checkbox', () => {
    const wrapper = mount(TaskItem, {
      props: { task: mockTask },
    })

    const checkbox = wrapper.find('[data-testid="task-checkbox"]')
    expect(checkbox.exists()).toBe(true)
    expect(checkbox.element.checked).toBe(false)
  })

  it('shows completed state with strikethrough', () => {
    const wrapper = mount(TaskItem, {
      props: { task: completedTask },
    })

    const checkbox = wrapper.find('[data-testid="task-checkbox"]')
    const description = wrapper.find('[data-testid="task-description"]')
    
    expect(checkbox.element.checked).toBe(true)
    expect(description.classes()).toContain('line-through')
  })

  it('emits toggle-completion event when checkbox is clicked', async () => {
    const wrapper = mount(TaskItem, {
      props: { task: mockTask },
    })

    const checkbox = wrapper.find('[data-testid="task-checkbox"]')
    await checkbox.trigger('change')

    expect(wrapper.emitted('toggle-completion')).toBeTruthy()
    expect(wrapper.emitted('toggle-completion')![0]).toEqual(['1'])
  })

  it('emits edit-task event when task is clicked', async () => {
    const wrapper = mount(TaskItem, {
      props: { task: mockTask },
    })

    const taskContent = wrapper.find('[data-testid="task-content"]')
    await taskContent.trigger('click')

    expect(wrapper.emitted('edit-task')).toBeTruthy()
    expect(wrapper.emitted('edit-task')![0]).toEqual(['1'])
  })

  it('shows delete button and emits delete-task event', async () => {
    const wrapper = mount(TaskItem, {
      props: { task: mockTask },
    })

    const deleteButton = wrapper.find('[data-testid="delete-button"]')
    expect(deleteButton.exists()).toBe(true)
    
    await deleteButton.trigger('click')

    expect(wrapper.emitted('delete-task')).toBeTruthy()
    expect(wrapper.emitted('delete-task')![0]).toEqual(['1'])
  })

  it('shows restore button for deleted tasks', () => {
    const wrapper = mount(TaskItem, {
      props: { task: deletedTask },
    })

    const restoreButton = wrapper.find('[data-testid="restore-button"]')
    expect(restoreButton.exists()).toBe(true)
    
    const deleteButton = wrapper.find('[data-testid="delete-button"]')
    expect(deleteButton.exists()).toBe(false)
  })

  it('emits restore-task event when restore button is clicked', async () => {
    const wrapper = mount(TaskItem, {
      props: { task: deletedTask },
    })

    const restoreButton = wrapper.find('[data-testid="restore-button"]')
    await restoreButton.trigger('click')

    expect(wrapper.emitted('restore-task')).toBeTruthy()
    expect(wrapper.emitted('restore-task')![0]).toEqual(['3'])
  })

  it('applies deleted styling to deleted tasks', () => {
    const wrapper = mount(TaskItem, {
      props: { task: deletedTask },
    })

    const container = wrapper.find('[data-testid="task-container"]')
    expect(container.classes()).toContain('border-red-200')
    expect(container.classes()).toContain('bg-red-50')
  })

  it('has beautiful card design with hover effects', () => {
    const wrapper = mount(TaskItem, {
      props: { task: mockTask },
    })

    const container = wrapper.find('[data-testid="task-container"]')
    expect(container.classes()).toContain('bg-white')
    expect(container.classes()).toContain('rounded-lg')
    expect(container.classes()).toContain('shadow-sm')
    expect(container.classes()).toContain('hover:shadow-md')
  })

  it('shows category badge with correct styling', () => {
    const wrapper = mount(TaskItem, {
      props: { task: mockTask },
    })

    const categoryBadge = wrapper.find('[data-testid="task-category"]')
    expect(categoryBadge.classes()).toContain('bg-indigo-100')
    expect(categoryBadge.classes()).toContain('text-indigo-800')
    expect(categoryBadge.classes()).toContain('rounded-full')
  })

  it('prevents checkbox click from triggering task edit', async () => {
    const wrapper = mount(TaskItem, {
      props: { task: mockTask },
    })

    const checkbox = wrapper.find('[data-testid="task-checkbox"]')
    await checkbox.trigger('click')

    // Should only emit toggle-completion, not edit-task
    expect(wrapper.emitted('toggle-completion')).toBeTruthy()
    expect(wrapper.emitted('edit-task')).toBeFalsy()
  })

  it('prevents delete button click from triggering task edit', async () => {
    const wrapper = mount(TaskItem, {
      props: { task: mockTask },
    })

    const deleteButton = wrapper.find('[data-testid="delete-button"]')
    await deleteButton.trigger('click')

    // Should only emit delete-task, not edit-task
    expect(wrapper.emitted('delete-task')).toBeTruthy()
    expect(wrapper.emitted('edit-task')).toBeFalsy()
  })

  it('handles long task descriptions gracefully', () => {
    const longTask: Task = {
      ...mockTask,
      description: 'This is a very long task description that should be handled gracefully by the component and not break the layout or cause overflow issues',
    }

    const wrapper = mount(TaskItem, {
      props: { task: longTask },
    })

    const description = wrapper.find('[data-testid="task-description"]')
    expect(description.exists()).toBe(true)
    expect(description.text()).toContain('This is a very long task description')
  })

  it('shows correct timestamps', () => {
    const wrapper = mount(TaskItem, {
      props: { task: mockTask },
    })

    // Should show created date - the formatDate function formats recent dates as time
    expect(wrapper.html()).toContain('Created')
  })

  it('applies correct cursor styles', () => {
    const wrapper = mount(TaskItem, {
      props: { task: mockTask },
    })

    const taskContent = wrapper.find('[data-testid="task-content"]')
    expect(taskContent.classes()).toContain('cursor-pointer')
  })

  it('has accessible button labels', () => {
    const wrapper = mount(TaskItem, {
      props: { task: mockTask },
    })

    const deleteButton = wrapper.find('[data-testid="delete-button"]')
    expect(deleteButton.attributes('aria-label')).toContain('Delete task')
  })

  it('shows restore button with correct label for deleted tasks', () => {
    const wrapper = mount(TaskItem, {
      props: { task: deletedTask },
    })

    const restoreButton = wrapper.find('[data-testid="restore-button"]')
    expect(restoreButton.attributes('aria-label')).toContain('Restore task')
  })
})