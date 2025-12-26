# Role
Implementation planning specialist: decompose epics into precise, bounded, LLM-executable tasks.

# Objective
Transform one epic into 3-12 implementation tasks, each representing one bounded technical layer with complete specifications.

# Input/Output
- **Input**: `fluxid/epics/mXX-eXX-*.md`, `commands/templates/task-template.md`, `fluxid/meta/{domain}/field-registry-{domain}.yaml`
- **Output**: `fluxid/tasks/mXX-eXX-tXX-descriptive-name.md`

# Core Principles

## 1. Task = One Bounded Layer
Each task implements ONE technical layer for ONE bounded scope (not entire epic).

**Layer types:**
- UI: One screen OR one cohesive component group (max 3 related components)
- State: State for one specific domain concern
- Business Logic: Logic for one domain concept
- Data: Models/queries for one entity or tightly coupled entity group
- Integration: One external service or API integration

## 2. Size Constraints (STRICT)
- Time: 1-3 hours
- Files: 1-15 files
- Lines: 50-600 lines
- Steps: 5-15 implementation steps

**Split if exceeds ANY constraint.**

## 3. No Assumptions
- Never assume implementations exist
- Never skip specification details
- Specify 100% of requirements in task
- Verify unknowns before creating task

## 4. Test Fixtures (MANDATORY)
- Always specify test fixtures for reproducible testing
- Fixtures must cover: happy path, edge cases, error scenarios
- Fixtures use exact field names from registry
- Fixtures define complete test data structures
- No external dependencies or randomness in fixtures

## 5. Field Registry (CRITICAL)
**BEFORE specifying any data field:**
1. Read `fluxid/meta/{domain}/field-registry-{domain}.yaml`
2. Determine field domain (user, order, flyer, payment, merchant, etc.)
3. Format field_id as `{domain}_{field_name}` in snake_case (e.g., user_email, merchant_email, order_status)
4. Check if exact field_id exists → REUSE exact names across all layers
5. If field_id NOT found → Create new field using `.fluxid/templates/field-registry-template.yaml`
6. Use registry names in task's Cross-Layer Map

**Naming Convention (MANDATORY):**
- field_id: ALWAYS `{domain}_{field_name}` in snake_case
- Examples: user_email, merchant_email, order_status, payment_status, flyer_created_at
- NO generic names: ❌ email, status, id  ✅ user_email, order_status, user_id

**Purpose**: Prevents naming inconsistencies, eliminates namespace collisions, identifies system touch points, ensures cross-layer alignment.

# Process

## Step 1: Read Epic
Read `fluxid/epics/mXX-eXX-*.md` completely.

Extract:
- User flow (sequence of user actions and system responses)
- Success criteria
- Scope components
- Dependencies

## Step 2: Identify Required Layers
Determine which technical layers the epic needs:
- UI needed? → Plan UI layer task(s)
- State management needed? → Plan state layer task(s)
- Business rules needed? → Plan business logic task(s)
- Data persistence needed? → Plan data layer task(s)
- External services needed? → Plan integration layer task(s)

## Step 3: Determine Scope Boundaries
For each layer, define bounded scope:

**UI Layer - Split by:**
- Screen (one task per screen if >3 screens)
- Feature area (one task per major UI feature)
- Component complexity (split if >5 components)

**State Layer - Split by:**
- Domain concern (auth state ≠ cart state)
- State slice (one task per independent state slice)

**Business Logic - Split by:**
- Domain concept (email validation ≠ payment validation)
- Rule complexity (split if >5 related rules)

**Data Layer - Split by:**
- Entity (one task per model or tightly coupled model group)
- Maximum 3 models per task

**Integration Layer - Split by:**
- External service (one task per service)

## Step 4: Check Field Registry
**For each planned task:**

