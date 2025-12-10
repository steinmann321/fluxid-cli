import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  timeout: 300000, // 5 minutes for long-running workflow tests
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: 'list',
  use: {
    trace: 'on-first-retry',
  },
});
