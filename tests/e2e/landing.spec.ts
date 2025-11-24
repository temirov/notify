import { expect, test } from '@playwright/test';
import { configureRuntime, expectHeroLoginButton, resetNotifications, stubExternalAssets } from './utils';

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
    await expectHeroLoginButton(page, 'Enter workspace');
    const signInSurface = page
      .getByTestId('landing-cta')
      .getByRole('button', { name: 'Enter workspace' })
      .first();
    await expect(signInSurface).toBeVisible();
  });

  test('completes Google/TAuth handshake and redirects to dashboard', async ({ page }) => {
    await page.goto('/index.html');
    await expectHeroLoginButton(page, 'Enter workspace');
    const heroButton = page
      .getByTestId('landing-cta')
      .getByRole('button', { name: 'Enter workspace' })
      .first();
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

  test('mpr-header attributes mirror runtime TAuth base URL', async ({ page }) => {
    await page.goto('/index.html');
    const runtimeBase = await page.evaluate(() => (window as any).__PINGUIN_CONFIG__?.tauthBaseUrl || '');
    const normalizedRuntimeBase = runtimeBase.replace(/\/$/, '');
    if (normalizedRuntimeBase) {
      await page.waitForFunction((expected) => {
        const header = document.querySelector('mpr-header');
        return header && header.getAttribute('base-url') === expected;
      }, normalizedRuntimeBase);
    }
    const headerBase = (await page.locator('mpr-header').first().getAttribute('base-url')) || '';
    if (normalizedRuntimeBase) {
      expect(headerBase).toBe(normalizedRuntimeBase);
    } else {
      expect(headerBase).not.toBe('');
    }
  });
});
