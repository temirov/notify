import { expect, test } from '@playwright/test';
import { configureRuntime, stubExternalAssets } from './utils';

test.describe('Landing page auth flow', () => {
  test.beforeEach(async ({ page }) => {
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
});
