export default defineNuxtRouteMiddleware(async () => {
  const { isAuthenticated, isLoading } = useAuth();

  // Wait for auth initialization to complete
  if (isLoading.value) {
    await new Promise<void>((resolve) => {
      const checkAuth = () => {
        if (!isLoading.value) {
          resolve();
        } else {
          setTimeout(checkAuth, 50);
        }
      };
      checkAuth();
    });
  }

  if (isAuthenticated.value) {
    return navigateTo("/dashboard");
  }
});
