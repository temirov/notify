#!/usr/bin/env node
const http = require('http');
const fs = require('fs');
const path = require('path');

const HOST = '127.0.0.1';
const PORT = process.env.PLAYWRIGHT_PORT ? Number(process.env.PLAYWRIGHT_PORT) : 4173;
const WEB_ROOT = path.resolve(__dirname, '../../web');
const AUTH_CLIENT_PATH = path.resolve(__dirname, './stubs/auth-client.js');

let serverState = createDefaultState();

function createDefaultState() {
  return {
    notifications: defaultNotifications(),
    failList: false,
    failReschedule: false,
    failCancel: false,
  };
}

function defaultNotifications() {
  const now = new Date();
  return [
    {
      notification_id: 'notif-1',
      notification_type: 'email',
      recipient: 'user@example.com',
      subject: 'Queued notification',
      message: 'Hello from tests',
      status: 'queued',
      created_at: now.toISOString(),
      updated_at: now.toISOString(),
      scheduled_time: new Date(now.getTime() + 3600 * 1000).toISOString(),
      retry_count: 0,
    },
  ];
}

function applyOverrides(payload) {
  if (Array.isArray(payload.notifications) && payload.notifications.length > 0) {
    serverState.notifications = payload.notifications;
  } else {
    serverState.notifications = defaultNotifications();
  }
  serverState.failList = Boolean(payload.failList);
  serverState.failReschedule = Boolean(payload.failReschedule);
  serverState.failCancel = Boolean(payload.failCancel);
}

function readJson(req) {
  return new Promise((resolve) => {
    let data = '';
    req.on('data', (chunk) => {
      data += chunk;
    });
    req.on('end', () => {
      try {
        resolve(JSON.parse(data || '{}'));
      } catch (error) {
        resolve({});
      }
    });
  });
}

function sendJson(res, statusCode, body, headers = {}) {
  res.writeHead(statusCode, {
    'Content-Type': 'application/json',
    'Cache-Control': 'no-store',
    ...headers,
  });
  res.end(body ? JSON.stringify(body) : undefined);
}

function serveStatic(filePath, res) {
  fs.readFile(filePath, (error, data) => {
    if (error) {
      res.writeHead(404);
      res.end('Not found');
      return;
    }
    const ext = path.extname(filePath).toLowerCase();
    const contentType =
      {
        '.html': 'text/html; charset=utf-8',
        '.js': 'text/javascript; charset=utf-8',
        '.css': 'text/css; charset=utf-8',
        '.json': 'application/json',
      }[ext] || 'application/octet-stream';
    res.writeHead(200, { 'Content-Type': contentType });
    res.end(data);
  });
}

const server = http.createServer(async (req, res) => {
  const url = new URL(req.url, `http://${req.headers.host}`);
  if (req.method === 'POST' && url.pathname === '/testing/reset') {
    const body = await readJson(req);
    applyOverrides(body || {});
    sendJson(res, 204, null);
    return;
  }

  if (req.method === 'GET' && url.pathname === '/api/notifications') {
    if (serverState.failList) {
      sendJson(res, 500, { error: 'list_failed' });
      return;
    }
    const statuses = url.searchParams.getAll('status').filter(Boolean);
    const filtered = filterNotifications(serverState.notifications, statuses);
    sendJson(res, 200, { notifications: filtered });
    return;
  }

  const scheduleMatch = url.pathname.match(/^\/api\/notifications\/([^/]+)\/schedule$/);
  if (scheduleMatch && req.method === 'PATCH') {
    if (serverState.failReschedule) {
      sendJson(res, 500, { error: 'reschedule_failed' });
      return;
    }
    const body = await readJson(req);
    const scheduled_time = body.scheduled_time || null;
    serverState.notifications = serverState.notifications.map((item) => {
      if (item.notification_id === scheduleMatch[1]) {
        return { ...item, scheduled_time, status: 'queued', updated_at: new Date().toISOString() };
      }
      return item;
    });
    const updated = serverState.notifications.find((item) => item.notification_id === scheduleMatch[1]);
    sendJson(res, 200, updated || {});
    return;
  }

  const cancelMatch = url.pathname.match(/^\/api\/notifications\/([^/]+)\/cancel$/);
  if (cancelMatch && req.method === 'POST') {
    if (serverState.failCancel) {
      sendJson(res, 500, { error: 'cancel_failed' });
      return;
    }
    serverState.notifications = serverState.notifications.map((item) => {
      if (item.notification_id === cancelMatch[1]) {
        return { ...item, status: 'cancelled', updated_at: new Date().toISOString() };
      }
      return item;
    });
    const updated = serverState.notifications.find((item) => item.notification_id === cancelMatch[1]);
    sendJson(res, 200, updated || {});
    return;
  }

  if (url.pathname === '/auth/nonce' && req.method === 'POST') {
    sendJson(res, 200, { nonce: Date.now().toString() });
    return;
  }

  if (url.pathname === '/auth/google' && req.method === 'POST') {
    sendJson(res, 200, { ok: true });
    return;
  }

  if (url.pathname === '/auth/refresh' && req.method === 'POST') {
    sendJson(res, 204, null);
    return;
  }

  if (url.pathname === '/static/auth-client.js') {
    serveStatic(AUTH_CLIENT_PATH, res);
    return;
  }

  // Serve static files from /web
  let relativePath = url.pathname;
  if (relativePath === '/' || relativePath === '') {
    relativePath = '/index.html';
  }
  const safePath = path.normalize(relativePath).replace(/^\.\/+/, '');
  const filePath = path.join(WEB_ROOT, safePath);
  serveStatic(filePath, res);
});

function filterNotifications(source, statuses) {
  if (!Array.isArray(source) || source.length === 0) {
    return [];
  }
  if (!statuses || statuses.length === 0) {
    return source;
  }
  const wanted = new Set(statuses.map((value) => String(value).toLowerCase()));
  return source.filter((item) => wanted.has(String(item.status).toLowerCase()));
}

server.listen(PORT, HOST, () => {
  console.log(`Playwright test server listening on http://${HOST}:${PORT}`);
});
