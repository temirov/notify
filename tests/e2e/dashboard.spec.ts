import { expect, test } from '@playwright/test';
import { configureRuntime, expectRenderedGoogleSignInMarkup, resetNotifications, stubExternalAssets } from './utils';

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page, request }) => {
    await resetNotifications(request);
    await stubExternalAssets(page);
  });

  test('renders Google sign-in markup when a guest visits the dashboard entry point', async ({ page }) => {
    await configureRuntime(page, { authenticated: false });
    await page.goto('/dashboard.html');
    await expectRenderedGoogleSignInMarkup(page, 'dashboard entry page');
  });

  test('renders Google sign-in markup for authenticated users as well', async ({ page }) => {
    await configureRuntime(page, { authenticated: true });
    await page.goto('/dashboard.html');
    await expectRenderedGoogleSignInMarkup(page, 'dashboard (authenticated)');
  });
});
