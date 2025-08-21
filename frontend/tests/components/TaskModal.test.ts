import { describe, it, expect, vi, beforeEach, nextTick } from 'vitest'
import { mount } from '@vue/test-utils'
import TaskModal from '../../components/TaskModal.vue'
import type { Task, Category } from '../../types'

// Mock Headless UI components
vi.mock('@headlessui/vue', () => ({
  Dialog: {
    template: '<div v-if="open" data-testid="modal-backdrop"><slot /></div>',
    props: ['open'],
    emits: ['close'],
  },
  DialogPanel: {
    template: '<div data-testid="modal-panel"><slot /></div>',
  },
  DialogTitle: {
    template: '<h3><slot /></h3>',
  },
  TransitionChild: {
    template: '<div><slot /></div>',
  },
  TransitionRoot: {
    template: '<div><slot /></div>',
    props: ['show'],
  },
  Listbox: {
    template: '<div><slot /></div>',
    props: ['modelValue'],
    emits: ['update:modelValue'],
  },
  ListboxButton: {
    template: '<button type="button"><slot /></button>',
  },
  ListboxOption: {
    template: '<div><slot /></div>',
    props: ['value'],
  },
  ListboxOptions: {
    template: '<div><slot /></div>',
  },
}))

describe('TaskModal', () => {
  const mockCategories: Category[] = [
    { name: 'Work', task_count: 5 },
    { name: 'Personal', task_count: 3 },
    { name: 'Shopping', task_count: 2 },
  ]

  const mockTask: Task = {
    id: '1',
    user_id: 'user1',
    description: 'Test task description',
    category: 'Work',
    completed: false,
    created_at: '2025-08-21T10:00:00Z',
    updated_at: '2025-08-21T10:00:00Z',
  }

  const defaultProps = {
    isOpen: true,
    categories: mockCategories,
    task: null,
    loading: false,
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders modal when isOpen is true', () => {
    const wrapper = mount(TaskModal, {
      props: defaultProps,
    })

    expect(wrapper.find('[data-testid="modal-backdrop"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="modal-panel"]').exists()).toBe(true)
  })

  it('does not render modal when isOpen is false', () => {
    const wrapper = mount(TaskModal, {
      props: {
        ...defaultProps,
        isOpen: false,
      },
    })

    expect(wrapper.find('[data-testid="modal-backdrop"]').exists()).toBe(false)
  })

  it('shows "Create Task" title when no task is provided', () => {
    const wrapper = mount(TaskModal, {
      props: defaultProps,
    })

    expect(wrapper.text()).toContain('Create Task')
  })

  it('shows "Edit Task" title when task is provided', () => {
    const wrapper = mount(TaskModal, {
      props: {
        ...defaultProps,
        task: mockTask,
      },
    })

    expect(wrapper.text()).toContain('Edit Task')
  })

  it('displays form with description textarea and category dropdown', () => {
    const wrapper = mount(TaskModal, {
      props: defaultProps,
    })

    expect(wrapper.find('[data-testid="task-description"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="task-category"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="task-description"]').element.tagName).toBe('TEXTAREA')
  })

  it('populates form with task data when editing', async () => {
    const wrapper = mount(TaskModal, {
      props: {
        ...defaultProps,
        task: mockTask,
      },
    })

    await nextTick()

    const descriptionField = wrapper.find('[data-testid="task-description"]')
    expect(descriptionField.element.value).toBe('Test task description')
  })

  it('shows Save and Cancel buttons', () => {
    const wrapper = mount(TaskModal, {
      props: defaultProps,
    })

    expect(wrapper.find('[data-testid="save-button"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="cancel-button"]').exists()).toBe(true)
  })

  it('emits close event when cancel button is clicked', async () => {
    const wrapper = mount(TaskModal, {
      props: defaultProps,
    })

    const cancelButton = wrapper.find('[data-testid="cancel-button"]')
    await cancelButton.trigger('click')

    expect(wrapper.emitted('close')).toBeTruthy()
  })

  it('emits save event with form data when save button is clicked', async () => {
    const wrapper = mount(TaskModal, {
      props: defaultProps,
    })

    const descriptionField = wrapper.find('[data-testid="task-description"]')
    await descriptionField.setValue('New task description')

    const saveButton = wrapper.find('[data-testid="save-button"]')
    await saveButton.trigger('click')

    expect(wrapper.emitted('save')).toBeTruthy()
    expect(wrapper.emitted('save')![0][0]).toEqual({
      description: 'New task description',
      category: 'Work', // Default category
    })
  })

  it('submits form when Enter key is pressed in textarea', async () => {
    const wrapper = mount(TaskModal, {
      props: defaultProps,
    })

    const descriptionField = wrapper.find('[data-testid="task-description"]')
    await descriptionField.setValue('Task via Enter key')
    await descriptionField.trigger('keydown.enter', { preventDefault: vi.fn() })

    expect(wrapper.emitted('save')).toBeTruthy()
    expect(wrapper.emitted('save')![0][0]).toEqual({
      description: 'Task via Enter key',
      category: 'Work',
    })
  })

  it('auto-focuses description field when modal opens', async () => {
    const wrapper = mount(TaskModal, {
      props: {
        ...defaultProps,
        isOpen: false,
      },
    })

    await wrapper.setProps({ isOpen: true })
    await nextTick()

    const descriptionField = wrapper.find('[data-testid="task-description"]')
    expect(document.activeElement).toBe(descriptionField.element)
  })

  it('shows categories in dropdown including "Create new category" option', () => {
    const wrapper = mount(TaskModal, {
      props: defaultProps,
    })

    expect(wrapper.text()).toContain('Work')
    expect(wrapper.text()).toContain('Personal')
    expect(wrapper.text()).toContain('Shopping')
    expect(wrapper.text()).toContain('Create new category')
  })

  it('allows creating new category via text input', async () => {
    const wrapper = mount(TaskModal, {
      props: defaultProps,
    })

    // Simulate selecting "Create new category" option
    const categoryListbox = wrapper.findComponent({ name: 'Listbox' })
    await categoryListbox.vm.$emit('update:modelValue', '')

    // Should show text input for new category
    const newCategoryInput = wrapper.find('[data-testid="new-category-input"]')
    expect(newCategoryInput.exists()).toBe(true)
  })

  it('validates required description field', async () => {
    const wrapper = mount(TaskModal, {
      props: defaultProps,
    })

    const saveButton = wrapper.find('[data-testid="save-button"]')
    await saveButton.trigger('click')

    const errorMessage = wrapper.find('[data-testid="description-error"]')
    expect(errorMessage.exists()).toBe(true)
    expect(errorMessage.text()).toContain('Description is required')
  })

  it('shows loading state on save button when loading', () => {
    const wrapper = mount(TaskModal, {
      props: {
        ...defaultProps,
        loading: true,
      },
    })

    const saveButton = wrapper.find('[data-testid="save-button"]')
    expect(saveButton.text()).toContain('Saving...')
    expect(saveButton.attributes('disabled')).toBeDefined()
  })

  it('clears form after successful save when creating new task', async () => {
    const wrapper = mount(TaskModal, {
      props: defaultProps,
    })

    const descriptionField = wrapper.find('[data-testid="task-description"]')
    await descriptionField.setValue('New task')

    const saveButton = wrapper.find('[data-testid="save-button"]')
    await saveButton.trigger('click')

    expect(wrapper.emitted('save')).toBeTruthy()

    // Simulate successful save by emitting save-success
    await wrapper.vm.$emit('save-success')

    expect(descriptionField.element.value).toBe('')
  })

  it('keeps modal open after creating task for continuous creation', async () => {
    const wrapper = mount(TaskModal, {
      props: defaultProps,
    })

    const descriptionField = wrapper.find('[data-testid="task-description"]')
    await descriptionField.setValue('New task')

    const saveButton = wrapper.find('[data-testid="save-button"]')
    await saveButton.trigger('click')

    // Modal should remain open and form should be cleared
    expect(wrapper.find('[data-testid="modal-backdrop"]').exists()).toBe(true)
    expect(wrapper.emitted('close')).toBeFalsy()
  })

  it('closes modal after editing existing task', async () => {
    const wrapper = mount(TaskModal, {
      props: {
        ...defaultProps,
        task: mockTask,
      },
    })

    const saveButton = wrapper.find('[data-testid="save-button"]')
    await saveButton.trigger('click')

    expect(wrapper.emitted('save')).toBeTruthy()
    expect(wrapper.emitted('close')).toBeTruthy()
  })

  it('handles escape key to close modal', async () => {
    const wrapper = mount(TaskModal, {
      props: defaultProps,
    })

    await wrapper.trigger('keydown.escape')

    expect(wrapper.emitted('close')).toBeTruthy()
  })

  it('has beautiful modal styling with backdrop', () => {
    const wrapper = mount(TaskModal, {
      props: defaultProps,
    })

    const panel = wrapper.find('[data-testid="modal-panel"]')
    expect(panel.classes()).toContain('bg-white')
    expect(panel.classes()).toContain('rounded-lg')
    expect(panel.classes()).toContain('shadow-xl')
  })

  it('validates description length constraints', async () => {
    const wrapper = mount(TaskModal, {
      props: defaultProps,
    })

    const descriptionField = wrapper.find('[data-testid="task-description"]')
    // Test with very long description (assuming max 500 characters)
    const longDescription = 'a'.repeat(501)
    await descriptionField.setValue(longDescription)

    const saveButton = wrapper.find('[data-testid="save-button"]')
    await saveButton.trigger('click')

    const errorMessage = wrapper.find('[data-testid="description-error"]')
    expect(errorMessage.exists()).toBe(true)
    expect(errorMessage.text()).toContain('Description is too long')
  })

  it('emits correct data when editing existing task', async () => {
    const wrapper = mount(TaskModal, {
      props: {
        ...defaultProps,
        task: mockTask,
      },
    })

    const descriptionField = wrapper.find('[data-testid="task-description"]')
    await descriptionField.setValue('Updated task description')

    const saveButton = wrapper.find('[data-testid="save-button"]')
    await saveButton.trigger('click')

    expect(wrapper.emitted('save')).toBeTruthy()
    expect(wrapper.emitted('save')![0][0]).toEqual({
      description: 'Updated task description',
      category: 'Work',
    })
  })
})