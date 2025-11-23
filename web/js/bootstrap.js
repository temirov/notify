// @ts-check

const DEFAULT_CONFIG = Object.freeze({
  tauthBaseUrl: 'http://localhost:8081',
  landingUrl: '/index.html',
  dashboardUrl: '/dashboard.html',
});
const RUNTIME_CONFIG_URL_HINT =
  typeof window.__PINGUIN_RUNTIME_CONFIG_URL === 'string'
    ? window.__PINGUIN_RUNTIME_CONFIG_URL.trim()
    : '';

function deriveApiOriginFromConfig(config) {
  const apiBase = config && typeof config.apiBaseUrl === 'string' ? config.apiBaseUrl : '';
  if (apiBase.startsWith('http://') || apiBase.startsWith('https://')) {
    try {
      return new URL(apiBase).origin;
    } catch {
      // ignore invalid URL
    }
  }
  if (apiBase.startsWith('/')) {
    return '';
  }
  const { protocol, hostname, port } = window.location;
  if (port === '4173') {
    return `${protocol}//${hostname}:8080`;
  }
  if (port && port.length > 0) {
    return `${protocol}//${hostname}:${port}`;
  }
  return `${protocol}//${hostname}`;
}

function resolveRuntimeConfigCandidates(config, hint) {
  const candidates = [];
  if (hint) {
    candidates.push(hint);
  }
  const candidate =
    config && typeof config.runtimeConfigUrl === 'string'
      ? config.runtimeConfigUrl.trim()
      : '';
  if (candidate && !candidates.includes(candidate)) {
    candidates.push(candidate);
  }
  if (!candidates.length) {
    candidates.push('/runtime-config');
  }
  const origin = deriveApiOriginFromConfig(config);
  const apiCandidate = origin ? `${origin}/runtime-config` : null;
  if (apiCandidate && apiCandidate !== candidates[0]) {
    candidates.push(apiCandidate);
  }
  return candidates;
}

async function fetchRuntimeConfig(config, hint) {
  const candidates = resolveRuntimeConfigCandidates(config, hint);
  let lastError;
  for (const url of candidates) {
    try {
      console.info('runtime_config_candidate', url);
      const response = await fetch(url, { credentials: 'omit' });
      if (!response.ok) {
        lastError = new Error(`runtime_config_${response.status}`);
        continue;
      }
      return response.json();
    } catch (error) {
      lastError = error;
    }
  }
  throw lastError || new Error('runtime_config_failed');
}

function loadAuthClient(baseUrl) {
  const script = document.createElement('script');
  script.defer = true;
  const normalized = (baseUrl || '').replace(/\/$/, '') || DEFAULT_CONFIG.tauthBaseUrl;
  script.src = `${normalized}/static/auth-client.js`;
  document.head.appendChild(script);
}

function mergeConfig(base, overrides) {
  if (!overrides || typeof overrides !== 'object') {
    return { ...base };
  }
  return { ...base, ...overrides };
}

(async function bootstrap() {
  const preloaded = window.__PINGUIN_CONFIG__;
  const skipRemote = Boolean(preloaded && preloaded.skipRemoteConfig);
  let effectiveConfig = mergeConfig(DEFAULT_CONFIG, preloaded || {});
  if (!skipRemote) {
    try {
      const remote = await fetchRuntimeConfig(preloaded || null, RUNTIME_CONFIG_URL_HINT);
      effectiveConfig = mergeConfig(effectiveConfig, remote);
    } catch (error) {
      console.warn('runtime config fetch failed', error);
    }
  }
  const finalConfig = { ...effectiveConfig };
  delete finalConfig.skipRemoteConfig;
  delete finalConfig.runtimeConfigUrl;
  window.__PINGUIN_CONFIG__ = finalConfig;
  loadAuthClient(window.__PINGUIN_CONFIG__.tauthBaseUrl);
  await import('./app.js');
})();
