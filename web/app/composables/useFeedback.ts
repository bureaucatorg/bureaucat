export interface FeedbackItem {
  id: string;
  message: string;
  source_origin?: string;
  user_agent?: string;
  created_at: string;
}

export interface PaginatedFeedback {
  items: FeedbackItem[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export function useFeedback() {
  const { getAuthHeader } = useAuth();

  async function list(
    page = 1,
    perPage = 50
  ): Promise<{ success: boolean; data?: PaginatedFeedback; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/admin/feedback?page=${page}&per_page=${perPage}`,
        { headers: { ...getAuthHeader() }, credentials: "include" }
      );
      if (!response.ok) {
        const data = await response.json().catch(() => ({}));
        return { success: false, error: data.message || `HTTP ${response.status}` };
      }
      return { success: true, data: await response.json() };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function remove(id: string): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(`/api/v1/admin/feedback/${id}`, {
        method: "DELETE",
        headers: { ...getAuthHeader() },
        credentials: "include",
      });
      if (!response.ok) {
        const data = await response.json().catch(() => ({}));
        return { success: false, error: data.message || `HTTP ${response.status}` };
      }
      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  return { list, remove };
}
