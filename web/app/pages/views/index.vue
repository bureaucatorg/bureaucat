<script setup lang="ts">
import {
  Eye,
  Loader2,
  ChevronLeft,
  Link as LinkIcon,
  Check,
  Lock,
  Users as UsersIcon,
  Filter,
  Layers,
  FolderKanban,
} from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { Project, ProjectView } from "~/types";

definePageMeta({
  middleware: ["auth"],
});

useSeoMeta({ title: "Views" });

const { getAuthHeader } = useAuth();

interface ProjectWithViews {
  project: Project;
  views: ProjectView[];
}

const groups = ref<ProjectWithViews[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);

async function fetchAllProjects(): Promise<Project[]> {
  // Walk all pages so the list is complete even for users in many projects.
  const collected: Project[] = [];
  let page = 1;
  const perPage = 50;
  for (;;) {
    const resp = await fetch(
      `/api/v1/projects?page=${page}&per_page=${perPage}`,
      { headers: getAuthHeader() }
    );
    if (!resp.ok) throw new Error("Failed to fetch projects");
    const data = await resp.json();
    const batch: Project[] = data.projects ?? [];
    collected.push(...batch);
    if (page >= (data.total_pages ?? 1) || batch.length === 0) break;
    page += 1;
  }
  return collected;
}

async function fetchViewsFor(projectKey: string): Promise<ProjectView[]> {
  const resp = await fetch(`/api/v1/projects/${projectKey}/views`, {
    headers: getAuthHeader(),
  });
  if (!resp.ok) return [];
  return (await resp.json()) as ProjectView[];
}

async function loadAll() {
  loading.value = true;
  error.value = null;
  try {
    const projects = await fetchAllProjects();
    const results = await Promise.all(
      projects.map(async (p) => ({
        project: p,
        views: await fetchViewsFor(p.project_key),
      }))
    );
    // Only keep projects that have at least one accessible view.
    groups.value = results
      .filter((g) => g.views.length > 0)
      .sort((a, b) => a.project.name.localeCompare(b.project.name));
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Failed to load views";
  } finally {
    loading.value = false;
  }
}

const totalViews = computed(() =>
  groups.value.reduce((sum, g) => sum + g.views.length, 0)
);

function viewHref(projectKey: string, v: ProjectView): string {
  const q = new URLSearchParams();
  q.set("view", v.slug);
  if (v.default_tab && v.default_tab !== "tasks") q.set("tab", v.default_tab);
  return `/projects/${projectKey}?${q.toString()}`;
}

function predicateCount(v: ProjectView): number {
  return (v.filter_tree?.children ?? []).filter((c) => c.predicate).length;
}

