<script setup lang="ts">
import { Check, ChevronsUpDown, Loader2, Play, Search } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { LocationQueryRaw } from "vue-router";
import { VueFlow, useVueFlow, type Edge, type Node, type NodeMouseEvent } from "@vue-flow/core";
import { Background } from "@vue-flow/background";
import { Controls } from "@vue-flow/controls";
import "@vue-flow/core/dist/style.css";
import "@vue-flow/core/dist/theme-default.css";
import "@vue-flow/controls/dist/style.css";
import { Button } from "~/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "~/components/ui/popover";
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "~/components/ui/command";
import { TooltipProvider } from "~/components/ui/tooltip";
import type { GraphTask, GraphUser, HoveredGraphNode, TaskGraph, TaskGraphFilterOptions } from "~/composables/useAdmin";
import { STATE_TYPE_OPTIONS, stateTypeLabel } from "~/components/filters/filterCatalog";
import { STATE_TYPE_COLORS } from "~/types/task";

definePageMeta({
  middleware: [
    "admin",
    (to) => {
      if (to.query.state_types === undefined) {
        return navigateTo({ path: to.path, query: { ...to.query, state_types: "unstarted,started,backlog" } }, { replace: true });
      }
    },
  ],
});

useSeoMeta({ title: "Graph View" });

const route = useRoute();
const router = useRouter();
const { getTaskGraph, getTaskGraphFilters } = useAdmin();
const { fitView, findNode, setCenter } = useVueFlow();

const loading = ref(false);
const graph = shallowRef<TaskGraph | null>(null);
const options = shallowRef<TaskGraphFilterOptions | null>(null);
const appliedKey = ref<string | null>(null);
const workspaceOpen = ref(false);
const projectsOpen = ref(false);
const stateTypesOpen = ref(false);
const usersOpen = ref(false);
const userSearchOpen = ref(false);
const hovered = ref<HoveredGraphNode | null>(null);
const hoveredEl = ref<HTMLElement | null>(null);
let hoverTimer: ReturnType<typeof setTimeout> | undefined;
let requestId = 0;

onMounted(async () => {
  const result = await getTaskGraphFilters();
  if (result.success && result.data) {
    options.value = result.data;
  } else {
    toast.error(result.error || "Failed to load filters");
  }
});

async function runGraph() {
  const id = ++requestId;
  const key = filtersKey.value;
  loading.value = true;
  hideTooltip();
  const result = await getTaskGraph({
    workspace: selectedWorkspace.value,
    projects: selectedProjects.value,
    stateTypes: selectedStateTypes.value,
    users: selectedUsers.value,
  });
  if (id !== requestId) return;
  if (result.success && result.data) {
    graph.value = result.data;
    appliedKey.value = key;
  } else {
    toast.error(result.error || "Failed to load graph");
  }
  loading.value = false;
}

const selectedWorkspace = computed(() =>
  typeof route.query.workspace === "string" ? route.query.workspace : "",
);
function listParam(name: string): string[] {
  const value = route.query[name];
  return typeof value === "string" && value ? value.split(",") : [];
}

const selectedProjects = computed(() => listParam("projects"));
const selectedStateTypes = computed(() => listParam("state_types"));
const selectedUsers = computed(() => listParam("users"));

const filtersKey = computed(() =>
  JSON.stringify([
    selectedWorkspace.value,
    [...selectedProjects.value].sort(),
    [...selectedStateTypes.value].sort(),
    [...selectedUsers.value].sort(),
  ]),
);
const canRun = computed(() => !loading.value && (!graph.value || filtersKey.value !== appliedKey.value));

function updateQuery(patch: { workspace?: string; projects?: string[]; stateTypes?: string[]; users?: string[] }) {
  const query: LocationQueryRaw = { ...route.query };
  if (patch.workspace !== undefined) query.workspace = patch.workspace || undefined;
  if (patch.projects) query.projects = patch.projects.length ? patch.projects.join(",") : undefined;
  // An empty value means "all"; a missing one triggers the default redirect.
  if (patch.stateTypes) query.state_types = patch.stateTypes.join(",");
  if (patch.users) query.users = patch.users.length ? patch.users.join(",") : undefined;
  router.replace({ query });
}

