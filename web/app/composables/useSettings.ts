export interface BrandingSettings {
  enabled: boolean;
  app_name: string;
}

export interface SSOProviderConfig {
  enabled: boolean;
  client_id: string;
  client_secret: string;
  issuer_url?: string;
  redirect_uri: string;
}

export interface SSOSettings {
  google: SSOProviderConfig;
  zitadel: SSOProviderConfig;
}

export interface SSOProvidersPublic {
  google: boolean;
  zitadel: boolean;
  zitadel_url?: string;
}

export interface SignupSettings {
  enabled: boolean;
}

export interface MattermostSettings {
  enabled: boolean;
  server_url: string;
  bot_token: string;
}

export interface FeedbackSettings {
  receive_enabled: boolean;
  send_to_main_enabled: boolean;
  store_sent_locally: boolean;
}

export interface FeedbackPublicSettings {
  send_to_main_enabled: boolean;
  store_sent_locally: boolean;
}

const branding = ref<BrandingSettings>({
  enabled: false,
  app_name: "Bureaucat",
});

const brandingLoaded = ref(false);

const signupSettings = ref<SignupSettings>({ enabled: true });
const signupSettingsLoaded = ref(false);

const ssoProviders = ref<SSOProvidersPublic>({ google: false, zitadel: false });
const ssoProvidersLoaded = ref(false);

const feedbackPublic = ref<FeedbackPublicSettings>({
  send_to_main_enabled: false,
  store_sent_locally: true,
});
const feedbackPublicLoaded = ref(false);

