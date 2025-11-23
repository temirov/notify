(function applyTauthConfig() {
  const config = window.PINGUIN_TAUTH_CONFIG || {};
  function applyAttributes() {
    const headers = document.querySelectorAll('mpr-header');
    if (!headers.length) {
      return;
    }
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
  }
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', applyAttributes);
  } else {
    applyAttributes();
  }
})();
