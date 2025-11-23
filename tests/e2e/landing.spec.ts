import { expect, test } from '@playwright/test';
import { configureRuntime, resetNotifications, stubExternalAssets } from './utils';

test.describe('Landing page auth flow', () => {
  test.beforeEach(async ({ page, request }) => {
    await resetNotifications(request);
    await stubExternalAssets(page);
    await configureRuntime(page, { authenticated: false });
  });

  test('shows CTA and disables button during GIS prep', async ({ page }) => {
    await page.goto('/index.html');
    const signInButton = page.getByTestId('landing-sign-in');
    await expect(signInButton).toBeVisible();
    await signInButton.click();
    await expect(page.locator('[data-testid="google-button-host"] button')).toBeVisible();
  });

  test('completes Google/TAuth handshake and redirects to dashboard', async ({ page }) => {
    await page.goto('/index.html');
    const noncePromise = page.waitForRequest(/\/auth\/nonce$/);
    await page.getByTestId('landing-sign-in').click();
    await expect(page.locator('[data-testid="google-button-host"] button')).toBeVisible();
    await noncePromise;
    const googleExchange = page.waitForRequest(/\/auth\/google$/);
    const navigation = page.waitForNavigation({ url: '**/dashboard.html' });
    await page.evaluate(() => {
      const googleStub = (window as any).__playwrightGoogle;
      googleStub?.trigger({ credential: 'playwright-token' });
    });
    await googleExchange;
    await navigation;
    await expect(page.getByTestId('notifications-table')).toBeVisible();
  });
});
