// @ts-check
import Alpine from 'https://cdn.jsdelivr.net/npm/alpinejs@3.13.5/dist/module.esm.js';
import { RUNTIME_CONFIG, STRINGS } from './constants.js';
import { createApiClient } from './core/apiClient.js';
import { createNotificationsTable } from './ui/notificationsTable.js';
import { dispatchRefresh } from './core/events.js';
import { createToastCenter } from './ui/toastCenter.js';

window.Alpine = Alpine;

const apiClient = createApiClient(RUNTIME_CONFIG.apiBaseUrl);
const authController = createAuthController(RUNTIME_CONFIG);

Alpine.store('auth', createAuthStore());

document.addEventListener('alpine:init', () => {
  Alpine.data('landingAuthPanel', () => createLandingAuthPanel(authController));
  Alpine.data('dashboardShell', () => createDashboardShell(authController));
  Alpine.data('notificationsTable', () =>
    createNotificationsTable({
      apiClient,
      strings: STRINGS.dashboard,
      actions: STRINGS.actions,
    }),
  );
  Alpine.data('toastCenter', () => createToastCenter());
});

Alpine.start();

document.addEventListener('DOMContentLoaded', () => {
  bootstrapPage(authController);
  ensureMprUiLoaded();
});

function createAuthStore() {
  return {
    profile: null,
    isAuthenticated: false,
    setProfile(profile) {
      this.profile = profile;
      this.isAuthenticated = Boolean(profile);
    },
    clear() {
      this.profile = null;
      this.isAuthenticated = false;
    },
  };
}

function createLandingAuthPanel(controller) {
  return {
    STRINGS,
    notice: STRINGS.auth.ready,
    isBusy: false,
    stopStatusWatcher: null,
    init() {
      this.notice = STRINGS.auth.ready;
      this.stopStatusWatcher = controller.onStatusChange((status) => {
        switch (status) {
          case 'hydrating':
            this.isBusy = true;
            this.notice = STRINGS.auth.signingIn;
            break;
          case 'ready':
            this.isBusy = false;
            if (!window.Alpine.store('auth').isAuthenticated) {
              this.notice = STRINGS.auth.ready;
            } else {
              this.notice = '';
            }
            break;
          case 'error':
            this.isBusy = false;
            this.notice = STRINGS.auth.failed;
            break;
          default:
            break;
        }
      });
    },
    async handleSignInClick() {
      if (this.isBusy) {
        return;
      }
      this.isBusy = true;
      this.notice = STRINGS.auth.signingIn;
      try {
        await controller.prepareGoogleButton(this.$refs.googleButton);
        this.notice = '';
      } catch (error) {
        console.error('prepare sign-in failed', error);
        this.notice = STRINGS.auth.failed;
      } finally {
        this.isBusy = false;
      }
    },
    $cleanup() {
      if (typeof this.stopStatusWatcher === 'function') {
        this.stopStatusWatcher();
      }
    },
  };
}

function createDashboardShell(controller) {
  return {
    strings: STRINGS.dashboard,
    actions: STRINGS.actions,
    stopAuthWatcher: null,
    stopStatusWatcher: null,
    hasHydrated: false,
    hasRedirected: false,
    previousAuthState: false,
    init() {
      const authStore = window.Alpine.store('auth');
      this.previousAuthState = authStore.isAuthenticated;
      this.hasHydrated = false;
      this.hasRedirected = false;
      this.stopAuthWatcher = this.$watch(
        () => authStore.isAuthenticated,
        (isAuthenticated) => {
          const shouldRedirect =
            !isAuthenticated && (this.previousAuthState || this.hasHydrated);
          this.previousAuthState = isAuthenticated;
          if (shouldRedirect) {
            this.redirectToLanding();
          }
        },
      );
      this.stopStatusWatcher = controller.onStatusChange((status) => {
        if (status === 'ready' || status === 'error') {
          this.hasHydrated = true;
          if (!authStore.isAuthenticated) {
            this.redirectToLanding();
          }
        }
      });
    },
    refreshNotifications() {
      dispatchRefresh();
    },
    async handleLogout() {
      await controller.logout();
      this.redirectToLanding();
    },
    redirectToLanding() {
      if (this.hasRedirected) {
        return;
      }
      this.hasRedirected = true;
      window.location.assign(RUNTIME_CONFIG.landingUrl);
    },
    $cleanup() {
      if (typeof this.stopAuthWatcher === 'function') {
        this.stopAuthWatcher();
      }
      if (typeof this.stopStatusWatcher === 'function') {
        this.stopStatusWatcher();
      }
    },
  };
}

function bootstrapPage(controller) {
  const pageId = document.body.dataset.page || 'landing';
  let redirected = false;
  controller
    .hydrate({
      onAuthenticated(profile) {
        const store = Alpine.store('auth');
        store.setProfile(profile);
        if (pageId === 'landing' && !redirected) {
          redirected = true;
          window.location.assign(RUNTIME_CONFIG.dashboardUrl);
        }
      },
      onUnauthenticated() {
        const store = Alpine.store('auth');
        store.clear();
        if (pageId === 'dashboard' && !redirected) {
          redirected = true;
          window.location.assign(RUNTIME_CONFIG.landingUrl);
        }
      },
    })
    .catch((error) => {
      console.error('auth bootstrap failed', error);
    });
}

