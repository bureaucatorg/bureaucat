<script setup lang="ts">
import { Search, ListTodo, Calendar, FolderKanban, CornerDownLeft, Loader2 } from "lucide-vue-next";
import type {
  SearchTaskResult,
  SearchCycleResult,
  SearchProjectResult,
} from "~/composables/useSearch";

const open = defineModel<boolean>("open", { default: false });

const { query, results, loading, search, reset } = useSearch();

type FlatItem =
  | { type: "task"; data: SearchTaskResult }
  | { type: "cycle"; data: SearchCycleResult }
  | { type: "project"; data: SearchProjectResult };

const flatItems = computed<FlatItem[]>(() => [
  ...results.value.tasks.map((t) => ({ type: "task" as const, data: t })),
  ...results.value.cycles.map((c) => ({ type: "cycle" as const, data: c })),
  ...results.value.projects.map((p) => ({ type: "project" as const, data: p })),
]);

const activeIndex = ref(0);
const inputRef = ref<HTMLInputElement | null>(null);
const listRef = ref<HTMLElement | null>(null);

function onInput(e: Event) {
  const value = (e.target as HTMLInputElement).value;
  search(value);
  activeIndex.value = 0;
}

watch(flatItems, () => {
  if (activeIndex.value >= flatItems.value.length) {
    activeIndex.value = Math.max(0, flatItems.value.length - 1);
  }
});

function hrefFor(item: FlatItem): string {
  if (item.type === "task") {
    return `/projects/${item.data.project_key}/tasks/${item.data.task_number}`;
  }
  if (item.type === "cycle") {
    return `/projects/${item.data.project_key}/cycles/${item.data.id}`;
  }
  return `/projects/${item.data.project_key}`;
}

function selectItem(item: FlatItem) {
  navigateTo(hrefFor(item));
  open.value = false;
}

function onKeydown(e: KeyboardEvent) {
  if (!open.value) return;
  if (e.key === "ArrowDown") {
    e.preventDefault();
    if (flatItems.value.length === 0) return;
    activeIndex.value = (activeIndex.value + 1) % flatItems.value.length;
    scrollActiveIntoView();
  } else if (e.key === "ArrowUp") {
    e.preventDefault();
    if (flatItems.value.length === 0) return;
    activeIndex.value =
      (activeIndex.value - 1 + flatItems.value.length) % flatItems.value.length;
    scrollActiveIntoView();
  } else if (e.key === "Enter") {
    const item = flatItems.value[activeIndex.value];
    if (item) {
      e.preventDefault();
      selectItem(item);
    }
  }
}

function scrollActiveIntoView() {
  nextTick(() => {
    const el = listRef.value?.querySelector<HTMLElement>(
      `[data-index="${activeIndex.value}"]`,
    );
    el?.scrollIntoView({ block: "nearest" });
  });
}

// Ctrl+K / Cmd+K to open globally
function onGlobalKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && (e.key === "k" || e.key === "K")) {
    e.preventDefault();
    open.value = !open.value;
  }
}

onMounted(() => {
  window.addEventListener("keydown", onGlobalKeydown);
});

onBeforeUnmount(() => {
  window.removeEventListener("keydown", onGlobalKeydown);
});

watch(open, (isOpen) => {
  if (isOpen) {
    activeIndex.value = 0;
    nextTick(() => inputRef.value?.focus());
  } else {
    reset();
  }
});

const hasResults = computed(() => flatItems.value.length > 0);
const showEmpty = computed(
  () => !loading.value && query.value.trim().length > 0 && !hasResults.value,
);
const showIdle = computed(() => query.value.trim().length === 0);

// Index offsets so rows can compute their flat index for activeIndex comparison
const taskOffset = 0;
const cycleOffset = computed(() => results.value.tasks.length);
const projectOffset = computed(
  () => results.value.tasks.length + results.value.cycles.length,
);

