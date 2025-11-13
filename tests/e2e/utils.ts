import { expect, Page } from '@playwright/test';

export async function configureRuntime(page: Page, options: { authenticated: boolean }) {
  const baseUrl = process.env.PLAYWRIGHT_BASE_URL || 'http://127.0.0.1:4173';
  await page.addInitScript(
    ({ base, authenticated }) => {
      window.__PINGUIN_CONFIG__ = {
        apiBaseUrl: '/api',
        tauthBaseUrl: base,
        googleClientId: 'playwright-client',
        landingUrl: '/index.html',
        dashboardUrl: '/dashboard.html',
      };
      window.__mockAuth = {
        authenticated,
        profile: {
          user_email: 'playwright@example.com',
          user_display_name: 'Playwright User',
          user_avatar_url: '',
        },
      };
    },
    { base: baseUrl, authenticated: options.authenticated },
  );
}

export async function stubExternalAssets(page: Page) {
  await page.route('https://accounts.google.com/gsi/client', (route) => {
    route.fulfill({
      contentType: 'text/javascript',
      body: 'window.google = { accounts: { id: { initialize() {}, renderButton(el) { el.innerHTML = "<button class=\'button secondary\'>Google</button>"; }, prompt() {} } } };',
    });
  });
  const emptyScript = { contentType: 'text/javascript', body: '' };
  await page.route('https://cdn.jsdelivr.net/gh/MarcoPoloResearchLab/mpr-ui@latest/mpr-ui.js', (route) =>
    route.fulfill(emptyScript),
  );
  await page.route('https://cdn.jsdelivr.net/gh/MarcoPoloResearchLab/mpr-ui@latest/mpr-ui.css', (route) =>
    route.fulfill({ contentType: 'text/css', body: '' }),
  );
}

export async function resetNotifications(request: import('@playwright/test').APIRequestContext, overrides = {}) {
  await request.post('/testing/reset', {
    data: overrides,
  });
}

export async function expectToast(page: Page, text: string) {
  await expect(page.getByRole('button', { name: text }).first()).toBeVisible();
}