function toggle(list: string[], value: string): string[] {
  return list.includes(value) ? list.filter((v) => v !== value) : [...list, value];
}

function selectWorkspace(key: string) {
  workspaceOpen.value = false;
  if (key !== selectedWorkspace.value) updateQuery({ workspace: key, projects: [] });
}

const workspaces = computed(() => options.value?.workspaces ?? []);
const userOptions = computed(() => options.value?.users ?? []);
const projects = computed(() =>
  (options.value?.projects ?? []).filter(
    (p) => !selectedWorkspace.value || p.workspace_key === selectedWorkspace.value,
  ),
);

const stateTypes = STATE_TYPE_OPTIONS;

const workspaceLabel = computed(
  () => workspaces.value.find((w) => w.key === selectedWorkspace.value)?.name ?? "All workspaces",
);
const projectsLabel = computed(() => {
  const count = selectedProjects.value.length;
  if (count === 0) return "All projects";
  if (count === 1) return selectedProjects.value[0];
  return `${count} projects`;
});
const stateTypesLabel = computed(() => {
  const count = selectedStateTypes.value.length;
  if (count === 0) return "All state types";
  if (count === 1) return stateTypeLabel(selectedStateTypes.value[0]!);
  return `${count} state types`;
});
// Command mounts several components per row, so long lists are filtered here and capped.
const OPTION_LIMIT = 50;

function capOptions<T>(index: { item: T; text: string }[], query: string, isSelected: (item: T) => boolean) {
  const q = query.trim().toLowerCase();
  const picked: T[] = [];
  const rest: T[] = [];
  let total = 0;
  for (const { item, text } of index) {
    if (q && !text.includes(q)) continue;
    total++;
    if (isSelected(item)) picked.push(item);
    else if (rest.length < OPTION_LIMIT) rest.push(item);
  }
  const items = [...picked, ...rest].slice(0, Math.max(OPTION_LIMIT, picked.length));
  return { items, total };
}

const workspaceQuery = ref("");
const projectQuery = ref("");
const userQuery = ref("");

watch(workspaceOpen, () => (workspaceQuery.value = ""));
watch(projectsOpen, () => (projectQuery.value = ""));
watch(usersOpen, () => (userQuery.value = ""));

const workspaceIndex = computed(() =>
  workspaces.value.map((w) => ({ item: w, text: `${w.name} ${w.key}`.toLowerCase() })),
);
const projectIndex = computed(() =>
  projects.value.map((p) => ({ item: p, text: `${p.key} ${p.name}`.toLowerCase() })),
);
const userIndex = computed(() =>
  userOptions.value.map((u) => ({
    item: u,
    text: `${u.first_name} ${u.last_name} ${u.username} ${u.email}`.toLowerCase(),
  })),
);

const workspaceMatches = computed(() =>
  capOptions(workspaceIndex.value, workspaceQuery.value, (w) => w.key === selectedWorkspace.value),
);
const projectMatches = computed(() => {
  const selected = new Set(selectedProjects.value);
  return capOptions(projectIndex.value, projectQuery.value, (p) => selected.has(p.key));
});
const userMatches = computed(() => {
  const selected = new Set(selectedUsers.value);
  return capOptions(userIndex.value, userQuery.value, (u) => selected.has(u.username));
});

function toggleProject(key: string) {
  projectQuery.value = "";
  updateQuery({ projects: toggle(selectedProjects.value, key) });
}

function toggleUser(username: string) {
  userQuery.value = "";
  updateQuery({ users: toggle(selectedUsers.value, username) });
}

const usersLabel = computed(() => {
  const count = selectedUsers.value.length;
  if (count === 0) return "All users";
  if (count === 1) {
    const u = userOptions.value.find((o) => o.username === selectedUsers.value[0]);
    return u ? `${u.first_name} ${u.last_name}` : selectedUsers.value[0];
  }
  return `${count} users`;
});

const filtered = computed(() => {
  const tasks = graph.value?.tasks ?? [];
  const edges = graph.value?.edges ?? [];
  const taskCounts = new Map<string, number>();
  for (const e of edges) taskCounts.set(e.user_id, (taskCounts.get(e.user_id) ?? 0) + 1);
  const users = (graph.value?.users ?? [])
    .filter((u) => taskCounts.has(u.id))
    .map((u) => ({ ...u, task_count: taskCounts.get(u.id)! }));
  return { users, tasks, edges };
});