function groupByLabel(g: string): string {
  if (!g || g === "none") return "None";
  return g.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

const copied = ref(false);
function copyLink() {
  navigator.clipboard.writeText(window.location.href);
  copied.value = true;
  toast.success("Link copied");
  setTimeout(() => {
    copied.value = false;
  }, 2000);
}

onMounted(loadAll);
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main id="main-content" class="flex-1">
      <div class="mx-auto max-w-6xl px-6 py-8">
        <nav class="mb-4 flex items-center gap-2 text-sm text-muted-foreground">
          <ChevronLeft class="size-4" />
          <span class="font-semibold text-amber-600 dark:text-amber-500">Views</span>
          <button
            aria-label="Copy link"
            class="ml-1 rounded-md p-1 text-muted-foreground/50 outline-none hover:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            @click="copyLink"
          >
            <Check v-if="copied" class="size-3.5 text-emerald-500" />
            <LinkIcon v-else class="size-3.5" />
          </button>
        </nav>

        <div class="mb-8 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 class="flex items-center gap-3 text-3xl font-bold tracking-tight">
              <Eye class="size-8" />
              Views
            </h1>
            <p class="mt-2 text-muted-foreground">
              Saved filter combinations across every project you belong to.
            </p>
          </div>
          <p v-if="!loading && totalViews > 0" class="text-sm text-muted-foreground">
            {{ totalViews }} view{{ totalViews === 1 ? "" : "s" }} across
            {{ groups.length }} project{{ groups.length === 1 ? "" : "s" }}
          </p>
        </div>

        <!-- Loading -->
        <div v-if="loading" class="flex items-center justify-center py-12">
          <Loader2 class="size-8 animate-spin text-muted-foreground" />
        </div>

        <!-- Error -->
        <div
          v-else-if="error"
          class="flex flex-col items-center justify-center rounded-lg border border-dashed py-16"
        >
          <p class="text-destructive">{{ error }}</p>
          <Button class="mt-4" variant="outline" @click="loadAll">
            Try again
          </Button>
        </div>

        <!-- Empty -->
        <div
          v-else-if="groups.length === 0"
          class="flex flex-col items-center justify-center rounded-lg border border-dashed py-16"
        >
          <Eye class="size-8 text-muted-foreground" />
          <h3 class="mt-4 font-semibold">No views yet</h3>
          <p class="mt-1 max-w-sm text-center text-sm text-muted-foreground">
            Open any project, apply some filters, and click
            <span class="font-medium">Save view</span> to create a reusable view.
          </p>
        </div>

        <!-- Groups -->
        <div v-else class="space-y-8">
          <section v-for="g in groups" :key="g.project.id">
            <div class="mb-3 flex items-center justify-between">
              <NuxtLink
                :to="`/projects/${g.project.project_key}`"
                class="group flex items-center gap-2 text-sm font-semibold hover:underline"
              >
                <FolderKanban class="size-4 text-muted-foreground group-hover:text-foreground" />
                <span>{{ g.project.name }}</span>
                <span class="rounded bg-muted px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                  {{ g.project.project_key }}
                </span>
              </NuxtLink>
              <span class="text-xs text-muted-foreground">
                {{ g.views.length }} view{{ g.views.length === 1 ? "" : "s" }}
              </span>
            </div>

            <div class="space-y-2">
              <NuxtLink
                v-for="v in g.views"
                :key="v.id"
                :to="viewHref(g.project.project_key, v)"
                class="group block rounded-lg border bg-card transition-colors hover:border-border/80 hover:bg-accent/30"
              >
                <div class="flex items-center gap-4 px-4 py-3">
                  <div class="flex size-8 shrink-0 items-center justify-center rounded-md bg-muted/50">
                    <component
                      :is="v.visibility === 'shared' ? UsersIcon : Lock"
                      class="size-3.5 text-muted-foreground"
                    />
                  </div>

                  <div class="min-w-0 flex-1">
                    <div class="flex items-center gap-2">
                      <span class="truncate font-medium group-hover:underline">{{ v.name }}</span>
                      <span
                        class="rounded-full bg-muted px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground"
                      >
                        {{ v.default_tab }}
                      </span>
                    </div>
                    <p
                      v-if="v.description"
                      class="mt-0.5 line-clamp-1 text-xs text-muted-foreground"
                    >
                      {{ v.description }}
                    </p>
                  </div>

                  <div class="hidden items-center gap-3 sm:flex">
                    <div
                      class="flex items-center gap-1.5 text-xs text-muted-foreground"
                      :title="`${predicateCount(v)} filter${predicateCount(v) === 1 ? '' : 's'}`"
                    >
                      <Filter class="size-3" />
                      <span>{{ predicateCount(v) }}</span>
                    </div>
                    <div
                      v-if="v.group_by && v.group_by !== 'none'"
                      class="flex items-center gap-1.5 text-xs text-muted-foreground"
                      :title="`Grouped by ${groupByLabel(v.group_by)}`"
                    >
                      <Layers class="size-3" />
                      <span>{{ groupByLabel(v.group_by) }}</span>
                    </div>
                  </div>
                </div>
              </NuxtLink>
            </div>
          </section>
        </div>
      </div>
    </main>
  </div>
</template>