export function useSettings() {
  const { getAuthHeader } = useAuth();

  // Computed app name - returns custom name if enabled, otherwise "Bureaucat"
  const appName = computed(() => {
    if (branding.value.enabled && branding.value.app_name) {
      return branding.value.app_name;
    }
    return "Bureaucat";
  });

  async function fetchBranding(): Promise<void> {
    if (brandingLoaded.value) return;

    try {
      const response = await fetch("/api/v1/settings/branding");
      if (response.ok) {
        const data = await response.json();
        branding.value = data;
      }
    } catch {
      // Use defaults on error
    } finally {
      brandingLoaded.value = true;
    }
  }

  async function updateBranding(
    settings: BrandingSettings
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch("/api/v1/admin/settings/branding", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        credentials: "include",
        body: JSON.stringify(settings),
      });

      if (!response.ok) {
        const data = await response.json();
        return { success: false, error: data.message || "Failed to update branding" };
      }

      const data = await response.json();
      branding.value = data;
      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  // --- Signup Settings ---

  async function fetchSignupSettings(): Promise<void> {
    if (signupSettingsLoaded.value) return;

    try {
      const response = await fetch("/api/v1/settings/signup");
      if (response.ok) {
        const data = await response.json();
        signupSettings.value = data;
      }
    } catch {
      // Default: enabled
    } finally {
      signupSettingsLoaded.value = true;
    }
  }

  async function updateSignupSettings(
    settings: SignupSettings
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch("/api/v1/admin/settings/signup", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        credentials: "include",
        body: JSON.stringify(settings),
      });

      if (!response.ok) {
        const data = await response.json();
        return { success: false, error: data.message || "Failed to update signup settings" };
      }

      const data = await response.json();
      signupSettings.value = data;
      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  // --- SSO Settings ---

  async function fetchSSOProviders(): Promise<void> {
    if (ssoProvidersLoaded.value) return;

    try {
      const response = await fetch("/api/v1/settings/sso");
      if (response.ok) {
        const data = await response.json();
        ssoProviders.value = data;
      }
    } catch {
      // Use defaults on error
    } finally {
      ssoProvidersLoaded.value = true;
    }
  }

  async function fetchSSOSettings(): Promise<{ success: boolean; data?: SSOSettings; error?: string }> {
    try {
      const response = await fetch("/api/v1/admin/settings/sso", {
        headers: { ...getAuthHeader() },
        credentials: "include",
      });

      if (!response.ok) {
        const data = await response.json();
        return { success: false, error: data.message || "Failed to fetch SSO settings" };
      }

      const data = await response.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function updateSSOSettings(
    settings: SSOSettings
  ): Promise<{ success: boolean; data?: SSOSettings; error?: string }> {
    try {
      const response = await fetch("/api/v1/admin/settings/sso", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        credentials: "include",
        body: JSON.stringify(settings),
      });

      if (!response.ok) {
        const data = await response.json();
        return { success: false, error: data.message || "Failed to update SSO settings" };
      }

      const data = await response.json();
      // Update public providers state too
      ssoProviders.value = {
        google: data.google?.enabled || false,
        zitadel: data.zitadel?.enabled || false,
      };
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  // --- Mattermost Settings ---

  async function fetchMattermostSettings(): Promise<{ success: boolean; data?: MattermostSettings; error?: string }> {
    try {
      const response = await fetch("/api/v1/admin/settings/mattermost", {
        headers: { ...getAuthHeader() },
        credentials: "include",
      });

      if (!response.ok) {
        const data = await response.json();
        return { success: false, error: data.message || "Failed to fetch Mattermost settings" };
      }

      const data = await response.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function updateMattermostSettings(
    settings: MattermostSettings
  ): Promise<{ success: boolean; data?: MattermostSettings; error?: string }> {
    try {
      const response = await fetch("/api/v1/admin/settings/mattermost", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        credentials: "include",
        body: JSON.stringify(settings),
      });

      if (!response.ok) {
        const data = await response.json();
        return { success: false, error: data.message || "Failed to update Mattermost settings" };
      }

      const data = await response.json();
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  // --- Feedback Settings ---

  async function fetchFeedbackPublicSettings(): Promise<void> {
    if (feedbackPublicLoaded.value) return;
    try {
      const response = await fetch("/api/v1/settings/feedback");
      if (response.ok) {
        feedbackPublic.value = await response.json();
      }
    } catch {
      // Stay with defaults.
    } finally {
      feedbackPublicLoaded.value = true;
    }
  }

  async function fetchFeedbackSettings(): Promise<{ success: boolean; data?: FeedbackSettings; error?: string }> {
    try {
      const response = await fetch("/api/v1/admin/settings/feedback", {
        headers: { ...getAuthHeader() },
        credentials: "include",
      });
      if (!response.ok) {
        const data = await response.json();
        return { success: false, error: data.message || "Failed to fetch feedback settings" };
      }
      return { success: true, data: await response.json() };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function updateFeedbackSettings(
    settings: FeedbackSettings
  ): Promise<{ success: boolean; data?: FeedbackSettings; error?: string }> {
    try {
      const response = await fetch("/api/v1/admin/settings/feedback", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        credentials: "include",
        body: JSON.stringify(settings),
      });
      if (!response.ok) {
        const data = await response.json();
        return { success: false, error: data.message || "Failed to update feedback settings" };
      }
      const data = await response.json();
      // Keep the public-facing cache in sync so the sidebar updates without reload.
      feedbackPublic.value = {
        send_to_main_enabled: data.send_to_main_enabled,
        store_sent_locally: data.store_sent_locally,
      };
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function testMattermostConnection(): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch("/api/v1/admin/settings/mattermost/test", {
        method: "POST",
        headers: { ...getAuthHeader() },
        credentials: "include",
      });

      if (!response.ok) {
        const data = await response.json();
        return { success: false, error: data.message || "Connection test failed" };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  return {
    branding,
    appName,
    brandingLoaded,
    fetchBranding,
    updateBranding,
    signupSettings,
    signupSettingsLoaded,
    fetchSignupSettings,
    updateSignupSettings,
    ssoProviders,
    ssoProvidersLoaded,
    fetchSSOProviders,
    fetchSSOSettings,
    updateSSOSettings,
    fetchMattermostSettings,
    updateMattermostSettings,
    testMattermostConnection,
    feedbackPublic,
    feedbackPublicLoaded,
    fetchFeedbackPublicSettings,
    fetchFeedbackSettings,
    updateFeedbackSettings,
  };
}
