import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig(({ isSsrBuild }) => ({
  plugins: [vue(), tailwindcss()],
  server: { port: 3042 },
  build: { copyPublicDir: !isSsrBuild },
  ssr: { noExternal: ["lucide-vue-next"] },
}));
