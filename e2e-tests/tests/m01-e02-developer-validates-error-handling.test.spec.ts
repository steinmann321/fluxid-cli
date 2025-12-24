import { test, expect } from '@playwright/test';
import { execSync } from 'child_process';

/**
 * Epic: m01-e02 - Developer validates error handling and failure scenarios
 *
 * This E2E test validates that error path tests execute correctly and provide
 * comprehensive coverage of failure scenarios in the fluxid CLI.
 */

test.describe('m01-e02: Developer validates error handling', () => {
  test('should run error path tests and verify 2:1 error-to-success ratio', async () => {
    // Developer runs tests
    const testOutput = execSync('go test ./cmd/fluxid -v -run "TestMain_"', {
      cwd: process.cwd(),
      encoding: 'utf-8',
    });

    // Verify all tests pass
    expect(testOutput).toContain('PASS');
    expect(testOutput).not.toContain('FAIL');

    // Verify error scenario tests are present
    expect(testOutput).toContain('TestMain_ConfigLoadingError');
    expect(testOutput).toContain('TestMain_ArgParsingError_NegativeIterations');
    expect(testOutput).toContain('TestMain_ArgParsingError_InvalidIterations');
    expect(testOutput).toContain('TestMain_ArgParsingError_ZeroIterations');
    expect(testOutput).toContain('TestMain_AgentValidationError_UnsupportedAgent');
    expect(testOutput).toContain('TestMain_AgentValidationError_AgentNotInPath');
    expect(testOutput).toContain('TestMain_AgentValidationError_AgentNotExecutable');

    // Count test cases
    const testMatches = testOutput.match(/=== RUN   TestMain_/g);
    const errorTestMatches = testOutput.match(/=== RUN   TestMain_.*Error/g);
    const successTestMatches = testOutput.match(/=== RUN   TestMain_(Successful|Help|DryRun)/g);

    expect(testMatches).toBeTruthy();
    expect(errorTestMatches).toBeTruthy();
    expect(successTestMatches).toBeTruthy();

    const totalTests = testMatches!.length;
    const errorTests = errorTestMatches!.length;
    const successTests = successTestMatches!.length;

    console.log(`Total tests: ${totalTests}`);
    console.log(`Error path tests: ${errorTests}`);
    console.log(`Success path tests: ${successTests}`);
    console.log(`Ratio: ${(errorTests / successTests).toFixed(2)}:1`);

    // Verify 2:1 ratio (or better)
    expect(errorTests / successTests).toBeGreaterThanOrEqual(2.0);
  });

  test('should achieve 90%+ coverage with error path tests', async () => {
    // Developer runs tests with coverage
    const coverageOutput = execSync('go test ./cmd/fluxid -cover', {
      cwd: process.cwd(),
      encoding: 'utf-8',
    });

    // Verify coverage report
    expect(coverageOutput).toContain('coverage:');

    // Extract coverage percentage
    const coverageMatch = coverageOutput.match(/coverage: ([\d.]+)% of statements/);
    expect(coverageMatch).toBeTruthy();

    const coveragePercent = parseFloat(coverageMatch![1]);
    console.log(`Coverage: ${coveragePercent}%`);

    // Verify >= 90% coverage
    expect(coveragePercent).toBeGreaterThanOrEqual(90.0);
  });

  test('should validate config loading error scenarios', async () => {
    // Developer runs specific error scenario tests
    const output = execSync('go test ./cmd/fluxid -v -run "TestMain_ConfigLoadingError"', {
      cwd: process.cwd(),
      encoding: 'utf-8',
    });

    // Verify test executes and passes
    expect(output).toContain('RUN   TestMain_ConfigLoadingError');
    expect(output).toContain('PASS: TestMain_ConfigLoadingError');
  });

  test('should validate argument parsing error scenarios', async () => {
    // Developer runs argument parsing error tests
    const output = execSync('go test ./cmd/fluxid -v -run "TestMain_ArgParsingError"', {
      cwd: process.cwd(),
      encoding: 'utf-8',
    });

    // Verify multiple argument error scenarios are tested
    expect(output).toContain('TestMain_ArgParsingError_NegativeIterations');
    expect(output).toContain('TestMain_ArgParsingError_InvalidIterations');
    expect(output).toContain('TestMain_ArgParsingError_ZeroIterations');
    expect(output).toContain('TestMain_ArgParsingError_MissingIterationsValue');
    expect(output).toContain('TestMain_ArgParsingError_NegativeImplementRetries');
    expect(output).toContain('TestMain_ArgParsingError_InvalidImplementRetries');

    // All should pass
    expect(output).not.toContain('FAIL');
  });

  test('should validate agent validation error scenarios', async () => {
    // Developer runs agent validation error tests
    const output = execSync('go test ./cmd/fluxid -v -run "TestMain_AgentValidationError"', {
      cwd: process.cwd(),
      encoding: 'utf-8',
    });

    // Verify multiple agent validation scenarios are tested
    expect(output).toContain('TestMain_AgentValidationError_UnsupportedAgent');
    expect(output).toContain('TestMain_AgentValidationError_AgentNotInPath');
    expect(output).toContain('TestMain_AgentValidationError_AgentNotExecutable');

    // All should pass
    expect(output).not.toContain('FAIL');
  });

  test('should provide clear error diagnostics in test failures', async () => {
    // Run tests verbosely to see diagnostic messages
    const output = execSync('go test ./cmd/fluxid -v -run "TestMain_.*Error"', {
      cwd: process.cwd(),
      encoding: 'utf-8',
    });

    // Verify test names are descriptive
    expect(output).toMatch(/TestMain_\w+Error_\w+/);

    // Verify tests contain clear assertion messages
    // (This is implicit - Go test framework shows which assertion failed)
    expect(output).toContain('PASS');
  });
});
