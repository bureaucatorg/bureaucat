<script setup lang="ts">
import { Loader2, Trash2, Lock, Check, Link, Cloud } from "lucide-vue-next";
import { toast } from "vue-sonner";
import { marked } from "marked";

const renderer = new marked.Renderer();
renderer.link = ({ href, title, text }) => {
  const titleAttr = title ? ` title="${title}"` : "";
  return `<a href="${href}"${titleAttr} target="_blank" rel="noopener noreferrer">${text}</a>`;
};
marked.setOptions({ breaks: true, gfm: true, renderer });

definePageMeta({
  middleware: ["auth"],
});

const route = useRoute();
const router = useRouter();

const projectKey = computed(() => route.params.key as string);
const pageNum = computed(() => parseInt(route.params.num as string));

const { currentProject, members, getProject, listMembers } = useProjects();
const { currentPage, getPage, updatePage, deletePage, clearCurrent } = usePages();
const { uploadFiles } = useFileAttach();

useHead({
  title: computed(() => currentPage.value?.title ?? `${projectKey.value} page`),
});

const loading = ref(true);
const error = ref<string | null>(null);
const deleting = ref(false);
const showDeleteConfirm = ref(false);
const copiedLink = ref(false);

// Editable state, kept in sync with the server via debounced autosave.
const title = ref("");
const content = ref("");
const lastSaved = reactive({ title: "", content: "" });
const ready = ref(false);
type SaveStatus = "idle" | "pending" | "saving" | "saved";
const saveStatus = ref<SaveStatus>("idle");
let saveTimer: ReturnType<typeof setTimeout> | null = null;

const isDisabled = computed(() => currentProject.value?.disabled ?? false);
const isMember = computed(
  () =>
    !isDisabled.value &&
    (currentProject.value?.role === "admin" || currentProject.value?.role === "member")
);

const renderedContent = computed(() => {
  const c = currentPage.value?.content;
  if (!c) return "";
  // If already HTML (from tiptap), render directly; otherwise convert markdown.
  return c.startsWith("<") ? c : (marked(c) as string);
});

async function loadData() {
  loading.value = true;
  error.value = null;

  if (!currentProject.value || currentProject.value.project_key !== projectKey.value) {
    const projectResult = await getProject(projectKey.value);
    if (!projectResult.success) {
      error.value = projectResult.error || "Failed to load project";
      loading.value = false;
      return;
    }
  }

  const pageResult = await getPage(projectKey.value, pageNum.value);
  if (!pageResult.success || !pageResult.data) {
    error.value = pageResult.error || "Page not found";
    loading.value = false;
    return;
  }

  title.value = pageResult.data.title;
  content.value = pageResult.data.content;
  lastSaved.title = pageResult.data.title;
  lastSaved.content = pageResult.data.content;
  saveStatus.value = "saved";

  await listMembers(projectKey.value);
  loading.value = false;
  // Defer enabling autosave until after the initial values settle so the
  // editor's own content-sync doesn't trigger a spurious save.
  nextTick(() => (ready.value = true));
}

function scheduleSave() {
  if (saveTimer) clearTimeout(saveTimer);
  saveTimer = setTimeout(flushSave, 1000);
}

async function flushSave() {
  if (saveTimer) {
    clearTimeout(saveTimer);
    saveTimer = null;
  }
  const nextTitle = title.value.trim();
  // Title is required; hold off (and keep the pending indicator) until non-empty.
  if (!nextTitle) return;
  if (nextTitle === lastSaved.title && content.value === lastSaved.content) {
    saveStatus.value = "saved";
    return;
  }
  saveStatus.value = "saving";
  const result = await updatePage(projectKey.value, pageNum.value, {
    title: nextTitle,
    content: content.value,
  });
  if (result.success) {
    lastSaved.title = nextTitle;
    lastSaved.content = content.value;
    saveStatus.value = "saved";
  } else {
    saveStatus.value = "pending";
    toast.error(result.error || "Failed to save");
  }
}

watch([title, content], () => {
  if (!ready.value || !isMember.value) return;
  saveStatus.value = "pending";
  scheduleSave();
});