function createAuthController(config) {
  let activeNonceToken = '';
  let lastCallbacks = { onAuthenticated: undefined, onUnauthenticated: undefined };
  const statusListeners = new Set();

  const applyProfile = (profile) => {
    const store = Alpine.store('auth');
    if (profile) {
      store.setProfile(profile);
    } else {
      store.clear();
    }
  };

  const invokeCallback = (name, payload) => {
    const callback = lastCallbacks[name];
    if (typeof callback === 'function') {
      callback(payload);
    }
  };

  const setStatus = (status) => {
    statusListeners.forEach((listener) => listener(status));
  };

  const sessionChannel =
    typeof BroadcastChannel !== 'undefined' ? new BroadcastChannel('auth') : null;
  if (sessionChannel) {
    sessionChannel.addEventListener('message', (event) => {
      if (event.data === 'logged_out') {
        applyProfile(null);
        invokeCallback('onUnauthenticated');
      }
      if (event.data === 'refreshed') {
        hydrate(lastCallbacks).catch((error) => {
          console.error('hydrate after refresh failed', error);
        });
      }
    });
  }

  async function hydrate(callbacks = {}) {
    lastCallbacks = callbacks;
    setStatus('hydrating');
    try {
      await waitFor(() => typeof window.initAuthClient === 'function');
      const result = await window.initAuthClient({
        baseUrl: config.tauthBaseUrl,
        onAuthenticated(profile) {
          applyProfile(profile);
          invokeCallback('onAuthenticated', profile);
        },
        onUnauthenticated() {
          applyProfile(null);
          invokeCallback('onUnauthenticated');
        },
      });
      setStatus('ready');
      return result;
    } catch (error) {
      setStatus('error');
      throw error;
    }
  }

  async function prepareGoogleButton(targetElement) {
    if (!targetElement) {
      throw new Error('Google button host missing');
    }
    await waitFor(() => window.google && window.google.accounts && window.google.accounts.id);
    const noncePayload = await tauthFetch(config, '/auth/nonce', { method: 'POST' });
    activeNonceToken = noncePayload?.nonce || '';
    if (!activeNonceToken) {
      throw new Error('nonce_unavailable');
    }
    window.google.accounts.id.initialize({
      client_id: config.googleClientId,
      nonce: activeNonceToken,
      ux_mode: 'popup',
      callback: (response) => {
        handleGoogleCredential(response).catch((error) => console.error('credential exchange failed', error));
      },
    });
    window.google.accounts.id.renderButton(targetElement, {
      theme: 'outline',
      size: 'large',
      width: 320,
      text: 'signin_with',
    });
    window.google.accounts.id.prompt();
  }

  async function handleGoogleCredential(response) {
    if (!response || !response.credential || !activeNonceToken) {
      throw new Error('missing_google_credential');
    }
    await tauthFetch(config, '/auth/google', {
      method: 'POST',
      body: JSON.stringify({
        google_id_token: response.credential,
        nonce_token: activeNonceToken,
      }),
    });
    activeNonceToken = '';
    await hydrate(lastCallbacks);
  }

  async function logout() {
    if (typeof window.logout === 'function') {
      await window.logout();
    }
    applyProfile(null);
  }

  function onStatusChange(listener) {
    statusListeners.add(listener);
    return () => statusListeners.delete(listener);
  }

  return { hydrate, prepareGoogleButton, logout, onStatusChange };
}

function waitFor(checkFn, timeout = 12000) {
  return new Promise((resolve, reject) => {
    const start = Date.now();
    const tick = () => {
      const result = checkFn();
      if (result) {
        resolve(result);
        return;
      }
      if (Date.now() - start > timeout) {
        reject(new Error('timeout'));
        return;
      }
      setTimeout(tick, 80);
    };
    tick();
  });
}

function tauthFetch(config, path, options = {}) {
  const url = new URL(path, config.tauthBaseUrl);
  return fetch(url.toString(), {
    method: options.method || 'GET',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'X-Requested-With': 'XMLHttpRequest',
      ...(options.headers || {}),
    },
    body: options.body,
  }).then(async (response) => {
    const payload = await response.json().catch(() => ({}));
    if (!response.ok) {
      const error = new Error(payload?.error || 'request_failed');
      throw error;
    }
    return payload;
  });
}

function ensureMprUiLoaded() {
  if (document.querySelector('script[data-mpr-ui="true"]')) {
    return;
  }
  const script = document.createElement('script');
  script.defer = true;
  script.src = 'https://cdn.jsdelivr.net/gh/MarcoPoloResearchLab/mpr-ui@latest/mpr-ui.js';
  script.dataset.mprUi = 'true';
  document.head.appendChild(script);
}
