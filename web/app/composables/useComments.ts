import type { Comment, CreateCommentRequest, UpdateCommentRequest } from "~/types";

interface CommentsState {
  comments: Comment[];
  loading: boolean;
}

const state = reactive<CommentsState>({
  comments: [],
  loading: false,
});

export function useComments() {
  const { getAuthHeader } = useAuth();

  async function listComments(
    projectKey: string,
    taskNum: number
  ): Promise<{ success: boolean; data?: Comment[]; error?: string }> {
    try {
      state.loading = true;
      const response = await fetch(
        `/api/v1/projects/${projectKey}/tasks/${taskNum}/comments`,
        { headers: getAuthHeader() }
      );

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to fetch comments" };
      }

      const comments: Comment[] = await response.json();
      state.comments = comments;
      return { success: true, data: comments };
    } catch {
      return { success: false, error: "Network error" };
    } finally {
      state.loading = false;
    }
  }

  async function createComment(
    projectKey: string,
    taskNum: number,
    data: CreateCommentRequest
  ): Promise<{ success: boolean; data?: Comment; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/tasks/${taskNum}/comments`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            ...getAuthHeader(),
          },
          body: JSON.stringify(data),
        }
      );

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to create comment" };
      }

      const comment: Comment = await response.json();
      state.comments.push(comment);
      return { success: true, data: comment };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function updateComment(
    projectKey: string,
    taskNum: number,
    commentId: string,
    data: UpdateCommentRequest
  ): Promise<{ success: boolean; data?: Comment; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/tasks/${taskNum}/comments/${commentId}`,
        {
          method: "PATCH",
          headers: {
            "Content-Type": "application/json",
            ...getAuthHeader(),
          },
          body: JSON.stringify(data),
        }
      );

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to update comment" };
      }

      const comment: Comment = await response.json();
      const idx = state.comments.findIndex((c) => c.id === commentId);
      if (idx !== -1) {
        state.comments[idx] = comment;
      }
      return { success: true, data: comment };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function deleteComment(
    projectKey: string,
    taskNum: number,
    commentId: string
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(
        `/api/v1/projects/${projectKey}/tasks/${taskNum}/comments/${commentId}`,
        {
          method: "DELETE",
          headers: getAuthHeader(),
        }
      );

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to delete comment" };
      }

      state.comments = state.comments.filter((c) => c.id !== commentId);
      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  function clearComments() {
    state.comments = [];
  }

  return {
    // State (readonly)
    comments: computed(() => state.comments),
    loading: computed(() => state.loading),

    // Methods
    listComments,
    createComment,
    updateComment,
    deleteComment,
    clearComments,
  };
}
