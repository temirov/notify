import { expect, test } from '@playwright/test';
import { configureRuntime, resetNotifications, stubExternalAssets, expectToast } from './utils';

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page, request }) => {
    await resetNotifications(request);
    await stubExternalAssets(page);
    await configureRuntime(page, { authenticated: true });
  });

  test('renders notification table and allows cancel', async ({ page }) => {
    await page.goto('/dashboard.html');
    await expect(page.getByTestId('notification-row')).toHaveCount(1);
    page.once('dialog', (dialog) => dialog.accept());
    await page.getByRole('button', { name: 'Cancel' }).click();
    await expectToast(page, 'Notification cancelled');
  });

  test('reschedule flow updates toast', async ({ page }) => {
    await page.goto('/dashboard.html');
    await page.getByRole('button', { name: 'Reschedule' }).click();
    const input = page.getByLabel('Delivery time');
    const newDate = new Date(Date.now() + 7200 * 1000).toISOString().slice(0, 16);
    await input.fill(newDate);
    await page.getByRole('button', { name: 'Save changes' }).click();
    await expectToast(page, 'Delivery time updated');
  });
});
