export default defineNuxtPlugin(async () => {
  const { fetchBranding } = useSettings();
  await fetchBranding();
});
