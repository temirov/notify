import { expect, test } from '@playwright/test';
import { configureRuntime, expectHeaderGoogleButton, resetNotifications, stubExternalAssets } from './utils';

test.describe('Landing page auth flow', () => {
  test.beforeEach(async ({ page, request }) => {
    page.on('console', (message) => {
      console.log('[landing]', message.type(), message.text());
    });
    await resetNotifications(request);
    await stubExternalAssets(page);
    await configureRuntime(page, { authenticated: false });
  });

  test('shows CTA and disables button during GIS prep', async ({ page }) => {
    await page.goto('/index.html');
    await expectHeaderGoogleButton(page);
    const signInSurface = page.getByRole('button', { name: 'Continue to dashboard' });
    await expect(signInSurface).toBeVisible();
  });

  test('completes Google/TAuth handshake and redirects to dashboard', async ({ page }) => {
    await page.goto('/index.html');
    await expectHeaderGoogleButton(page);
    const heroButton = page.getByRole('button', { name: 'Continue to dashboard' });
    await heroButton.click();
    await expect(heroButton).toBeVisible();
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
