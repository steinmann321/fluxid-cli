import { test, expect } from '@playwright/test';
import { spawn, ChildProcess } from 'child_process';
import * as path from 'path';
import * as fs from 'fs';

// Helper to strip ANSI color codes
function stripAnsi(str: string): string {
  return str.replace(/\u001b\[[0-9;]*m/g, '');
}

test.describe('m01-e01: User runs workflow to completion', () => {
  const projectRoot = path.resolve(__dirname, '../..');
  const fluxidCli = path.join(projectRoot, 'fluxid-cli');
  const epicId = 'm01-e01-user-runs-workflow-to-completion.md';
  const progressFile = path.join(projectRoot, 'fluxid/progress.yaml');

  // Clean up progress state before each test
  test.beforeEach(async () => {
    if (fs.existsSync(progressFile)) {
      const content = fs.readFileSync(progressFile, 'utf-8');
      const cleanedContent = content.replace(/m01-e01:\s*\n\s*status:\s*done/g, 'm01-e01:\n  status: pending');
      fs.writeFileSync(progressFile, cleanedContent);
    }
  });

  test('should start automation with defaults and complete all loops', async () => {
    // Verify CLI exists
    expect(fs.existsSync(fluxidCli)).toBeTruthy();
    
    // Spawn fluxid CLI process in dry-run mode
    const child: ChildProcess = spawn(fluxidCli, ['--claude', '--dry-run', epicId], {
      cwd: projectRoot,
      env: { ...process.env },
    });
    
    let stdout = '';
    let stderr = '';
    let sessionId = '';
    
    // Collect output
    if (child.stdout) {
      child.stdout.on('data', (data) => {
        const output = data.toString();
        stdout += output;
        
        // Extract session ID from initialization output
        const cleanOutput = stripAnsi(output);
        const sessionIdMatch = cleanOutput.match(/Session ID:\s+([a-f0-9-]+)/);
        if (sessionIdMatch) {
          sessionId = sessionIdMatch[1];
        }
      });
    }
    
    if (child.stderr) {
      child.stderr.on('data', (data) => {
        stderr += data.toString();
      });
    }
    
    // Wait for process to complete (with timeout)
    const exitCode = await Promise.race([
      new Promise<number>((resolve) => {
        child.on('close', (code) => {
          resolve(code || 0);
        });
      }),
      new Promise<number>((_, reject) => {
        setTimeout(() => reject(new Error('Process timeout after 60s')), 60000);
      })
    ]);
    
    // Strip ANSI codes for assertion
    const cleanStdout = stripAnsi(stdout);
    const cleanStderr = stripAnsi(stderr);
    
    // Assertions
    console.log('STDOUT (clean):', cleanStdout);
    console.log('STDERR (clean):', cleanStderr);
    console.log('Exit Code:', exitCode);
    
    // 1. Verify initialization shows session ID (UUID v4 format)
    expect(cleanStdout).toContain('Session ID:');
    expect(sessionId).toMatch(/^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[a-f0-9]{4}-[a-f0-9]{12}$/);
    
    // 2. Verify initialization shows agent selection
    expect(cleanStdout).toMatch(/Agent:\s+claude/);
    
    // 3. Verify initialization shows loop counts
    expect(cleanStdout).toMatch(/Review Cycles:\s+20/);
    expect(cleanStdout).toMatch(/Implement Retries:\s+3/);
    
    // 4. Verify workflow started
    expect(cleanStdout).toContain('Starting workflow loop');
    
    // 5. Verify exit code 0 (success)
    expect(exitCode).toBe(0);
    
    // 6. Verify completion summary appears
    expect(cleanStdout).toMatch(/WORKFLOW COMPLETED SUCCESSFULLY|COMPLETED SUCCESSFULLY/);
  });

  test('should generate unique UUID v4 session ID across runs', async () => {
    // Run CLI twice and verify different session IDs
    const getSessionId = async (): Promise<string> => {
      const child = spawn(fluxidCli, ['--claude', '--dry-run', epicId], {
        cwd: projectRoot,
      });
      
      let stdout = '';
      if (child.stdout) {
        child.stdout.on('data', (data) => {
          stdout += data.toString();
        });
      }
      
      await Promise.race([
        new Promise<void>((resolve) => {
          child.on('close', () => resolve());
        }),
        new Promise<void>((_, reject) => {
          setTimeout(() => reject(new Error('Process timeout')), 60000);
        })
      ]);
      
      const cleanStdout = stripAnsi(stdout);
      const match = cleanStdout.match(/Session ID:\s+([a-f0-9-]+)/);
      return match ? match[1] : '';
    };
    
    const sessionId1 = await getSessionId();
    const sessionId2 = await getSessionId();
    
    console.log('Session ID 1:', sessionId1);
    console.log('Session ID 2:', sessionId2);
    
    // Both should be valid UUIDs
    expect(sessionId1).toMatch(/^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[a-f0-9]{4}-[a-f0-9]{12}$/);
    expect(sessionId2).toMatch(/^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[a-f0-9]{4}-[a-f0-9]{12}$/);
    
    // They should be different
    expect(sessionId1).not.toBe(sessionId2);
  });

  test('should propagate FLUXID_SESSION_ID to child processes', async () => {
    // Create a test helper script to verify environment variable
    const testHelperDir = path.join(projectRoot, 'fluxid/tmp');
    if (!fs.existsSync(testHelperDir)) {
      fs.mkdirSync(testHelperDir, { recursive: true });
    }
    
    const testHelper = path.join(testHelperDir, 'test-env-helper.sh');
    const testHelperContent = `#!/usr/bin/env bash
# Test helper to verify FLUXID_SESSION_ID is set
if [[ -n "\${FLUXID_SESSION_ID:-}" ]]; then
  echo "FLUXID_SESSION_ID_FOUND:\${FLUXID_SESSION_ID}" >&2
fi
`;
    fs.writeFileSync(testHelper, testHelperContent);
    fs.chmodSync(testHelper, '0755');
    
    // Patch epic-loop.sh to call our helper
    const epicLoopScript = path.join(projectRoot, '.fluxid/scripts/loop/epic-loop.sh');
    const originalContent = fs.readFileSync(epicLoopScript, 'utf-8');
    const patchedContent = originalContent.replace(
      /DRY_RUN=false/,
      `DRY_RUN=false\n${testHelper}`
    );
    fs.writeFileSync(epicLoopScript, patchedContent);
    
    try {
      // Run CLI
      const child = spawn(fluxidCli, ['--claude', '--dry-run', epicId], {
        cwd: projectRoot,
        env: { ...process.env },
      });
      
      let stdout = '';
      let stderr = '';
      let sessionId = '';
      
      if (child.stdout) {
        child.stdout.on('data', (data) => {
          const output = data.toString();
          stdout += output;
          const cleanOutput = stripAnsi(output);
          const match = cleanOutput.match(/Session ID:\s+([a-f0-9-]+)/);
          if (match) sessionId = match[1];
        });
      }
      
      if (child.stderr) {
        child.stderr.on('data', (data) => {
          stderr += data.toString();
        });
      }
      
      await Promise.race([
        new Promise<void>((resolve) => {
          child.on('close', () => resolve());
        }),
        new Promise<void>((_, reject) => {
          setTimeout(() => reject(new Error('Process timeout')), 60000);
        })
      ]);
      
      console.log('STDERR:', stderr);
      console.log('Session ID:', sessionId);
      
      // Verify the environment variable was propagated
      expect(stderr).toContain('FLUXID_SESSION_ID_FOUND:');
      expect(stderr).toContain(sessionId);
    } finally {
      // Restore original script
      fs.writeFileSync(epicLoopScript, originalContent);
      fs.unlinkSync(testHelper);
    }
  });
});
