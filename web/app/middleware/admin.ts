export default defineNuxtRouteMiddleware(async () => {
  const { isAuthenticated, isLoading, user } = useAuth();

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

  if (!isAuthenticated.value) {
    return navigateTo("/signin");
  }

  if (user.value?.user_type !== "admin") {
    return abortNavigation({
      statusCode: 403,
      statusMessage: "Admin access required",
    });
  }
});
