import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import TaskFilters from '../../components/TaskFilters.vue'
import type { Category } from '../../types'

// Mock dependencies
vi.mock('@headlessui/vue', () => ({
  Listbox: {
    name: 'Listbox',
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
  Switch: {
    template: '<button type="button" :class="modelValue ? \'bg-indigo-600\' : \'bg-gray-200\'"><slot /></button>',
    props: ['modelValue'],
    emits: ['update:modelValue'],
  },
  SwitchGroup: {
    template: '<div><slot /></div>',
  },
  SwitchLabel: {
    template: '<label><slot /></label>',
  },
}))

// Mock Heroicons
vi.mock('@heroicons/vue/24/outline', () => ({
  ChevronUpDownIcon: {
    template: '<svg></svg>',
  },
  CheckIcon: {
    template: '<svg></svg>',
  },
}))

describe('TaskFilters', () => {
  const mockCategories: Category[] = [
    { name: 'Work', task_count: 5 },
    { name: 'Personal', task_count: 3 },
    { name: 'Shopping', task_count: 2 },
  ]

  const defaultProps = {
    categories: mockCategories,
    selectedCategory: 'All',
    showCompleted: true,
    showDeleted: false,
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders correctly with default props', () => {
    const wrapper = mount(TaskFilters, {
      props: defaultProps,
    })

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.find('[data-testid="category-filter"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="show-completed-toggle"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="show-deleted-toggle"]').exists()).toBe(true)
  })

  it('displays "All Categories" option plus user categories in dropdown', () => {
    const wrapper = mount(TaskFilters, {
      props: defaultProps,
    })

    // Should show current selection
    expect(wrapper.text()).toContain('All Categories')
    
    // When opened, should show all options including "All"
    const categoryOptions = wrapper.findAll('[data-testid="category-option"]')
    expect(categoryOptions.length).toBeGreaterThanOrEqual(1) // At least "All" option
  })

  it('emits category-change event when category is selected', async () => {
    const wrapper = mount(TaskFilters, {
      props: defaultProps,
    })

    // Simulate category selection
    await wrapper.findComponent({ name: 'Listbox' }).vm.$emit('update:modelValue', 'Work')
    
    expect(wrapper.emitted('category-change')).toBeTruthy()
    expect(wrapper.emitted('category-change')![0]).toEqual(['Work'])
  })

  it('displays correct category name based on selectedCategory prop', () => {
    const wrapper = mount(TaskFilters, {
      props: {
        ...defaultProps,
        selectedCategory: 'Work',
      },
    })

    expect(wrapper.text()).toContain('Work')
  })

  it('shows task counts for each category', () => {
    const wrapper = mount(TaskFilters, {
      props: defaultProps,
    })

    // Should display categories with their task counts
    expect(wrapper.html()).toMatch(/Work.*5/)
    expect(wrapper.html()).toMatch(/Personal.*3/)
    expect(wrapper.html()).toMatch(/Shopping.*2/)
  })

  it('emits toggle-completed event when show completed toggle is clicked', async () => {
    const wrapper = mount(TaskFilters, {
      props: defaultProps,
    })

    const completedToggle = wrapper.findComponent('[data-testid="show-completed-toggle"]')
    await completedToggle.vm.$emit('update:modelValue', false)

    expect(wrapper.emitted('toggle-completed')).toBeTruthy()
    expect(wrapper.emitted('toggle-completed')![0]).toEqual([false])
  })

  it('emits toggle-deleted event when show deleted toggle is clicked', async () => {
    const wrapper = mount(TaskFilters, {
      props: defaultProps,
    })

    const deletedToggle = wrapper.findComponent('[data-testid="show-deleted-toggle"]')
    await deletedToggle.vm.$emit('update:modelValue', true)

    expect(wrapper.emitted('toggle-deleted')).toBeTruthy()
    expect(wrapper.emitted('toggle-deleted')![0]).toEqual([true])
  })

  it('shows correct toggle states based on props', () => {
    const wrapper = mount(TaskFilters, {
      props: {
        ...defaultProps,
        showCompleted: false,
        showDeleted: true,
      },
    })

    const completedToggle = wrapper.findComponent('[data-testid="show-completed-toggle"]')
    const deletedToggle = wrapper.findComponent('[data-testid="show-deleted-toggle"]')

    expect(completedToggle.props('modelValue')).toBe(false)
    expect(deletedToggle.props('modelValue')).toBe(true)
  })

  it('handles empty categories array', () => {
    const wrapper = mount(TaskFilters, {
      props: {
        ...defaultProps,
        categories: [],
      },
    })

    expect(wrapper.exists()).toBe(true)
    // Should still show "All Categories" option
    expect(wrapper.text()).toContain('All Categories')
  })

  it('displays toggle labels correctly', () => {
    const wrapper = mount(TaskFilters, {
      props: defaultProps,
    })

    expect(wrapper.text()).toContain('Show completed tasks')
    expect(wrapper.text()).toContain('Show deleted tasks')
  })

  it('applies correct styling to toggles based on state', () => {
    const wrapper = mount(TaskFilters, {
      props: {
        ...defaultProps,
        showCompleted: true,
        showDeleted: false,
      },
    })

    const completedToggle = wrapper.findComponent('[data-testid="show-completed-toggle"]')
    const deletedToggle = wrapper.findComponent('[data-testid="show-deleted-toggle"]')

    // Active toggle should have indigo background
    expect(completedToggle.classes()).toContain('bg-indigo-600')
    // Inactive toggle should have gray background
    expect(deletedToggle.classes()).toContain('bg-gray-200')
  })

  it('uses compact design suitable for top of task list', () => {
    const wrapper = mount(TaskFilters, {
      props: defaultProps,
    })

    // Should have horizontal layout and compact spacing
    const container = wrapper.find('[data-testid="filters-container"]')
    expect(container.exists()).toBe(true)
    expect(container.classes()).toContain('flex')
  })
})