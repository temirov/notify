(function applyTauthConfig() {
  const config = window.PINGUIN_TAUTH_CONFIG || {};
  if (!window.__PINGUIN_CONFIG__) {
    window.__PINGUIN_CONFIG__ = {};
  }
  if (config.baseUrl) {
    window.__PINGUIN_CONFIG__.tauthBaseUrl = config.baseUrl;
  }
  if (config.googleClientId) {
    window.__PINGUIN_CONFIG__.googleClientId = config.googleClientId;
  }
  function applyAttributes() {
    const headers = document.querySelectorAll('mpr-header');
    headers.forEach((header) => {
      if (config.googleClientId) {
        header.setAttribute('site-id', config.googleClientId);
      }
      if (config.baseUrl) {
        header.setAttribute('base-url', config.baseUrl);
      }
      if (!header.getAttribute('login-path')) {
        header.setAttribute('login-path', '/auth/google');
      }
      if (!header.getAttribute('logout-path')) {
        header.setAttribute('logout-path', '/auth/logout');
      }
      if (!header.getAttribute('nonce-path')) {
        header.setAttribute('nonce-path', '/auth/nonce');
      }
    });
    const loginButtons = document.querySelectorAll('mpr-login-button');
    loginButtons.forEach((button) => {
      if (config.googleClientId) {
        button.setAttribute('site-id', config.googleClientId);
      }
      if (config.baseUrl) {
        button.setAttribute('base-url', config.baseUrl);
      }
      button.setAttribute('login-path', '/auth/google');
      button.setAttribute('logout-path', '/auth/logout');
      button.setAttribute('nonce-path', '/auth/nonce');
    });
  }
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', applyAttributes);
  } else {
    applyAttributes();
  }
})();
