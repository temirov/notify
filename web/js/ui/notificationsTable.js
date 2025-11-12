// @ts-check
import { STATUS_LABELS, STATUS_OPTIONS } from '../constants.js';

/** @typedef {import('../types.d.js').NotificationItem} NotificationItem */

const inputFormatter = {
  toControlValue(isoString) {
    if (!isoString) {
      return '';
    }
    const date = new Date(isoString);
    if (Number.isNaN(date.getTime())) {
      return '';
    }
    const pad = (value) => String(value).padStart(2, '0');
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(
      date.getHours(),
    )}:${pad(date.getMinutes())}`;
  },
  toIso(controlValue) {
    if (!controlValue) {
      return null;
    }
    const date = new Date(controlValue);
    if (Number.isNaN(date.getTime())) {
      return null;
    }
    return date.toISOString();
  },
};

/**
 * @param {{ apiClient: ReturnType<typeof import('../core/apiClient.js').createApiClient> }} options
 */
export function createNotificationsTable(options) {
  const { apiClient } = options;
  const authStore = () => window.Alpine.store('auth');

  return {
    strings: options.strings,
    notifications: /** @type {NotificationItem[]} */ ([]),
    statusFilter: 'all',
    isLoading: false,
    errorMessage: '',
    scheduleDialogVisible: false,
    scheduleForm: {
      id: '',
      scheduledTime: '',
    },
    STATUS_OPTIONS,
    init() {
      this.refreshIfAuthenticated();
      this.$watch(
        () => authStore().isAuthenticated,
        (isAuthenticated) => {
          if (isAuthenticated) {
            this.loadNotifications();
          } else {
            this.notifications = [];
          }
        },
      );
      document.addEventListener('notifications:refresh', () => {
        if (authStore().isAuthenticated) {
          this.loadNotifications();
        }
      });
    },
    async loadNotifications() {
      if (!authStore().isAuthenticated) {
        return;
      }
      this.isLoading = true;
      this.errorMessage = '';
      try {
        const statuses = this.statusFilter === 'all' ? [] : [this.statusFilter];
        this.notifications = await apiClient.listNotifications(statuses);
      } catch (error) {
        this.errorMessage = (error && error.message) || 'Unable to load notifications.';
      } finally {
        this.isLoading = false;
      }
    },
    async refreshIfAuthenticated() {
      if (authStore().isAuthenticated) {
        await this.loadNotifications();
      }
    },
    formatStatus(status) {
      return STATUS_LABELS[status] || status;
    },
    formatTimestamp(isoString) {
      if (!isoString) {
        return '—';
      }
      const date = new Date(isoString);
      if (Number.isNaN(date.getTime())) {
        return '—';
      }
      return date.toLocaleString();
    },
    openScheduleDialog(notification) {
      this.scheduleForm.id = notification.id;
      this.scheduleForm.scheduledTime = inputFormatter.toControlValue(notification.scheduledFor);
      this.scheduleDialogVisible = true;
      const dialog = this.$refs.scheduleDialog;
      if (dialog && typeof dialog.showModal === 'function') {
        dialog.showModal();
      }
    },
    closeScheduleDialog() {
      this.scheduleDialogVisible = false;
      const dialog = this.$refs.scheduleDialog;
      if (dialog && typeof dialog.close === 'function') {
        dialog.close();
      }
    },
    async submitSchedule(event) {
      event?.preventDefault();
      const isoValue = inputFormatter.toIso(this.scheduleForm.scheduledTime);
      if (!isoValue) {
        this.errorMessage = 'Please provide a valid date/time in the future.';
        return;
      }
      try {
        await apiClient.rescheduleNotification(this.scheduleForm.id, isoValue);
        await this.loadNotifications();
        this.closeScheduleDialog();
      } catch (error) {
        this.errorMessage = (error && error.message) || 'Unable to reschedule notification.';
      }
    },
    async cancelNotification(notificationId) {
      if (!authStore().isAuthenticated) {
        return;
      }
      this.isLoading = true;
      try {
        await apiClient.cancelNotification(notificationId);
        await this.loadNotifications();
      } catch (error) {
        this.errorMessage = (error && error.message) || 'Unable to cancel notification.';
      } finally {
        this.isLoading = false;
      }
    },
  };
}
