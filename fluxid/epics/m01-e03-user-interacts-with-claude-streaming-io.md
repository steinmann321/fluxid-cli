---
id: m01-e03
title: User interacts with Claude via real-time streamed I/O
milestone: m01
status: pending
patterns: []  # Implementation order: foundational→middle→presentation layers
---

# Epic: User interacts with Claude via real-time streamed I/O

## Overview
User runs `fluxid --claude` → Claude prompts during a phase → User types input in the terminal → System forwards stdin to Claude while streaming stdout/stderr back in real time → Phase completes without truncation or lag → Workflow continues.

## Scope
- User actions: Observe output; provide input when prompted; continue monitoring
- System responses: Maintain bidirectional piping; ensure low-latency stdout/stderr passthrough; accept user stdin; handle partial lines and bursts; preserve ordering between streams
- Screens/states: Terminal streaming state; prompt moments; input echo; continued progress

## Success Criteria
- [ ] Real-time stdout/stderr passthrough with acceptable latency [Test: generate burst output; verify interleaving and timing under <200ms average latency]
- [ ] Stdin from user is delivered to Claude reliably [Test: prompt→respond interaction; assert Claude receives exact input bytes]
- [ ] No output truncation or buffer deadlocks [Test: long outputs; backpressure simulation]
- [ ] Stream ordering is sensible and readable [Test: mixed stdout/stderr; ensure readable interleaving]
- [ ] User can complete interactive phase and see workflow continue [Test: after input, next phase starts; output confirms]

## Dependencies
- Depends on `m01-e01` (core orchestration and streaming setup)
- External prerequisite: Claude CLI available on PATH

## E2E Test Mapping
**Test File**: `m01_e03_t01_user-interacts-streaming-io.yaml`

**Test Flow**:
1. Invoke `fluxid --claude` with a stubbed interactive Claude phase
2. Observe prompt; type input
3. Verify input reaches child process and output continues streaming
4. Assert workflow continues post-interaction

**Key Assertions**:
- Low-latency streaming observed
- Input echoed and consumed by Claude
- No deadlocks or truncation

## Completion Checklist
- [ ] All success criteria validated
- [ ] E2E test exists and passes
- [ ] Epic contributes to milestone
- [ ] No regressions
- [ ] ONE atomic flow (not multiple flows)

## Notes
- Stream handling must be resilient to variable output rates and interactive pauses