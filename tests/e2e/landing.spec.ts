import { expect, test } from '@playwright/test';
import { configureRuntime, expectRenderedGoogleSignInMarkup, resetNotifications, stubExternalAssets } from './utils';

test.describe('Landing page authentication markup', () => {
  test.beforeEach(async ({ page, request }) => {
    await resetNotifications(request);
    await stubExternalAssets(page);
    await configureRuntime(page, { authenticated: false });
  });

  test('renders hero CTA', async ({ page }) => {
    await page.goto('/index.html');
    await expect(page.getByTestId('landing-cta')).toBeVisible();
  });

  test('HTML includes an explicit Google sign-in markup snippet', async ({ page }) => {
    await page.goto('/index.html');
    await expectRenderedGoogleSignInMarkup(page, 'landing page');
  });
});
