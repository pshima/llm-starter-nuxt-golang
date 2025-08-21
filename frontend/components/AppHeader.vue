<template>
  <header class="bg-white shadow-sm sticky top-0 z-40">
    <nav class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
      <div class="flex h-16 items-center justify-between">
        <!-- App Name/Logo -->
        <div class="flex items-center">
          <h1 class="text-xl font-bold text-gray-900">Task Tracker</h1>
        </div>

        <!-- Desktop Navigation -->
        <div class="hidden md:block" data-testid="desktop-navigation">
          <div class="ml-10 flex items-baseline space-x-4">
            <a
              href="/dashboard"
              @click.prevent="navigateTo('/dashboard')"
              data-testid="nav-dashboard"
              class="rounded-md px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 hover:text-gray-900 transition-colors"
              :class="$route.path === '/dashboard' ? 'bg-gray-100 text-gray-900' : ''"
            >
              Dashboard
            </a>
            <a
              href="/tasks"
              @click.prevent="navigateTo('/tasks')"
              data-testid="nav-tasks"
              class="rounded-md px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 hover:text-gray-900 transition-colors"
              :class="$route.path === '/tasks' ? 'bg-gray-100 text-gray-900' : ''"
            >
              Tasks
            </a>
          </div>
        </div>

        <!-- Desktop Profile Dropdown -->
        <div class="hidden md:block" v-if="authStore.isAuthenticated && authStore.user">
          <div class="ml-4 flex items-center md:ml-6">
            <Menu as="div" class="relative ml-3">
              <div>
                <MenuButton
                  data-testid="user-menu-button"
                  class="relative flex max-w-xs items-center rounded-full bg-white text-sm text-gray-700 hover:text-gray-900 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 transition-colors"
                >
                  <span class="sr-only">Open user menu</span>
                  <div class="flex items-center space-x-2">
                    <div class="h-8 w-8 rounded-full bg-indigo-500 flex items-center justify-center">
                      <span class="text-sm font-medium text-white">
                        {{ userInitials }}
                      </span>
                    </div>
                    <span class="hidden lg:block">{{ authStore.user.displayName }}</span>
                    <ChevronDownIcon class="h-4 w-4" aria-hidden="true" />
                  </div>
                </MenuButton>
              </div>
              <transition
                enter-active-class="transition ease-out duration-100"
                enter-from-class="transform opacity-0 scale-95"
                enter-to-class="transform opacity-100 scale-100"
                leave-active-class="transition ease-in duration-75"
                leave-from-class="transform opacity-100 scale-100"
                leave-to-class="transform opacity-0 scale-95"
              >
                <MenuItems
                  data-testid="user-dropdown"
                  class="absolute right-0 z-10 mt-2 w-48 origin-top-right rounded-md bg-white py-1 shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none"
                >
                  <MenuItem v-slot="{ active }">
                    <button
                      @click="handleLogout"
                      data-testid="logout-button"
                      :class="[
                        active ? 'bg-gray-100' : '',
                        'block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 transition-colors'
                      ]"
                    >
                      Logout
                    </button>
                  </MenuItem>
                </MenuItems>
              </transition>
            </Menu>
          </div>
        </div>

        <!-- Mobile menu button -->
        <div class="md:hidden">
          <button
            @click="toggleMobileMenu"
            data-testid="mobile-menu-button"
            type="button"
            class="relative inline-flex items-center justify-center rounded-md p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-indigo-500"
            aria-controls="mobile-menu"
            aria-expanded="false"
          >
            <span class="sr-only">Open main menu</span>
            <Bars3Icon v-if="!mobileMenuOpen" class="block h-6 w-6" aria-hidden="true" />
            <XMarkIcon v-else class="block h-6 w-6" aria-hidden="true" />
          </button>
        </div>
      </div>

      <!-- Mobile menu -->
      <div 
        data-testid="mobile-menu"
        :class="mobileMenuOpen ? 'block' : 'hidden'"
        class="md:hidden"
        id="mobile-menu"
      >
        <div class="space-y-1 px-2 pb-3 pt-2 sm:px-3">
          <button
            @click="navigateTo('/dashboard'); mobileMenuOpen = false"
            class="block rounded-md px-3 py-2 text-base font-medium text-gray-700 hover:bg-gray-100 hover:text-gray-900 w-full text-left transition-colors"
            :class="$route.path === '/dashboard' ? 'bg-gray-100 text-gray-900' : ''"
          >
            Dashboard
          </button>
          <button
            @click="navigateTo('/tasks'); mobileMenuOpen = false"
            class="block rounded-md px-3 py-2 text-base font-medium text-gray-700 hover:bg-gray-100 hover:text-gray-900 w-full text-left transition-colors"
            :class="$route.path === '/tasks' ? 'bg-gray-100 text-gray-900' : ''"
          >
            Tasks
          </button>
        </div>
        
        <!-- Mobile Profile Section -->
        <div class="border-t border-gray-200 pb-3 pt-4" v-if="authStore.isAuthenticated && authStore.user">
          <div class="flex items-center px-5">
            <div class="h-10 w-10 rounded-full bg-indigo-500 flex items-center justify-center">
              <span class="text-sm font-medium text-white">
                {{ userInitials }}
              </span>
            </div>
            <div class="ml-3">
              <div class="text-base font-medium text-gray-800">{{ authStore.user.displayName }}</div>
              <div class="text-sm text-gray-500">{{ authStore.user.email }}</div>
            </div>
          </div>
          <div class="mt-3 space-y-1 px-2">
            <button
              @click="handleLogout"
              class="block rounded-md px-3 py-2 text-base font-medium text-gray-700 hover:bg-gray-100 hover:text-gray-900 w-full text-left transition-colors"
            >
              Logout
            </button>
          </div>
        </div>
      </div>
    </nav>
  </header>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Menu, MenuButton, MenuItem, MenuItems } from '@headlessui/vue'
import { Bars3Icon, XMarkIcon, ChevronDownIcon } from '@heroicons/vue/24/outline'

const authStore = useAuthStore()
const mobileMenuOpen = ref(false)

const userInitials = computed(() => {
  if (!authStore.user?.displayName) return 'U'
  
  const names = authStore.user.displayName.split(' ')
  if (names.length === 1) {
    return names[0].charAt(0).toUpperCase()
  }
  return (names[0].charAt(0) + names[names.length - 1].charAt(0)).toUpperCase()
})

const toggleMobileMenu = () => {
  mobileMenuOpen.value = !mobileMenuOpen.value
}

const handleLogout = async () => {
  await authStore.logout()
  mobileMenuOpen.value = false
  navigateTo('/login')
}
</script>