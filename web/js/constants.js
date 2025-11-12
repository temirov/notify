// @ts-check

/** @type {Window & typeof globalThis & { __PINGUIN_CONFIG__?: Record<string, unknown> }} */
const runtimeWindow = window;
const rawConfig = runtimeWindow.__PINGUIN_CONFIG__ ?? {};

const normalizeUrl = (value, fallback) => {
  if (!value || typeof value !== "string") {
    return fallback;
  }
  return value.trim().replace(/\/$/, "") || fallback;
};

export const RUNTIME_CONFIG = Object.freeze({
  apiBaseUrl: normalizeUrl(rawConfig.apiBaseUrl, "/api"),
  tauthBaseUrl: normalizeUrl(rawConfig.tauthBaseUrl, "http://localhost:8081"),
  googleClientId: String(rawConfig.googleClientId || "YOUR_GOOGLE_WEB_CLIENT_ID"),
  landingUrl: String(rawConfig.landingUrl || "/index.html"),
  dashboardUrl: String(rawConfig.dashboardUrl || "/dashboard.html"),
});

export const STRINGS = Object.freeze({
  appName: "Pinguin Notification Service",
  landing: {
    eyebrow: "Trusted delivery infrastructure",
    headline: "Deliver email and SMS notifications with confidence",
    subheadline:
      "Authenticate with Google Identity, preview schedules, and manage queued notifications from a single workspace.",
    ctaPrimary: "Sign in with Google",
    ctaSecondary: "Explore platform",
    securityCopy: "Your session stays protected by HttpOnly cookies minted by TAuth.",
  },
  dashboard: {
    title: "Scheduled notifications",
    subtitle: "Review delivery status, adjust schedules, or cancel queued jobs in a single view.",
    emptyState: "No notifications yet. Start by sending one via the CLI or gRPC client.",
    scheduleDialogTitle: "Reschedule notification",
    scheduleDialogDescription: "Select a new delivery time. Notifications can only be edited while queued.",
  },
  auth: {
    signingIn: "Preparing Google Sign-In…",
    ready: "Sign in to continue",
    failed: "We could not reach the authentication service. Please retry.",
    loggedOut: "Session ended. Redirecting…",
  },
  actions: {
    refresh: "Refresh",
    reschedule: "Reschedule",
    cancel: "Cancel",
    saveChanges: "Save changes",
    close: "Close",
    logout: "Sign out",
  },
});

export const STATUS_LABELS = Object.freeze({
  queued: "Queued",
  sent: "Sent",
  errored: "Errored",
  cancelled: "Cancelled",
});

export const STATUS_OPTIONS = Object.freeze([
  { value: "all", label: "All statuses" },
  { value: "queued", label: STATUS_LABELS.queued },
  { value: "sent", label: STATUS_LABELS.sent },
  { value: "errored", label: STATUS_LABELS.errored },
  { value: "cancelled", label: STATUS_LABELS.cancelled },
]);
