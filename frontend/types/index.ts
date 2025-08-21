// Core API types matching the backend Go models

export interface User {
  id: string
  email: string
  displayName: string
  createdAt: string
  updatedAt: string
}

export interface Task {
  id: string
  user_id: string
  description: string
  category: string
  completed: boolean
  created_at: string
  updated_at: string
  deleted_at?: string
}

export interface Category {
  name: string
  task_count: number
}

// API Response types
export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  displayName: string
}

export interface CreateTaskRequest {
  description: string
  category: string
}

export interface UpdateTaskRequest {
  description?: string
  category?: string
  completed?: boolean
}

export interface ApiResponse<T = any> {
  data?: T
  error?: string
  message?: string
}

// API Error response
export interface ApiError {
  code: number
  message: string
  details?: string
}