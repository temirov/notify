import { expect, Page } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';

const projectRoot = path.resolve(__dirname, '..', '..');
const mprUiScript = fs.readFileSync(path.join(projectRoot, 'tools/mpr-ui/mpr-ui.js'), 'utf-8');
const mprUiStyles = fs.readFileSync(path.join(projectRoot, 'tools/mpr-ui/mpr-ui.css'), 'utf-8');
const authClientStub = fs.readFileSync(
  path.join(projectRoot, 'tests/support/stubs/auth-client.js'),
  'utf-8',
);

export async function configureRuntime(page: Page, options: { authenticated: boolean }) {
  const baseUrl = process.env.PLAYWRIGHT_BASE_URL || 'http://127.0.0.1:4173';
  await page.addInitScript(
    ({ authenticated }) => {
      if (!window.name) {
        const defaultProfile = {
          user_email: 'playwright@example.com',
          user_display_name: 'Playwright User',
          user_avatar_url: '',
        };
        window.name = JSON.stringify({
          __mockAuth: {
            authenticated,
            profile: defaultProfile,
          },
        });
      }
    },
    { authenticated: options.authenticated },
  );
  await page.addInitScript(
    ({ base, authenticated }) => {
      window.__PINGUIN_CONFIG__ = {
        apiBaseUrl: '/api',
        tauthBaseUrl: base,
        landingUrl: '/index.html',
        dashboardUrl: '/dashboard.html',
        runtimeConfigUrl: '/runtime-config',
        skipRemoteConfig: true,
      };
      window.__PINGUIN_RUNTIME_CONFIG_URL = '/runtime-config';
      const defaultProfile = {
        user_email: 'playwright@example.com',
        user_display_name: 'Playwright User',
        user_avatar_url: '',
      };
      const storedState = (() => {
        try {
          return window.name ? JSON.parse(window.name) : null;
        } catch {
          return null;
        }
      })();
      const session = storedState?.__mockAuth || {
        authenticated,
        profile: defaultProfile,
      };
      session.profile = session.profile || defaultProfile;
      window.__mockAuth = session;
      window.__persistMockAuth = () => {
        const payload = storedState || {};
        payload.__mockAuth = window.__mockAuth;
        try {
          window.name = JSON.stringify(payload);
        } catch {
          // ignore
        }
      };
      window.__persistMockAuth();
    },
    { base: baseUrl, authenticated: options.authenticated },
  );
  await page.addInitScript(({ base }) => {
    window.PINGUIN_TAUTH_CONFIG = {
      baseUrl: base,
      googleClientId: '991677581607-r0dj8q6irjagipali0jpca7nfp8sfj9r.apps.googleusercontent.com',
    };
  }, { base: baseUrl });
}

export async function stubExternalAssets(page: Page) {
  await page.route('https://accounts.google.com/gsi/client', (route) => {
    const googleStub = `
      window.__playwrightGoogle = {
        callback: null,
        trigger(payload) {
          if (!this.callback) {
            return;
          }
          window.__mockAuth = window.__mockAuth || { authenticated: false };
          window.__mockAuth.authenticated = true;
          window.__mockAuth.profile =
            window.__mockAuth.profile || {
              user_email: 'playwright@example.com',
              user_display_name: 'Playwright User',
              user_avatar_url: '',
            };
          window.__persistMockAuth && window.__persistMockAuth();
          this.callback(payload || { credential: 'playwright-token' });
        },
      };
      window.google = {
        accounts: {
          id: {
            initialize(config) {
              window.__playwrightGoogle.callback = config && config.callback;
            },
            renderButton(el, options) {
              var label = (options && options.text) || "Sign in";
              el.innerHTML = "<button class='button secondary'>" + label + "</button>";
            },
            prompt() {},
          },
        },
      };
    `;
    route.fulfill({
      contentType: 'text/javascript',
      body: googleStub,
    });
  });
  await page.route('https://cdn.jsdelivr.net/gh/MarcoPoloResearchLab/mpr-ui@latest/mpr-ui.js', (route) =>
    route.fulfill({ contentType: 'text/javascript', body: mprUiScript }),
  );
  await page.route('https://cdn.jsdelivr.net/gh/MarcoPoloResearchLab/mpr-ui@latest/mpr-ui.css', (route) =>
    route.fulfill({ contentType: 'text/css', body: mprUiStyles }),
  );
  await page.route('**/static/auth-client.js', (route) =>
    route.fulfill({ contentType: 'text/javascript', body: authClientStub }),
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

export async function expectRenderedGoogleSignInMarkup(page: Page, context = 'page') {
  const html = await page.content();
  const normalized = html.toLowerCase();
  expect(
    normalized.includes('google sign in') || normalized.includes('<mpr-login-button'),
    `expected ${context} HTML to contain Google sign-in markup`,
  ).toBe(true);
}