const isMac = computed(() => {
  if (typeof navigator === "undefined") return false;
  return /Mac|iPhone|iPad/.test(navigator.platform);
});
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent
      :show-close-button="false"
      class="top-[12%] translate-y-0 sm:max-w-xl p-0 gap-0 overflow-hidden border-border/60"
      @keydown="onKeydown"
    >
      <!-- visually-hidden title/description for a11y (DialogContent requires them) -->
      <DialogTitle class="sr-only">Global search</DialogTitle>
      <DialogDescription class="sr-only">
        Search across tasks, cycles, and projects you have access to.
      </DialogDescription>

      <div class="flex items-center gap-3 border-b border-border/60 px-4 py-3">
        <Search class="size-4 text-muted-foreground shrink-0" />
        <input
          ref="inputRef"
          :value="query"
          type="text"
          placeholder="Search tasks, cycles, projects…"
          autocomplete="off"
          spellcheck="false"
          class="w-full bg-transparent text-[15px] placeholder:text-muted-foreground focus:outline-none"
          @input="onInput"
        />
        <Loader2 v-if="loading" class="size-4 animate-spin text-muted-foreground" />
        <kbd
          class="hidden sm:inline-flex h-5 select-none items-center gap-1 rounded border border-border/60 bg-muted/40 px-1.5 font-mono text-[10px] font-medium text-muted-foreground"
        >
          esc
        </kbd>
      </div>

      <div
        ref="listRef"
        class="max-h-[min(440px,60vh)] overflow-y-auto py-1"
      >
        <div v-if="showIdle" class="px-4 py-10 text-center text-sm text-muted-foreground">
          Start typing to search across everything you can access.
        </div>

        <div v-else-if="showEmpty" class="px-4 py-10 text-center text-sm text-muted-foreground">
          No matches for <span class="text-foreground">"{{ query }}"</span>.
        </div>

        <template v-else>
          <!-- Tasks -->
          <div v-if="results.tasks.length" class="mb-1">
            <div
              class="px-4 pt-2 pb-1 text-[10px] font-semibold uppercase tracking-[0.12em] text-muted-foreground/80"
            >
              Tasks
            </div>
            <button
              v-for="(task, i) in results.tasks"
              :key="task.id"
              type="button"
              :data-index="taskOffset + i"
              class="w-full flex items-center gap-3 px-4 py-2 text-left transition-colors"
              :class="
                activeIndex === taskOffset + i
                  ? 'bg-accent/70'
                  : 'hover:bg-accent/40'
              "
              @click="selectItem({ type: 'task', data: task })"
              @mouseenter="activeIndex = taskOffset + i"
            >
              <ListTodo class="size-4 shrink-0 text-muted-foreground" />
              <div class="min-w-0 flex-1 flex items-center gap-2">
                <span
                  class="font-mono text-[11px] text-muted-foreground shrink-0 tabular-nums"
                >
                  {{ task.task_key }}
                </span>
                <span class="truncate text-sm">{{ task.title }}</span>
              </div>
              <div
                class="shrink-0 hidden sm:flex items-center gap-1.5 text-[11px] text-muted-foreground"
              >
                <span
                  class="inline-block size-1.5 rounded-full"
                  :style="{ backgroundColor: task.state_color }"
                />
                <span>{{ task.state_name }}</span>
              </div>
            </button>
          </div>

          <!-- Cycles -->
          <div v-if="results.cycles.length" class="mb-1">
            <div
              class="px-4 pt-2 pb-1 text-[10px] font-semibold uppercase tracking-[0.12em] text-muted-foreground/80"
            >
              Cycles
            </div>
            <button
              v-for="(cycle, i) in results.cycles"
              :key="cycle.id"
              type="button"
              :data-index="cycleOffset + i"
              class="w-full flex items-center gap-3 px-4 py-2 text-left transition-colors"
              :class="
                activeIndex === cycleOffset + i
                  ? 'bg-accent/70'
                  : 'hover:bg-accent/40'
              "
              @click="selectItem({ type: 'cycle', data: cycle })"
              @mouseenter="activeIndex = cycleOffset + i"
            >
              <Calendar class="size-4 shrink-0 text-muted-foreground" />
              <div class="min-w-0 flex-1">
                <span class="truncate text-sm">{{ cycle.title }}</span>
              </div>
              <div
                class="shrink-0 hidden sm:flex items-center gap-1.5 text-[11px] text-muted-foreground"
              >
                <span class="font-mono">{{ cycle.project_key }}</span>
                <span class="text-muted-foreground/60">·</span>
                <span>{{ cycle.start_date }} → {{ cycle.end_date }}</span>
              </div>
            </button>
          </div>

          <!-- Projects -->
          <div v-if="results.projects.length" class="mb-1">
            <div
              class="px-4 pt-2 pb-1 text-[10px] font-semibold uppercase tracking-[0.12em] text-muted-foreground/80"
            >
              Projects
            </div>
            <button
              v-for="(project, i) in results.projects"
              :key="project.id"
              type="button"
              :data-index="projectOffset + i"
              class="w-full flex items-center gap-3 px-4 py-2 text-left transition-colors"
              :class="
                activeIndex === projectOffset + i
                  ? 'bg-accent/70'
                  : 'hover:bg-accent/40'
              "
              @click="selectItem({ type: 'project', data: project })"
              @mouseenter="activeIndex = projectOffset + i"
            >
              <img
                v-if="project.icon_url"
                :src="project.icon_url"
                :alt="project.name"
                class="size-4 shrink-0 rounded object-cover"
              />
              <FolderKanban v-else class="size-4 shrink-0 text-muted-foreground" />
              <div class="min-w-0 flex-1 flex items-center gap-2">
                <span
                  class="font-mono text-[11px] text-muted-foreground shrink-0"
                >
                  {{ project.project_key }}
                </span>
                <span class="truncate text-sm">{{ project.name }}</span>
              </div>
              <div
                v-if="project.description"
                class="shrink-0 hidden md:block max-w-[16rem] truncate text-[11px] text-muted-foreground"
              >
                {{ project.description }}
              </div>
            </button>
          </div>
        </template>
      </div>

      <div
        class="flex items-center justify-between border-t border-border/60 bg-muted/20 px-4 py-2 text-[11px] text-muted-foreground"
      >
        <div class="flex items-center gap-3">
          <span class="flex items-center gap-1">
            <kbd
              class="inline-flex h-4 min-w-4 items-center justify-center rounded border border-border/60 bg-background px-1 font-mono text-[10px]"
              >↑</kbd
            >
            <kbd
              class="inline-flex h-4 min-w-4 items-center justify-center rounded border border-border/60 bg-background px-1 font-mono text-[10px]"
              >↓</kbd
            >
            navigate
          </span>
          <span class="flex items-center gap-1">
            <kbd
              class="inline-flex h-4 min-w-4 items-center justify-center rounded border border-border/60 bg-background px-1 font-mono text-[10px]"
            >
              <CornerDownLeft class="size-2.5" />
            </kbd>
            open
          </span>
        </div>
        <span class="flex items-center gap-1">
          <kbd
            class="inline-flex h-4 items-center rounded border border-border/60 bg-background px-1 font-mono text-[10px]"
          >
            {{ isMac ? "⌘" : "Ctrl" }}
          </kbd>
          <kbd
            class="inline-flex h-4 min-w-4 items-center justify-center rounded border border-border/60 bg-background px-1 font-mono text-[10px]"
            >K</kbd
          >
          to toggle
        </span>
      </div>
    </DialogContent>
  </Dialog>
</template>
