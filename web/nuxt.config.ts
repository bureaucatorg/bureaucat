// https://nuxt.com/docs/api/configuration/nuxt-config
import tailwindcss from "@tailwindcss/vite";

export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",
  devtools: { enabled: true },
  css: ["~/assets/css/tailwind.css"],

  app: {
    head: {
      titleTemplate: "%s",
      meta: [
        { name: "description", content: "Bureaucracy That Actually Moves" },
        { property: "og:type", content: "website" },
        { property: "og:image", content: "/og.svg" },
        { name: "twitter:card", content: "summary_large_image" },
        { name: "twitter:image", content: "/og.svg" },
      ],
      htmlAttrs: { lang: "en" },
      link: [{ rel: "icon", href: "/favicon.ico" }],
    },
  },

  components: [
    {
      path: "~/components",
      pathPrefix: false,
    },
  ],

  runtimeConfig: {
    public: {
      nodeEnv: process.env.NODE_ENV || "development",
    },
  },

  vite: {
    plugins: [tailwindcss()],
  },

  modules: ["shadcn-nuxt", "@nuxtjs/color-mode"],

  colorMode: {
    classSuffix: "",
    preference: "system",
    storageKey: "bureaucat-color-mode",
  },

  ssr: false,

  shadcn: {
    /**
     * Prefix for all the imported component.
     * @default "Ui"
     */
    prefix: "",
    /**
     * Directory that the component lives in.
     * Will respect the Nuxt aliases.
     * @link https://nuxt.com/docs/api/nuxt-config#alias
     * @default "@/components/ui"
     */
    componentDir: "@/components/ui",
  },
});

