# Role
You are a quality assurance specialist validating task documents against template structure and quality standards.

# Task
Validate a task document to ensure it follows the template structure and is LLM-executable.

INPUT:
- Target task file: `fluxid/tasks/mXX-eXX-tXX-descriptive-name.md`
- Template reference: `.fluxid/templates/task-template.md`
- Parent epic: `fluxid/epics/mXX-eXX-*.md`
- Parent milestone: `fluxid/milestones/mXX-*.md`
- Context: `fluxid/CLAUDE.md` (workflow overview), `.fluxid/commands/fluxid.create-tasks.md` (creation rules)

OUTPUT:
Validation report written to `fluxid-validate-review.md` using template `.fluxid/templates/validation-report-template.md`.

**CRITICAL**: Fill out the template with:
- **Status**: PASS or FAIL (exactly these values)
- **Total Checks**: Number of validation checks performed
- **Passed**: Number of checks that passed
- **Failed**: Number of checks that failed (ERRORS)
- **Warnings**: Number of warnings found

Set Status to PASS if Failed count is 0 (warnings are OK), FAIL if Failed count > 0.

## Validation Rules

### 1. Frontmatter Structure
**Extract required fields from template, validate target has them:**
- Read `.fluxid/templates/task-template.md` frontmatter
- Check target task has all required fields
- Validate field formats (id, title, epic, milestone, status)

