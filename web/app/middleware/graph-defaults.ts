export default defineNuxtRouteMiddleware((to) => {
  const defaults: Record<string, string> = {};
  if (to.query.state_types === undefined) defaults.state_types = "unstarted,started,backlog";

  // The personal graph defaults to the viewer's own tasks; admin graph shows everyone.
  const { user } = useAuth();
  if (to.path === "/graph" && to.query.users === undefined && user.value?.username) {
    defaults.users = user.value.username;
  }

  if (Object.keys(defaults).length) {
    return navigateTo({ path: to.path, query: { ...to.query, ...defaults } }, { replace: true });
  }
});