const flow = computed(() => {
  const { users, tasks, edges } = filtered.value;

  const tasksByUser = new Map<string, string[]>();
  for (const e of edges) {
    const list = tasksByUser.get(e.user_id) ?? [];
    list.push(e.task_id);
    tasksByUser.set(e.user_id, list);
  }

  // Seed order places each user next to its tasks so the layout starts clustered.
  const order: string[] = [];
  const placed = new Set<string>();
  for (const u of users) {
    order.push(`u:${u.id}`);
    for (const taskId of tasksByUser.get(u.id) ?? []) {
      if (placed.has(taskId)) continue;
      placed.add(taskId);
      order.push(`t:${taskId}`);
    }
  }

  const positions = forceLayout(
    order,
    edges.map((e) => [`u:${e.user_id}`, `t:${e.task_id}`]),
  );

  const nodes: Node[] = [
    ...users.map((u) => ({ id: `u:${u.id}`, type: "user", position: positions.get(`u:${u.id}`)!, data: u })),
    ...tasks.map((t) => ({ id: `t:${t.id}`, type: "task", position: positions.get(`t:${t.id}`)!, data: t })),
  ];
  const flowEdges: Edge[] = edges.map((e) => ({
    id: `${e.user_id}:${e.task_id}`,
    source: `u:${e.user_id}`,
    target: `t:${e.task_id}`,
    type: "straight",
    style: { stroke: "var(--muted-foreground)", strokeOpacity: 0.4 },
  }));

  return { nodes, edges: flowEdges };
});

const sortedUsers = computed(() =>
  [...filtered.value.users].sort((a, b) =>
    `${a.first_name} ${a.last_name}`.localeCompare(`${b.first_name} ${b.last_name}`),
  ),
);

function focusUser(user: GraphUser) {
  userSearchOpen.value = false;
  const node = findNode(`u:${user.id}`);
  if (!node) return;
  setCenter(node.position.x + node.dimensions.width / 2, node.position.y + node.dimensions.height / 2, {
    zoom: 1.5,
    duration: 600,
  });
}

function hideTooltip() {
  clearTimeout(hoverTimer);
  hovered.value = null;
  hoveredEl.value = null;
}

function onNodeMouseEnter({ node }: NodeMouseEvent) {
  clearTimeout(hoverTimer);
  hoverTimer = setTimeout(() => {
    hoveredEl.value = document.querySelector<HTMLElement>(`.vue-flow__node[data-id="${CSS.escape(node.id)}"]`);
    hovered.value = { type: node.type, data: node.data } as HoveredGraphNode;
  }, 150);
}

function onNodeClick({ node }: NodeMouseEvent) {
  hideTooltip();
  if (node.type === "task") {
    const t = node.data as GraphTask;
    navigateTo(`/projects/${t.project_key}/tasks/${t.task_number}`);
  } else {
    navigateTo(`/profile/${(node.data as GraphUser).id}`);
  }
}
</script>

