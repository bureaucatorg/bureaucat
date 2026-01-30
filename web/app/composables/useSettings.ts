export interface BrandingSettings {
  enabled: boolean;
  app_name: string;
}

const branding = ref<BrandingSettings>({
  enabled: false,
  app_name: "Bureaucat",
});

const brandingLoaded = ref(false);

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

  return {
    branding,
    appName,
    brandingLoaded,
    fetchBranding,
    updateBranding,
  };
}
