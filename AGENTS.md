# AGENTS.md

## Agent Configuration

This file specifies the agent configuration for Claude Code when working on this project.

### Required Agents

When starting a conversation - ensure all agents in .claude/agents/ are utilized.

### Usage Guidelines

- Use multiple agents concurrently when tasks can be parallelized
- Launch agents proactively when their specialized capabilities match the task
- The general-purpose agent should be used for:
  - Open-ended searches requiring multiple rounds of globbing and grepping
  - Complex multi-step research tasks
  - Understanding large codebases
  - Finding specific implementations across many files

### Agent Communication

- Each agent invocation is stateless
- Provide detailed task descriptions to agents
- Specify exactly what information should be returned
- Trust agent outputs as they have been optimized for accuracy