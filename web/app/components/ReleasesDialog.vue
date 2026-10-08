<script setup lang="ts">
import { ExternalLink, Loader2 } from "lucide-vue-next";
import { sanitizeHtml } from "~/utils/markdown";

interface Release {
  name: string;
  tag_name: string;
  html_url: string;
  body_html: string;
  published_at: string;
  prerelease: boolean;
}

const props = defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  "update:open": [value: boolean];
}>();

const RELEASES_PAGE = "https://github.com/bureaucatorg/bureaucat/releases";

const { getAuthHeader } = useAuth();

const releases = ref<Release[]>([]);
const currentVersion = ref("");
const loading = ref(false);
const error = ref(false);

const normalize = (v: string) => v.trim().replace(/^v/i, "");

function bodyHTML(html: string): string {
  return sanitizeHtml(html).replace(/<a /g, '<a target="_blank" rel="noopener noreferrer" ');
}

function formatDate(s: string): string {
  return new Date(s).toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" });
}

async function load() {
  loading.value = true;
  error.value = false;
  try {
    const res = await fetch("/api/v1/releases", {
      headers: { ...getAuthHeader() },
      credentials: "include",
    });
    if (!res.ok) throw new Error();
    const data = await res.json();
    releases.value = data.releases || [];
    currentVersion.value = data.current_version || "";
  } catch {
    error.value = true;
  } finally {
    loading.value = false;
  }
}

watch(
  () => props.open,
  (open) => {
    if (open && !releases.value.length) load();
  }
);
</script>

<template>
  <Dialog :open="open" @update:open="(v: boolean) => emit('update:open', v)">
    <DialogContent class="flex max-h-[85vh] flex-col gap-0 p-0 sm:max-w-2xl">
      <DialogHeader class="border-b px-6 py-4">
        <DialogTitle>What's new</DialogTitle>
        <DialogDescription>
          Release notes from
          <a
            :href="RELEASES_PAGE"
            target="_blank"
            rel="noopener noreferrer"
            class="font-medium underline underline-offset-2 hover:text-foreground"
          >GitHub</a>.
          <template v-if="currentVersion">
            This instance is running <span class="font-mono">{{ currentVersion }}</span>.
          </template>
        </DialogDescription>
      </DialogHeader>

      <div class="min-h-0 flex-1 overflow-y-auto px-6 py-4">
        <div v-if="loading" class="flex justify-center py-10">
          <Loader2 class="size-5 animate-spin text-muted-foreground" />
        </div>

        <div v-else-if="error" class="py-10 text-center text-sm text-muted-foreground">
          Couldn't load releases.
          <button type="button" class="font-medium underline underline-offset-2 hover:text-foreground" @click="load">
            Retry
          </button>
        </div>

        <p v-else-if="!releases.length" class="py-10 text-center text-sm text-muted-foreground">
          No releases found.
        </p>

        <div v-else class="divide-y">
          <article v-for="r in releases" :key="r.tag_name" class="py-5 first:pt-0 last:pb-0">
            <div class="flex flex-wrap items-center gap-2">
              <a
                :href="r.html_url"
                target="_blank"
                rel="noopener noreferrer"
                class="group inline-flex items-center gap-1 font-semibold hover:underline"
              >
                {{ r.name || r.tag_name }}
                <ExternalLink class="size-3.5 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100" />
              </a>
              <Badge
                v-if="currentVersion && normalize(r.tag_name) === normalize(currentVersion)"
                variant="secondary"
              >
                Installed
              </Badge>
              <Badge v-if="r.prerelease" variant="outline">Pre-release</Badge>
              <span class="ml-auto text-xs text-muted-foreground">{{ formatDate(r.published_at) }}</span>
            </div>
            <div
              v-if="r.body_html"
              class="prose mt-2 max-w-none break-words text-sm text-muted-foreground"
              v-html="bodyHTML(r.body_html)"
            />
          </article>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>
