interface User {
  id: string;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  user_type: string;
  created_at: string;
}

interface PaginatedUsersResponse {
  users: User[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

interface TokenInfo {
  id: string;
  user_id: string;
  username: string;
  email: string;
  created_at: string;
  expires_at: string;
}

interface PaginatedTokensResponse {
  tokens: TokenInfo[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

interface CreateUserData {
  username: string;
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  user_type: string;
}

export function useAdmin() {
  const { getAuthHeader } = useAuth();

  async function listUsers(
    page = 1,
    perPage = 20,
    search = ""
  ): Promise<{
    success: boolean;
    data?: PaginatedUsersResponse;
    error?: string;
  }> {
    try {
      const params = new URLSearchParams({ page: String(page), per_page: String(perPage) });
      if (search) params.set("search", search);
      const response = await fetch(
        `/api/v1/admin/users?${params}`,
        { headers: getAuthHeader() }
      );

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to fetch users" };
      }

      const data = await response.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function createUser(userData: CreateUserData): Promise<{
    success: boolean;
    data?: User;
    error?: string;
  }> {
    try {
      const response = await fetch("/api/v1/admin/users", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify(userData),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to create user" };
      }

      const data = await response.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function deleteUser(userId: string): Promise<{
    success: boolean;
    error?: string;
  }> {
    try {
      const response = await fetch(`/api/v1/admin/users/${userId}`, {
        method: "DELETE",
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to delete user" };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function listTokens(
    page = 1,
    perPage = 20
  ): Promise<{
    success: boolean;
    data?: PaginatedTokensResponse;
    error?: string;
  }> {
    try {
      const response = await fetch(
        `/api/v1/admin/tokens?page=${page}&per_page=${perPage}`,
        { headers: getAuthHeader() }
      );

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to fetch tokens" };
      }

      const data = await response.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function revokeToken(tokenId: string): Promise<{
    success: boolean;
    error?: string;
  }> {
    try {
      const response = await fetch(`/api/v1/admin/tokens/${tokenId}`, {
        method: "DELETE",
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to revoke token" };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function cleanupExpiredTokens(): Promise<{
    success: boolean;
    deleted?: number;
    error?: string;
  }> {
    try {
      const response = await fetch("/api/v1/admin/tokens/expired", {
        method: "DELETE",
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to cleanup tokens" };
      }

      const data = await response.json();
      return { success: true, deleted: data.deleted };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function updateUserRole(userId: string, userType: string): Promise<{
    success: boolean;
    data?: User;
    error?: string;
  }> {
    try {
      const response = await fetch(`/api/v1/admin/users/${userId}/role`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify({ user_type: userType }),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to update role" };
      }

      const data = await response.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function resetUserPassword(userId: string, password: string): Promise<{
    success: boolean;
    error?: string;
  }> {
    try {
      const response = await fetch(`/api/v1/admin/users/${userId}/password`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify({ password }),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to reset password" };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  return {
    listUsers,
    createUser,
    deleteUser,
    updateUserRole,
    resetUserPassword,
    listTokens,
    revokeToken,
    cleanupExpiredTokens,
  };
}