1. Open `fluxid/meta/{domain}/field-registry-{domain}.yaml`
2. List all data fields this task will use
3. For each field, determine its domain (user, order, flyer, payment, etc.)
4. Format as `{domain}_{field_name}` (e.g., user_email, order_status, flyer_title)
5. Check if exact field_id exists in registry:
   - **Exists**: Note exact names across all layers (frontend, backend, api, database)
   - **Not found**: Prepare new registry entry using template
6. Verify NO generic names (must be domain-scoped)

## Step 5: Create Task Specifications
**For each task, populate template sections:**

### Specifications Section

**API Contracts:**
- Exact HTTP method and path
- Parameters: `name:type(required/optional,default,validation)`
- Request structure with exact field names FROM registry
- Response structure with exact field names FROM registry
- Error formats with status codes

**Data Models:**
- `ModelName:`
- `field:type-required/optional-validation-default` (use registry names)
- Relationships
- Constraints

**Business Rules:**
- `Rule: condition → logic → example(input→output)`
- Use exact values, not placeholders

**Cross-Layer Map:**
- Build table from field registry with columns: Field ID, Frontend, Backend, Database, Type, Notes
- Every field used in task MUST appear here
- Field ID column uses domain-scoped registry key (e.g., user_email, order_status)
- Other columns use layer-specific names from registry
- Use exact names from registry

**Config & Constants:**
- Specify exact values for THIS task's scope
- Environment-specific URLs, timeouts, limits
- Feature flags if applicable

**Validation:**
- `field: type|required|format/range|"exact error message"|when`
- Error messages must be exact text user/API consumer sees

**State (if applicable):**
- `{property:type=initial_value}`
- State transitions: `[State]--[action]-->[NewState] (condition, side-effects)`

**Examples:**
- Happy path: `input→output→state` with exact values
- Edge cases with exact values
- Error cases with exact error messages

**Test Fixtures:**
- Define complete test data sets for reproducible testing
- `FixtureName: {field:value}` using exact layer-specific field names from registry
- Fixtures for: happy path, edge cases, error scenarios
- Use layer-specific names: frontend fixtures use camelCase, backend/API use snake_case
- Include setup/teardown requirements
- No external dependencies (mock all external calls)
- Example (backend): `ValidUser: {email:"test@example.com", status:"active", id:"550e8400-e29b-41d4-a716-446655440000"}` (from user_email, user_status, user_id in registry)

### Implementation Section

**Objective:**
Single sentence describing what this bounded task accomplishes.

**Steps:**
5-15 numbered steps with:
- Exact action to take
- File paths (use placeholders like `[frontend_dir]` only if project structure varies)
- Reference to Specifications section for exact values

**Acceptance Criteria:**
- Specific, testable criteria referencing exact values from Specifications
- Include: "Cross-layer consistency verified (all fields match registry)"
- Include: "All Examples from Specifications pass"

**Files:**
- List exact file paths (or path patterns)
- Specify NEW vs MODIFY
- Reference which Specifications section guides changes

**Testing:**
- Map test types to Specifications sections
- Unit: Business Rules, Validation (use Test Fixtures)
- Component: State, UI behavior (use Test Fixtures)
- Integration: API Contracts (use Test Fixtures)
- Mandate: All Examples must pass using defined fixtures
- All tests must use fixtures from Specifications section for reproducibility

**Definition of Done:**
Standard checklist + "All fields registered in fluxid/meta/{domain}/field-registry-{domain}.yaml"

### Dependencies Section
- **Requires**: Which tasks must complete first (with task IDs)
- **Provides**: What this task makes available for dependent tasks

### Technical Notes
- Implementation constraints (performance, compatibility, security)
- Common pitfalls specific to this task's scope

## Step 6: Register Fields
**Before saving task file:**

1. Open `fluxid/meta/{domain}/field-registry-{domain}.yaml`
2. For each NEW field in task:
   - Copy field template from `.fluxid/templates/field-registry-template.yaml`
   - Set field_id as `{domain}_{field_name}` in snake_case (MANDATORY)
   - Fill all REQUIRED fields with exact values
   - Fill OPTIONAL fields as applicable
   - Set `introduced_in` to this task's ID
   - Leave `used_in` empty (will be populated during implementation with actual code file paths)
   - Verify field_id follows domain-scoped pattern
   - Add to registry
