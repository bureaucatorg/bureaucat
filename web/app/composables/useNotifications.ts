import type { NotificationListResponse } from "~/types";

// Module-level shared state so the navbar badge and the popover stay in sync.
const unreadCount = ref(0);

export function useNotifications() {
  const { getAuthHeader, isAuthenticated } = useAuth();

  async function listNotifications(
    page = 1,
    perPage = 20
  ): Promise<{ success: boolean; data?: NotificationListResponse }> {
    try {
      const params = new URLSearchParams({
        page: String(page),
        per_page: String(perPage),
      });
      const response = await fetch(`/api/v1/me/notifications?${params}`, {
        headers: getAuthHeader(),
      });
      if (!response.ok) return { success: false };
      const data: NotificationListResponse = await response.json();
      unreadCount.value = data.unread_count ?? 0;
      return { success: true, data };
    } catch {
      return { success: false };
    }
  }

  async function refreshUnreadCount(): Promise<void> {
    if (!isAuthenticated.value) {
      unreadCount.value = 0;
      return;
    }
    try {
      const response = await fetch("/api/v1/me/notifications/unread_count", {
        headers: getAuthHeader(),
      });
      if (response.ok) {
        const data = await response.json();
        unreadCount.value = data.count ?? 0;
      }
    } catch {
      // silently fail
    }
  }

  async function markRead(id: string): Promise<void> {
    try {
      const response = await fetch(`/api/v1/me/notifications/${id}/read`, {
        method: "POST",
        headers: getAuthHeader(),
      });
      if (response.ok && unreadCount.value > 0) {
        unreadCount.value -= 1;
      }
    } catch {
      // silently fail
    }
  }

  async function markAllRead(): Promise<void> {
    try {
      const response = await fetch("/api/v1/me/notifications/read_all", {
        method: "POST",
        headers: getAuthHeader(),
      });
      if (response.ok) {
        unreadCount.value = 0;
      }
    } catch {
      // silently fail
    }
  }

  async function fetchNotificationSettings(): Promise<{
    success: boolean;
    data?: { email_enabled: boolean; email_available: boolean };
  }> {
    try {
      const response = await fetch("/api/v1/me/notification_settings", {
        headers: getAuthHeader(),
      });
      if (!response.ok) return { success: false };
      return { success: true, data: await response.json() };
    } catch {
      return { success: false };
    }
  }

  async function updateEmailNotifications(enabled: boolean): Promise<{ success: boolean }> {
    try {
      const response = await fetch("/api/v1/me/notification_settings", {
        method: "PUT",
        headers: { "Content-Type": "application/json", ...getAuthHeader() },
        body: JSON.stringify({ email_enabled: enabled }),
      });
      return { success: response.ok };
    } catch {
      return { success: false };
    }
  }

  return {
    unreadCount,
    fetchNotificationSettings,
    updateEmailNotifications,
    listNotifications,
    refreshUnreadCount,
    markRead,
    markAllRead,
  };
}