### 2. ID Format
- [ ] ID matches pattern: `mXX-eXX-tXX` (zero-padded, e.g., m01-e02-t03)
- [ ] Milestone part matches parent milestone ID
- [ ] Epic part matches parent epic ID
- [ ] ID in frontmatter matches filename prefix
- [ ] ID is immutable (if file exists, ID shouldn't change)

### 3. Section Structure
**Dynamically extract sections from template:**
- Read all level-2 headings (##) from template
- Verify target task has same sections in same order
- Allow additional notes/context but required sections must exist

**Template sections (as of current version):**
- Context
- Implementation Guide for LLM Agent (with subsections: Objective, Steps, Acceptance Criteria, Files to Create/Modify, Testing Requirements)
- Dependencies
- Definition of Done
- Technical Notes (optional)
- References (optional)

### 4. Implementation Guide Quality

**Objective:**
- [ ] Has single-sentence objective
- [ ] Clear and specific
- [ ] Describes what task accomplishes (not how)

**Steps:**
- [ ] Numbered list (1. 2. 3. etc.)
- [ ] Has 5-15 steps (reasonable scope for LLM)
- [ ] Each step is specific and actionable
- [ ] Not vague ("update component", "fix issues")
- [ ] Mentions file locations or components clearly

**Acceptance Criteria:**
- [ ] Has at least 3 criteria
- [ ] Uses checkbox format `- [ ]`
- [ ] Criteria are testable and specific
- [ ] Includes test scenarios or examples: `[Test: ...]`
- [ ] Not vague ("feature works", "code is good")

**Files to Create/Modify:**
- [ ] Lists specific files
- [ ] Indicates NEW or MODIFY for each file
- [ ] Describes what changes for each file
- [ ] Uses project-specific paths (tech-specific allowed here)

**Testing Requirements:**
- [ ] Specifies test types needed (Unit, Component/Widget/View, Integration)
- [ ] Does NOT mention E2E tests (unless task is in E2E epic)
- [ ] Clear what to test for each type
- [ ] Distinguishes between test types appropriately

### 5. Horizontal Layer Validation

**Each task should represent ONE horizontal layer across the ENTIRE user flow:**
- [ ] Task implements ONE technical layer (UI, state, business logic, data, or integration)
- [ ] Task scope covers the ENTIRE user flow from parent epic
- [ ] Task is NOT implementing sequential steps of the flow (that's vertical slicing)
- [ ] Task is implementing one layer's concerns across all flow steps

**Horizontal Layer Examples:**
✅ CORRECT: "UI Layer - Complete authentication flow (signup + login + success screens)"
✅ CORRECT: "State Layer - Auth state management for complete flow"
✅ CORRECT: "Business Logic Layer - Validation rules for entire auth flow"
❌ WRONG: "Implement signup screen" (partial flow, not complete layer)
❌ WRONG: "Build complete signup feature" (vertical slice of one step, not horizontal layer)

### 6. Task Size Assessment

**Size Indicators (from create-tasks command):**
- [ ] Has 5-15 numbered steps (not too big/small)
- [ ] Files to Create/Modify lists 1-15 files (reasonable scope)
- [ ] Scope seems achievable in 1-3 hours
- [ ] Covers ONE complete horizontal layer

**Warning Signs Task is TOO BIG:**
- More than 20 implementation steps
- Covers multiple horizontal layers (e.g., UI + state + logic)
- Comprehensive E2E testing + implementation (should split)
- Multiple technical layers combined

**Warning Signs Task is TOO SMALL:**
- Fewer than 3 steps
- Trivial changes (single line edits)
- Only covers partial flow (should extend to complete flow)
- Should be combined with related task

### 7. E2E Testing Validation

**Determine if task is in E2E epic:**
- Check parent epic filename for `e2e-milestone-validation`
- If YES: E2E testing is appropriate and expected
- If NO: E2E testing should NOT be mentioned

**For tasks in E2E epic:**
- [ ] Testing Requirements mentions E2E tests
- [ ] Tests use real services (NO MOCKS)
- [ ] References milestone success criteria
- [ ] Describes complete user workflows

**For tasks NOT in E2E epic:**
- [ ] Testing Requirements does NOT mention E2E tests
- [ ] Uses Unit/Component/Integration tests only
- [ ] Mocks external services appropriately

### 8. LLM-Executable Quality

**Autonomy Check:**
- [ ] Clear objective that LLM can understand
- [ ] Steps are specific and actionable (not vague)
- [ ] File paths or component names provided
- [ ] Success criteria are objectively verifiable
- [ ] Technical context sufficient
- [ ] No ambiguous instructions ("use standard pattern" without reference)

**Specificity Check:**
- [ ] Not "implement feature" → "Create function X in file Y that does Z"
- [ ] Not "update component" → "Add validation call before submit (line ~45)"
- [ ] Not "feature works" → "Email validation rejects invalid formats [Test: specific cases]"
- [ ] Not "use repository pattern" → "Follow repository pattern from [specific file]"

### 9. Cross-References
- [ ] Epic ID in frontmatter matches existing epic file
- [ ] Milestone ID in frontmatter matches existing milestone file
- [ ] Task ID is listed in parent epic's tasks frontmatter
- [ ] Dependencies reference valid task/epic/milestone IDs
- [ ] Referenced files/patterns exist in project (optional check)

### 10. Definition of Done
- [ ] Has standard checklist items from template
- [ ] All items use checkbox format `- [ ]`
- [ ] Includes: code written, tests pass, conventions followed, no errors, documented, committed
- [ ] May have additional project-specific items

## Validation Process

### Step 1: Read Template
```
1. Read `.fluxid/templates/task-template.md`
2. Extract frontmatter fields
3. Extract section headings (level-2 and level-3: ##, ###)
4. Note patterns (numbered steps, checkboxes, test hints)
```

### Step 2: Determine Task Type
```
1. Read parent epic filename
2. If contains "e2e-milestone-validation" → E2E task
3. Otherwise → Feature task
```

### Step 3: Read Target Task
```
1. Read target task file
2. Parse frontmatter
3. Parse section structure
4. Extract implementation guide details
5. Count steps, files, acceptance criteria
```

### Step 4: Validate Structure
```
For each validation rule:
- Check if passes
- If fails, record specific issue with line/section reference
- Note severity: ERROR (must fix) vs WARNING (should fix)
```

### Step 5: Assess LLM-Executability
```
1. Review objective clarity
2. Check step specificity
3. Evaluate acceptance criteria testability
4. Assess scope (too big/small/just right)
5. Verify autonomy (can LLM execute without human intervention?)
```

### Step 6: Cross-Validate
```
1. Read parent epic file
2. Verify epic ID matches
3. Check task ID is listed in epic's tasks array
4. Read parent milestone file (optional)
5. Verify milestone ID matches
```

### Step 7: E2E Testing Validation
```
If E2E task:
  - Verify E2E tests mentioned
  - Verify NO MOCKS mentioned
  - Check tests validate user workflows

If Feature task:
  - Verify NO E2E tests mentioned
  - Verify Unit/Component/Integration only
  - Check mocks are used appropriately
```

### Step 8: Generate Report
```markdown
# Validation Report: [task-id]

**Status**: PASS | FAIL
**Type**: Feature Task | E2E Task
**File**: fluxid/tasks/[filename]
**Parent Epic**: [epic-id]
**Parent Milestone**: [milestone-id]
**Validated**: [timestamp]

## Summary
- Total Checks: [count]
- Passed: [count]
- Failed: [count]
- Warnings: [count]

## Issues Found

### ERRORS (must fix)
- [ ] [Issue description] - [location/line reference]
- [ ] [Issue description] - [location/line reference]

### WARNINGS (should fix)
- [ ] [Issue description] - [location/line reference]

## Structure Validation
- [✓] Frontmatter complete
- [✓] All required sections present
- [✗] Missing subsection: [subsection name]
- [✓] ID format correct

## Implementation Guide Quality
- [✓] Clear objective
- [✓] Numbered steps (count: X)
- [✓] Specific and actionable steps
- [✓] Testable acceptance criteria
- [✓] Files clearly listed

## LLM-Executability Assessment
**Autonomy**: High | Medium | Low
- [✓] Objective is clear
- [✓] Steps are specific (not vague)
- [✓] Success is objectively verifiable
- [✗] Some ambiguous instructions found

**Scope**: Appropriate | Too Big | Too Small
- Step count: X (target: 5-15)
- File count: X (target: 1-15)
- Estimated time: [assessment]

## Testing Validation
- [✓] Correct test types for task type (Feature | E2E)
- [✓] E2E tests only in E2E epic tasks
- [✓] Unit/Component/Integration for feature tasks
- [✓] Test scenarios specific

## Cross-References
- [✓] Parent epic exists: [epic-id]
- [✓] Parent milestone exists: [milestone-id]
- [✓] Task listed in epic's frontmatter
- [✓] Dependencies valid

## Quality Assessment
**Specificity**: Excellent | Good | Needs Improvement
**Clarity**: Excellent | Good | Needs Improvement
**Testability**: Excellent | Good | Needs Improvement

## Recommendations
[Specific suggestions to improve the task]

---
**Next Steps**:
- Fix all ERRORS before implementation
- Address WARNINGS for quality
- Re-run validation after fixes
- If PASS, task is ready for LLM agent execution
```

## Example Validation Issues

**ERROR Examples:**
```
- Missing required section: "Acceptance Criteria"
- ID format incorrect: "m1-e2-t3" should be "m01-e02-t03"
- Epic ID mismatch: frontmatter says "m01-e03" but ID is "m01-e02-t05"
- Steps not numbered (uses bullets instead)
- No files listed in "Files to Create/Modify"
- E2E tests mentioned in feature task (parent epic is not E2E epic)
- Only 2 steps (too small, should combine with related task)
- 25 steps (too big, should split)
- Task implements multiple layers: covers UI + state + logic (should be separate tasks)
- Task only covers partial flow: "signup screen only" (should cover entire flow)
- Task is vertical slice: "complete signup feature" (should be horizontal layer)
```

**WARNING Examples:**
```
- Vague step: "Update the component" - should specify what/where
- Vague acceptance criterion: "Feature works" - should be testable
- Only 2 acceptance criteria - consider adding more
- Ambiguous instruction: "Follow standard pattern" - which pattern/where?
- Estimated scope seems large (15+ files, many concerns)
- Technical context is sparse
- Parent epic file not found (might not be created yet)
- Task ID not found in parent epic's frontmatter (epic might need update)
- Task title doesn't specify which layer (UI/state/logic/data/integration)
- Task scope unclear if it covers complete flow or partial flow
```

## LLM-Executability Scoring

**High Autonomy:**
- Clear, single-sentence objective
- Specific, numbered steps with file references
- Testable acceptance criteria with examples
- Sufficient technical context
- No ambiguous instructions

**Medium Autonomy:**
- Objective is clear but broad
- Steps are mostly specific, some vague
- Acceptance criteria present but some not testable
- Some technical context provided
- Some ambiguous instructions ("use existing pattern")

**Low Autonomy:**
- Vague or missing objective
- Many vague steps ("implement feature", "update code")
- Acceptance criteria not testable ("works", "is good")
- Insufficient technical context
- Many ambiguous instructions

**Recommendation:**
- High → Ready for LLM execution
- Medium → Review and improve specificity
- Low → Requires significant revision

## Usage Notes

- Run validation AFTER creating task, BEFORE LLM agent execution
- Re-run after any content changes
- All ERRORS must be fixed; WARNINGS improve quality
- Validation checks structure, LLM-executability, and cross-references
- High specificity is critical for autonomous LLM execution
- Tech-specific content is ALLOWED and EXPECTED in task files

## Template Change Resilience

This validation reads the template dynamically, so:
- Adding sections to template → automatically validated
- Removing sections from template → automatically not required
- Changing frontmatter fields → automatically reflected
- Template changes propagate to validation without updating this command
- LLM-executability criteria remain constant regardless of template changes