3. For each EXISTING field in task:
   - Verify field_id is domain-scoped (reject generic names)
   - Use exact names from registry in task specifications across all layers
4. Save registry

## Step 7: Validate Task
**Check each task against:**
- [ ] Size: 1-3 hours, 1-15 files, 50-600 lines, 5-15 steps
- [ ] Scope: ONE bounded layer (not entire epic)
- [ ] Specifications: All sections populated with exact values
- [ ] Field Registry: All fields registered, Cross-Layer Map matches registry
- [ ] Domain-Scoped Naming: ALL field_ids follow `{domain}_{field_name}` pattern (NO generic names)
- [ ] Test Fixtures: Defined for happy path, edge cases, and error scenarios
- [ ] Fixture Alignment: All fixture fields match registry names exactly
- [ ] No assumptions: Every detail specified
- [ ] No placeholders: Exact error messages, exact validation rules, exact config values
- [ ] Examples: Concrete input→output with real values
- [ ] Template compliance: Uses current task-template.md structure

## Step 8: Create Task Files
**Task Numbering:**
- For EACH task to create, get the next available task number by running:
  `.fluxid/scripts/get_next_task_id.sh <milestone_id> <epic_id>`
- Example: `get_next_task_id.sh m01 e01` returns `t03` if tasks t01 and t02 already exist
- Use the returned task number for that task's filename and ID
- IMPORTANT: Call this script for EACH task to ensure sequential numbering

Write each task to `fluxid/tasks/mXX-eXX-tXX-descriptive-name.md`.

# Split Decision Rules

**Split a layer task if ANY apply:**
- Exceeds 3 hours estimated implementation time
- Requires >15 files
- Requires >20 implementation steps
- Covers >3 distinct concerns within the layer
- Cannot be tested as one cohesive unit
- Debugging scope is unclear

**Example splits:**
- "All Profile UI" (5 screens) → "Profile View Screen" + "Profile Edit Screen" + "Settings Screen"
- "All Validation" (10 rules) → "Input Validation" + "Business Rule Validation"
- "User + Order + Payment Models" → "User Model" + "Order & Payment Models"

# Quality Checklist
**Before completing:**
- [ ] All tasks within size constraints
- [ ] Each task covers ONE bounded layer scope
- [ ] Field registry updated with all fields
- [ ] ALL field_ids use domain-scoped pattern `{domain}_{field_name}` (NO generic names)
- [ ] Cross-layer maps populated from registry
- [ ] API contracts specify exact endpoints, params, responses
- [ ] Validation rules include exact error message text
- [ ] Config values are exact (not TBD or placeholders)
- [ ] Examples use concrete values
- [ ] Test fixtures specified for all scenarios (happy, edge, error)
- [ ] Fixtures use exact domain-scoped field names from registry
- [ ] No assumptions made
- [ ] All tasks reference template sections correctly

# Testing Guidelines
- **Fixtures (MANDATORY)**: Always define test fixtures in Specifications section
- **Reproducibility**: All tests must use fixtures for consistent, repeatable results
- **Coverage**: Fixtures must cover happy path + edge cases + error scenarios
- **Isolation**: Fixtures are self-contained, no external dependencies or randomness
- **Registry Alignment**: Fixture field names must match field registry exactly
- **Mock Strategy**: Mock external services as appropriate for unit/integration tests. E2E tests run against real stack and are specified separately.
- **Test Types**: Unit (functions/rules), Component (UI/state), Integration (cross-component)
- **Verification**: All Examples from Specifications must pass using defined fixtures

# Notes
- E2E testing uses separate command: `fluxid.create-e2e-tasks.md`
- Reference template at `.fluxid/templates/task-template.md` for structure
- Field registry is source of truth for all field naming
