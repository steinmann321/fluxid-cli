# Role
Generate the master progress tracking file by reading all milestones and epics, and writing a lightweight status map.

# Task
Create `fluxid/progress.yaml` with structured YAML hierarchy (Milestone → Epic) using filename-based artifacts and explicit status fields.

# Process

1. Read all milestone files: `fluxid/milestones/*.md`
2. Read all epic files: `fluxid/epics/*.md`
3. Build progress YAML with milestones and epics only (no tasks)
4. Write to: `fluxid/progress.yaml`

# Output Structure

```yaml
project: YourProject
last_updated: "YYYY-MM-DD"

milestones:
  - id: m01
    artifact: "m01-single-user-vocabulary-practice"
    status: pending
    epics:
      - id: m01-e01
        artifact: "m01-e01-user-creates-ai-generated-vocabulary-list"
        status: pending
      - id: m01-e02
        artifact: "m01-e02-user-practices-vocabulary-with-flashcards"
        status: pending

  - id: m02
    artifact: "m02-multi-list-management"
    status: pending
    epics: []
```

# Rules
- All `status` fields start as `pending`
- Use milestone/epic IDs from frontmatter or filenames:
  - `milestones/m01-*.md` → `id: m01`, `artifact` from filename stem
  - `epics/m01-e01-*.md` → `id: m01-e01`, `artifact` from filename stem
- Include all milestones found in `fluxid/milestones/`
- Set `last_updated` to today's date (YYYY-MM-DD format)
- Keep structure clean and consistent
- Do **not** add tasks or deliverables to this file (milestones/epics only)

# Notes
This YAML file is the single source of truth for project progress. It's parsed using `yq` for robust, error-free progress tracking and validated conceptually against `.fluxid/templates/progress-schema.yaml`.
