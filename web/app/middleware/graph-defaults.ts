export default defineNuxtRouteMiddleware((to) => {
  if (to.query.state_types === undefined) {
    return navigateTo({ path: to.path, query: { ...to.query, state_types: "unstarted,started,backlog" } }, { replace: true });
  }
});
