# Task Tracker Frontend Makefile

.PHONY: help dev build test screenshots clean

# Default target
help:
	@echo "Available commands:"
	@echo "  make dev          - Start development server"
	@echo "  make build        - Build for production"
	@echo "  make test         - Run all tests"
	@echo "  make screenshots  - Capture application screenshots"
	@echo "  make clean        - Clean build artifacts and screenshots"

# Development
dev:
	npm run dev

# Build
build:
	npm run build

# Testing
test:
	npm run test

# Screenshots
screenshots:
	@echo "🚀 Starting screenshot capture..."
	@echo "📋 Ensure frontend (localhost:3000) and backend (localhost:8080) are running"
	npm run screenshots

screenshots-headless:
	@echo "🚀 Starting headless screenshot capture..."
	npm run screenshots:headless

# Cleanup
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf dist/
	rm -rf .nuxt/
	rm -rf node_modules/.cache/
	@echo "🧹 Cleaning screenshots..."
	rm -rf screenshots/
	rm -f ../SCREENSHOTS.md
	@echo "✅ Cleanup complete"

# Setup
install:
	npm install
	npx playwright install

# Quick setup for development
setup: install
	@echo "✅ Setup complete. You can now run:"
	@echo "   make dev          - Start development"
	@echo "   make screenshots  - Capture UI screenshots"