---
id: m06
title: Users can test workflows with dry-run and control output format
status: pending
---

# Milestone: Users can test workflows with dry-run and control output format

## Deliverable
Users can test workflow configuration and loop logic without invoking actual agents via `--fluxid-dry-run` flag (simulates execution, prints what would happen, generates synthetic PASS reports for loop testing). Users can also control initialization status output format via `--fluxid-output {json|yaml}` for programmatic parsing in automation scenarios.

**What users can do:**
- Test configuration with `fluxid --fluxid-dry-run --claude [args]`
- See simulated workflow execution: which phases would run, loop progression, command files used
- Validate configuration files and command file references without running actual agent
- Test loop logic with synthetic PASS reports (verify loop counts, phase order)
- Get initialization status in JSON format: `fluxid --fluxid-output json --claude [args]`
- Get initialization status in YAML format: `fluxid --fluxid-output yaml --claude [args]`
- Use structured output for automation scripts that parse fluxid status

## Success Criteria
- [ ] Dry-run flag implemented: `--fluxid-dry-run`
  - [ ] Performs full configuration resolution and validation
  - [ ] Checks command files exist and are readable
  - [ ] Validates loop counts and configuration values
  - [ ] Prints simulated execution plan:
    - [ ] "Would execute: Iteration 1, Retry 1, Phase: IMPLEMENT"
    - [ ] "Would execute: Iteration 1, Retry 1, Phase: COMMIT"
    - [ ] Shows which command file would be used for each phase
  - [ ] Skips actual agent process spawning
  - [ ] Generates synthetic PASS reports to simulate loop progression
  - [ ] Allows full loop testing without agent invocations
  - [ ] Exits with success code when simulation completes
- [ ] Output format flag implemented: `--fluxid-output {json|yaml}`
  - [ ] Controls initialization status output format only (not agent output)
  - [ ] Default: human-readable text format
  - [ ] JSON format: structured JSON object with all config values
  - [ ] YAML format: structured YAML document with all config values
  - [ ] Includes: session ID, agent selection, loop counts, command file paths, phase toggles
  - [ ] Useful for automation scripts that parse fluxid configuration
- [ ] Dry-run works with all configuration sources (config files, env vars, CLI flags)
- [ ] Dry-run validates command file references and reports missing files
- [ ] Dry-run respects phase toggles (e.g., `--fluxid-no-commit`)
- [ ] Output format applies to initialization status only (agent output still streams normally)
- [ ] Complete UI: clear dry-run simulation output, structured JSON/YAML formats
- [ ] Full backend: simulation mode logic, synthetic report generation, structured serialization
- [ ] Can be deployed independently: works with all previous milestones
- [ ] Requires no additional milestones: full dry-run and output format capability

## Validation Questions
**Before marking this milestone complete, answer:**
- [ ] Can a real user perform complete workflow testing with only this milestone? **YES** - they can validate configuration and test loop logic without running agents
- [ ] Is it polished enough to ship publicly? **YES** - clear simulation output, useful for CI/CD integration
- [ ] Does it solve a real problem end-to-end? **YES** - enables safe testing of configuration changes
- [ ] Does it include both complete UI and functional backend integration? **YES** - dry-run output (UI) + simulation logic (backend)
- [ ] Can it run independently without waiting for other milestones? **YES** - builds on m01-m05, adds testing conveniences
- [ ] Would you personally use this if it were released today? **YES** - essential for validating configuration before committing

## Vertical Slice - All Layers Included
This milestone includes:
- **Dry-Run Mode**: Simulation execution path, synthetic report generation
- **Execution Plan Output**: Phase simulation printing, command file display
- **Configuration Validation**: Full validation without execution (file checks, schema validation)
- **Loop Simulation**: Synthetic loop progression with PASS reports
- **Output Format Control**: JSON/YAML serialization of initialization status
- **Structured Data Models**: Configuration state representation for serialization
- **Format Selection**: Flag parsing and format routing logic

## Notes
- Dry-run mode is **essential for testing** - validate configuration changes before running expensive agent workflows
- Output format control enables **automation integration** - external systems can parse fluxid status programmatically
- Synthetic PASS reports allow testing loop progression without real agent execution
- This completes all **core capabilities** from the product analysis
- Dry-run validates configuration correctness but doesn't test actual agent behavior
- Future work might include additional conveniences, monitoring, or advanced IPC features