<template>
  <div class="flex h-screen flex-col">
    <Navbar />

    <main id="main-content" class="relative min-h-0 flex-1">
      <div class="absolute inset-x-3 top-3 z-10 flex flex-wrap items-center gap-2">
        <Popover v-model:open="workspaceOpen">
          <PopoverTrigger as-child>
            <Button variant="outline" size="sm" class="max-w-48 justify-between gap-2 bg-background/90 shadow-sm backdrop-blur">
              <span class="truncate">{{ workspaceLabel }}</span>
              <ChevronsUpDown class="size-3.5 shrink-0 opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent align="start" class="w-64 p-0">
            <Command>
              <CommandInput placeholder="Search workspaces..." @input="workspaceQuery = ($event.target as HTMLInputElement).value" />
              <CommandList>
                <CommandEmpty>No workspace found.</CommandEmpty>
                <CommandGroup>
                  <CommandItem value="all-workspaces" @select="selectWorkspace('')">
                    <Check :class="['size-4', selectedWorkspace ? 'opacity-0' : 'opacity-100']" />
                    All workspaces
                  </CommandItem>
                  <CommandItem v-for="w in workspaceMatches.items" :key="w.key" :value="`${w.name} ${w.key}`" @select="selectWorkspace(w.key)">
                    <Check :class="['size-4', selectedWorkspace === w.key ? 'opacity-100' : 'opacity-0']" />
                    <span class="truncate">{{ w.name }}</span>
                  </CommandItem>
                </CommandGroup>
                <p v-if="workspaceMatches.total > workspaceMatches.items.length" class="px-3 py-2 text-xs text-muted-foreground">
                  Showing {{ workspaceMatches.items.length }} of {{ workspaceMatches.total }}. Type to narrow down.
                </p>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>

        <Popover v-model:open="projectsOpen">
          <PopoverTrigger as-child>
            <Button variant="outline" size="sm" class="max-w-48 justify-between gap-2 bg-background/90 shadow-sm backdrop-blur">
              <span class="truncate">{{ projectsLabel }}</span>
              <ChevronsUpDown class="size-3.5 shrink-0 opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent align="start" class="w-72 p-0">
            <Command>
              <CommandInput placeholder="Search projects..." @input="projectQuery = ($event.target as HTMLInputElement).value" />
              <CommandList>
                <CommandEmpty>No project found.</CommandEmpty>
                <CommandGroup>
                  <CommandItem
                    v-if="selectedProjects.length"
                    value="clear-projects"
                    class="text-muted-foreground"
                    @select="updateQuery({ projects: [] })"
                  >
                    Clear selection
                  </CommandItem>
                  <CommandItem v-for="p in projectMatches.items" :key="p.key" :value="`${p.key} ${p.name}`" @select="toggleProject(p.key)">
                    <Check :class="['size-4', selectedProjects.includes(p.key) ? 'opacity-100' : 'opacity-0']" />
                    <span class="font-mono text-xs">{{ p.key }}</span>
                    <span class="truncate text-muted-foreground">{{ p.name }}</span>
                  </CommandItem>
                </CommandGroup>
                <p v-if="projectMatches.total > projectMatches.items.length" class="px-3 py-2 text-xs text-muted-foreground">
                  Showing {{ projectMatches.items.length }} of {{ projectMatches.total }}. Type to narrow down.
                </p>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>

        <Popover v-model:open="stateTypesOpen">
          <PopoverTrigger as-child>
            <Button variant="outline" size="sm" class="max-w-48 justify-between gap-2 bg-background/90 shadow-sm backdrop-blur">
              <span class="truncate">{{ stateTypesLabel }}</span>
              <ChevronsUpDown class="size-3.5 shrink-0 opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent align="start" class="w-64 p-0">
            <Command>
              <CommandInput placeholder="Search state types..." />
              <CommandList>
                <CommandEmpty>No state type found.</CommandEmpty>
                <CommandGroup>
                  <CommandItem
                    v-if="selectedStateTypes.length"
                    value="clear-state-types"
                    class="text-muted-foreground"
                    @select="updateQuery({ stateTypes: [] })"
                  >
                    Clear selection
                  </CommandItem>
                  <CommandItem
                    v-for="st in stateTypes"
                    :key="st.id"
                    :value="st.label"
                    @select="updateQuery({ stateTypes: toggle(selectedStateTypes, st.id) })"
                  >
                    <Check :class="['size-4', selectedStateTypes.includes(st.id) ? 'opacity-100' : 'opacity-0']" />
                    <span class="size-2 shrink-0 rounded-full" :style="{ backgroundColor: STATE_TYPE_COLORS[st.id] }" />
                    <span class="truncate">{{ st.label }}</span>
                  </CommandItem>
                </CommandGroup>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>

        <Popover v-model:open="usersOpen">
          <PopoverTrigger as-child>
            <Button variant="outline" size="sm" class="max-w-48 justify-between gap-2 bg-background/90 shadow-sm backdrop-blur">
              <span class="truncate">{{ usersLabel }}</span>
              <ChevronsUpDown class="size-3.5 shrink-0 opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent align="start" class="w-72 p-0">
            <Command>
              <CommandInput placeholder="Search users..." @input="userQuery = ($event.target as HTMLInputElement).value" />
              <CommandList>
                <CommandEmpty>No user found.</CommandEmpty>
                <CommandGroup>
                  <CommandItem
                    v-if="selectedUsers.length"
                    value="clear-users"
                    class="text-muted-foreground"
                    @select="updateQuery({ users: [] })"
                  >
                    Clear selection
                  </CommandItem>
                  <CommandItem
                    v-for="u in userMatches.items"
                    :key="u.username"
                    :value="`${u.first_name} ${u.last_name} ${u.username} ${u.email}`"
                    @select="toggleUser(u.username)"
                  >
                    <Check :class="['size-4 shrink-0', selectedUsers.includes(u.username) ? 'opacity-100' : 'opacity-0']" />
                    <span class="truncate">{{ u.first_name }} {{ u.last_name }}</span>
                    <span class="ml-auto truncate text-xs text-muted-foreground">@{{ u.username }}</span>
                  </CommandItem>
                </CommandGroup>
                <p v-if="userMatches.total > userMatches.items.length" class="px-3 py-2 text-xs text-muted-foreground">
                  Showing {{ userMatches.items.length }} of {{ userMatches.total }}. Type to narrow down.
                </p>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>

        <Button size="sm" class="gap-1.5 shadow-sm" :disabled="!canRun" @click="runGraph">
          <Loader2 v-if="loading" class="size-3.5 animate-spin" />
          <Play v-else class="size-3.5" />
          Run
        </Button>

        <Popover v-model:open="userSearchOpen">
          <PopoverTrigger as-child>
            <Button variant="outline" size="sm" class="gap-2 bg-background/90 shadow-sm backdrop-blur" :disabled="!graph">
              <Search class="size-3.5 opacity-60" />
              Find user
            </Button>
          </PopoverTrigger>
          <PopoverContent align="start" class="w-72 p-0">
            <Command>
              <CommandInput placeholder="Search by name or email..." />
              <CommandList>
                <CommandEmpty>No user found.</CommandEmpty>
                <CommandGroup>
                  <CommandItem v-for="u in sortedUsers" :key="u.id" :value="`${u.first_name} ${u.last_name} ${u.email}`" @select="focusUser(u)">
                    <span class="truncate">{{ u.first_name }} {{ u.last_name }}</span>
                    <span class="ml-auto truncate text-xs text-muted-foreground">{{ u.email }}</span>
                  </CommandItem>
                </CommandGroup>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>

        <span
          v-if="graph"
          class="ml-auto rounded-md bg-background/90 px-2 py-1 text-xs text-muted-foreground shadow-sm backdrop-blur"
        >
          {{ filtered.users.length }} users · {{ filtered.tasks.length }} tasks
        </span>
      </div>

      <div v-if="!graph" class="flex h-full items-center justify-center text-sm text-muted-foreground">
        <Loader2 v-if="loading" class="size-6 animate-spin" />
        <span v-else>Pick your filters and press Run to load the graph</span>
      </div>

      <div
        v-else-if="flow.nodes.length === 0"
        class="flex h-full items-center justify-center text-sm text-muted-foreground"
      >
        No tasks match these filters
      </div>

      <TooltipProvider v-else>
        <GraphNodeTooltip :node="hovered" :reference="hoveredEl" />
        <VueFlow
          class="absolute inset-0"
          :nodes="flow.nodes"
          :edges="flow.edges"
          :nodes-connectable="false"
          :min-zoom="0.05"
          @node-click="onNodeClick"
          @node-mouse-enter="onNodeMouseEnter"
          @node-mouse-leave="hideTooltip"
          @node-drag-start="hideTooltip"
          @move-start="hideTooltip"
          @nodes-initialized="fitView()"
        >
          <template #node-user="{ data }">
            <GraphUserNode :data="data" />
          </template>
          <template #node-task="{ data }">
            <GraphTaskNode :data="data" />
          </template>
          <Background :gap="20" />
          <Controls :show-interactive="false" />
        </VueFlow>
      </TooltipProvider>
    </main>
  </div>
</template>

<style>
.vue-flow__controls-button {
  background: var(--card);
  border-color: var(--border);
  color: var(--foreground);
  fill: currentColor;
}

.vue-flow__controls-button:hover {
  background: var(--muted);
}
</style>
