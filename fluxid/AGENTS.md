# Agent Guidelines for fluxid AppFactory

## Naming Conventions

**Company Name:** fluxid (all lowercase)
**Project Name:** fluxid AppFactory

## Glossary & Paths

Reference this section for all fluxid-specific terms and paths

**Product:**
- Location: `fluxid/product/` (analysis, requirements, context docs)
- Definition: The high level view on what is being built. Product defines user needs, features, and business goals. Contains product analysis, market context, and high-level requirements.

**Milestone:**
- Location: `fluxid/milestones/mXX-*.md`
- Definition: A vertical slice. A complete, demonstrable product capability state. When a milestone is done, the product has reached a specific functional level that can be run end-to-end.

**Epic:**
- Location: `fluxid/epics/mXX-eXX-*.md`
- Definition: One complete user journey through the milestone functionality, end-to-end testable

**Task:**
- Location: `fluxid/tasks/mXX-eXX-tXX-*.md`
- Definition: The implementation specification

**Template:**
- Location: `.fluxid/templates/task-template.md`
- Definition: Defines the structure of Task files (sections, fields, format)

**Registry/Field Registry:**
- Location: `fluxid/meta/{domain}/field-registry-{domain}.yaml`
- Definition: Canonical field names for a domain (e.g., flyer, pagination, user)
- Usage: Resolve field names from Cross-Layer Map
- **CRITICAL**: This file is the backbone of the fluxid build process. Always keep it up to date!

**Dependencies:**
- Location: Inside Task file under `Dependencies.Requires`
- Format: List of task IDs (e.g., `m01-e01-t01`)
- Action: Read `fluxid/tasks/{task-id}-*.md` to understand outputs

**Patterns:**
- Location: `.fluxid/patterns/flutter/` (or `/django/`, `/app/`)
- Definition: Example code showing correct implementation style

**Hooks:**
- Location: `.fluxid/hooks/{app|django|flutter}/`
- Definition: Quality gate scripts (linting, test checks, pre-commit)

**E2E Task:**
- Location: `fluxid/e2e/mXX-eXX-tXX-e2e-*.md`
- Definition: End-to-end test scenarios

**Progress Tracker:**
- Location: `fluxid/progress.yaml`
- Definition: Current status of all tasks/epics/milestones
