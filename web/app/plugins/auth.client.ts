export default defineNuxtPlugin(async () => {
  const { initAuth } = useAuth();
  await initAuth();
});