async function handleDelete() {
  deleting.value = true;
  const result = await deletePage(projectKey.value, pageNum.value);
  deleting.value = false;
  if (!result.success) {
    toast.error(result.error || "Failed to delete page");
    return;
  }
  showDeleteConfirm.value = false;
  toast.success("Page deleted");
  router.push(`/projects/${projectKey.value}?tab=pages`);
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function copyLink() {
  navigator.clipboard.writeText(window.location.href);
  copiedLink.value = true;
  toast.success("Link copied");
  setTimeout(() => {
    copiedLink.value = false;
  }, 2000);
}

onMounted(loadData);
onBeforeUnmount(() => {
  // Flush any pending edits on navigation away (fire-and-forget).
  if (saveStatus.value === "pending") flushSave();
  clearCurrent();
});
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main id="main-content" class="flex-1">
      <div class="mx-auto max-w-6xl px-6 py-8">
        <!-- Loading -->
        <div v-if="loading" class="flex items-center justify-center py-20">
          <Loader2 class="size-8 animate-spin text-muted-foreground" />
        </div>

        <!-- Error -->
        <div v-else-if="error" class="flex flex-col items-center justify-center py-20">
          <p class="text-lg text-destructive">{{ error }}</p>
          <Button class="mt-4" variant="outline" as-child>
            <NuxtLink :to="`/projects/${projectKey}?tab=pages`">
              Back to Pages
            </NuxtLink>
          </Button>
        </div>

        <!-- Page content -->
        <template v-else-if="currentPage">
          <!-- Breadcrumb -->
          <nav class="mb-6 flex items-center gap-2 text-sm text-muted-foreground">
            <NuxtLink to="/projects" class="hover:text-foreground">Projects</NuxtLink>
            <span>/</span>
            <NuxtLink
              :to="`/projects/${projectKey}`"
              class="font-semibold text-amber-600 hover:text-amber-700 dark:text-amber-500 dark:hover:text-amber-400"
            >
              {{ projectKey }}
            </NuxtLink>
            <span>/</span>
            <NuxtLink
              :to="`/projects/${projectKey}?tab=pages`"
              class="hover:text-foreground"
            >
              Pages
            </NuxtLink>
            <button
              aria-label="Copy link"
              class="ml-1 rounded-md p-1 text-muted-foreground/50 hover:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none"
              @click="copyLink"
            >
              <Check v-if="copiedLink" class="size-3.5 text-emerald-500" />
              <Link v-else class="size-3.5" />
            </button>
          </nav>

          <div
            v-if="isDisabled"
            class="mb-6 flex items-center gap-2 rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-sm text-amber-700 dark:text-amber-400"
          >
            <Lock class="size-4 shrink-0" />
            <span>
              This project is disabled and read-only. No changes can be made until an admin re-enables it in Settings.
            </span>
          </div>

          <!-- Title -->
          <div class="flex items-start justify-between gap-3">
            <input
              v-if="isMember"
              v-model="title"
              placeholder="Untitled"
              class="w-full border-0 bg-transparent p-0 text-3xl font-bold outline-none placeholder:text-muted-foreground/50"
            />
            <h1 v-else class="text-3xl font-bold">{{ currentPage.title }}</h1>

            <Button
              v-if="isMember"
              variant="ghost"
              size="icon"
              aria-label="Delete page"
              class="mt-1 size-8 shrink-0 text-muted-foreground hover:text-destructive"
              @click="showDeleteConfirm = true"
            >
              <Trash2 class="size-4" />
            </Button>
          </div>

          <!-- Meta + autosave status -->
          <div class="mt-1.5 flex items-center gap-2 text-xs text-muted-foreground">
            <span>
              Created by {{ currentPage.creator_first_name }}
              {{ currentPage.creator_last_name }} · Updated
              {{ formatDate(currentPage.updated_at) }}
            </span>
            <template v-if="isMember">
              <span>·</span>
              <span class="flex items-center gap-1">
                <Loader2
                  v-if="saveStatus === 'saving'"
                  class="size-3 animate-spin"
                />
                <Cloud v-else class="size-3" />
                <span>
                  {{
                    saveStatus === "saving"
                      ? "Saving…"
                      : saveStatus === "pending"
                        ? "Unsaved changes"
                        : "Saved"
                  }}
                </span>
              </span>
            </template>
          </div>

          <!-- Content -->
          <div class="mt-6">
            <TiptapEditor
              v-if="isMember"
              v-model="content"
              borderless
              tables
              :members="members"
              :upload-handler="uploadFiles"
            />
            <div
              v-else-if="currentPage.content"
              class="prose prose-sm max-w-none dark:prose-invert"
              v-html="renderedContent"
            />
            <p v-else class="text-sm italic text-muted-foreground">
              This page has no content yet.
            </p>
          </div>
        </template>
      </div>
    </main>

    <Dialog v-model:open="showDeleteConfirm">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Delete page</DialogTitle>
          <DialogDescription>
            This will delete "{{ currentPage?.title }}". This action cannot be undone.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" :disabled="deleting" @click="showDeleteConfirm = false">
            Cancel
          </Button>
          <Button variant="destructive" :disabled="deleting" @click="handleDelete">
            <Loader2 v-if="deleting" class="mr-2 size-4 animate-spin" />
            Delete
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
