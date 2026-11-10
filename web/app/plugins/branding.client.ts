export default defineNuxtPlugin(async () => {
  const { fetchBranding, faviconUrl } = useSettings();
  await fetchBranding();

  useHead({
    link: computed(() =>
      faviconUrl.value
        ? [
            { key: "favicon", rel: "icon", href: faviconUrl.value },
            { key: "favicon-ico", rel: "icon", href: faviconUrl.value },
          ]
        : []
    ),
  });
});
