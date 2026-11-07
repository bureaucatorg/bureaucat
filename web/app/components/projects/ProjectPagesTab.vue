<script setup lang="ts">
import { Plus, FileText, Loader2, Search, X } from "lucide-vue-next";
import { toast } from "vue-sonner";

const props = defineProps<{
  projectKey: string;
  canWrite: boolean;
}>();

const { pages, loading, listPages, createPage } = usePages();

const showCreate = ref(false);
const creating = ref(false);
const newTitle = ref("");

const searchQuery = ref("");
const firstLoad = ref(true);
let searchTimer: ReturnType<typeof setTimeout> | null = null;

async function fetchPages() {
  await listPages(props.projectKey, searchQuery.value);
  firstLoad.value = false;
}

// Debounce search input so we don't fire a request per keystroke.
watch(searchQuery, () => {
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = setTimeout(fetchPages, 300);
});

function openCreate() {
  newTitle.value = "";
  showCreate.value = true;
}

async function submitCreate() {
  if (!newTitle.value.trim()) return;
  creating.value = true;
  const result = await createPage(props.projectKey, { title: newTitle.value.trim() });
  creating.value = false;
  if (!result.success || !result.data) {
    toast.error(result.error || "Failed to create page");
    return;
  }
  showCreate.value = false;
  navigateTo(`/projects/${props.projectKey}/pages/${result.data.page_number}`);
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

onMounted(fetchPages);
onDataChanged(fetchPages);

watch(
  () => props.projectKey,
  () => {
    searchQuery.value = "";
    firstLoad.value = true;
    fetchPages();
  }
);
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-lg font-semibold">Pages</h2>
        <p class="text-sm text-muted-foreground">
          Documentation for this project. Write specs, runbooks, and notes.
        </p>
      </div>
      <Button v-if="canWrite" @click="openCreate">
        <Plus class="mr-2 size-4" />
        Create Page
      </Button>
    </div>

    <!-- First load -->
    <div v-if="firstLoad && loading" class="flex items-center justify-center py-12">
      <Loader2 class="size-6 animate-spin text-muted-foreground" />
    </div>

    <!-- Project has no pages at all -->
    <div
      v-else-if="!loading && pages.length === 0 && !searchQuery"
      class="flex flex-col items-center justify-center rounded-lg border border-dashed py-16"
    >
      <div class="flex size-16 items-center justify-center rounded-full bg-muted">
        <FileText class="size-8 text-muted-foreground" />
      </div>
      <h3 class="mt-4 text-lg font-semibold">No pages yet</h3>
      <p class="mt-2 max-w-sm text-center text-sm text-muted-foreground">
        Create a page to document anything about this project.
      </p>
      <Button v-if="canWrite" class="mt-4" @click="openCreate">
        <Plus class="mr-2 size-4" />
        Create Page
      </Button>
    </div>

    <template v-else>
      <div class="relative">
        <Search
          class="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
        />
        <Input
          v-model="searchQuery"
          placeholder="Search pages by title or content..."
          class="pl-9 pr-9"
        />
        <button
          v-if="searchQuery"
          type="button"
          aria-label="Clear search"
          class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
          @click="searchQuery = ''"
        >
          <X class="size-4" />
        </button>
      </div>

      <div v-if="loading && pages.length === 0" class="flex items-center justify-center py-12">
        <Loader2 class="size-6 animate-spin text-muted-foreground" />
      </div>

      <div
        v-else-if="pages.length === 0"
        class="rounded-lg border border-dashed py-12 text-center text-sm text-muted-foreground"
      >
        No pages match "{{ searchQuery }}".
      </div>

      <div v-else class="divide-y rounded-lg border">
        <NuxtLink
          v-for="p in pages"
          :key="p.id"
          :to="`/projects/${projectKey}/pages/${p.page_number}`"
          class="flex items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/50"
        >
          <FileText class="size-4 shrink-0 text-muted-foreground" />
          <span class="flex-1 truncate font-medium">{{ p.title }}</span>
          <span class="shrink-0 text-xs text-muted-foreground">
            Updated {{ formatDate(p.updated_at) }}
          </span>
        </NuxtLink>
      </div>
    </template>

    <Dialog v-model:open="showCreate">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Create page</DialogTitle>
          <DialogDescription>
            Give your page a title. You can add content next.
          </DialogDescription>
        </DialogHeader>
        <form class="space-y-4" @submit.prevent="submitCreate">
          <div class="space-y-2">
            <Label for="page_title">Title</Label>
            <Input
              id="page_title"
              v-model="newTitle"
              placeholder="e.g. Onboarding guide"
              autofocus
              :disabled="creating"
            />
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              :disabled="creating"
              @click="showCreate = false"
            >
              Cancel
            </Button>
            <Button type="submit" :disabled="creating || !newTitle.trim()">
              <Loader2 v-if="creating" class="mr-2 size-4 animate-spin" />
              Create
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </div>
</template>
