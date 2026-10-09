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
import type { GraphTask, GraphUser, HoveredGraphNode, TaskGraph, TaskGraphFilterOptions } from "~/composables/useTaskGraph";
import { STATE_TYPE_OPTIONS, stateTypeLabel } from "~/components/filters/filterCatalog";
import { STATE_TYPE_COLORS } from "~/types/task";

const props = defineProps<{ apiBase: string }>();

const route = useRoute();
const router = useRouter();
const { getTaskGraph, getTaskGraphFilters } = useTaskGraph(props.apiBase);
const { fitView, findNode, findEdge, setCenter } = useVueFlow();

const loading = ref(false);
const graph = shallowRef<TaskGraph | null>(null);
const options = shallowRef<TaskGraphFilterOptions | null>(null);
const appliedKey = ref<string | null>(null);
const workspaceOpen = ref(false);
const projectsOpen = ref(false);
const stateTypesOpen = ref(false);
const usersOpen = ref(false);
const userSearchOpen = ref(false);
const taskSearchOpen = ref(false);
const hovered = ref<HoveredGraphNode | null>(null);
const hoveredEl = ref<HTMLElement | null>(null);
let hoverTimer: ReturnType<typeof setTimeout> | undefined;
let requestId = 0;
const highlightedNode = ref<string | null>(null);
let highlightedIds = { nodes: [] as string[], edges: [] as string[] };

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
    highlightedNode.value = null;
    highlightedIds = { nodes: [], edges: [] };
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
  if (patch.users) query.users = patch.users.join(",");
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

const taskSearchQuery = ref("");
watch(taskSearchOpen, () => (taskSearchQuery.value = ""));

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

const taskSearchIndex = computed(() =>
  filtered.value.tasks.map((t) => ({
    item: t,
    text: `${t.project_key}-${t.task_number} ${t.title}`.toLowerCase(),
  })),
);
const taskSearchMatches = computed(() => capOptions(taskSearchIndex.value, taskSearchQuery.value, () => false));

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
  const taskIds = new Set(tasks.map((t) => t.id));
  const subtaskEdges = tasks
    .filter((t) => t.parent_id && taskIds.has(t.parent_id))
    .map((t) => ({ parent_id: t.parent_id!, child_id: t.id }));
  const blockerEdges = (graph.value?.blocker_edges ?? []).filter(
    (e) => taskIds.has(e.blocker_id) && taskIds.has(e.blocked_id),
  );
  const assignedIds = new Set(edges.map((e) => e.task_id));
  const unassignedCount = tasks.filter((t) => !assignedIds.has(t.id)).length;
  return { users, tasks, edges, subtaskEdges, blockerEdges, unassignedCount };
});

const flow = computed(() => {
  const { users, tasks, edges, subtaskEdges, blockerEdges } = filtered.value;

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
  for (const t of tasks) {
    if (!placed.has(t.id)) order.push(`t:${t.id}`);
  }

  const positions = forceLayout(order, [
    ...edges.map((e): [string, string] => [`u:${e.user_id}`, `t:${e.task_id}`]),
    ...subtaskEdges.map((e): [string, string] => [`t:${e.parent_id}`, `t:${e.child_id}`]),
    ...blockerEdges.map((e): [string, string] => [`t:${e.blocker_id}`, `t:${e.blocked_id}`]),
  ]);

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
  for (const e of subtaskEdges) {
    flowEdges.push({
      id: `s:${e.parent_id}:${e.child_id}`,
      source: `t:${e.parent_id}`,
      target: `t:${e.child_id}`,
      type: "directed",
      class: "subtask",
      data: { color: "var(--graph-subtask)" },
      style: { stroke: "var(--graph-subtask)", strokeOpacity: 0.8, strokeWidth: 1.5 },
    });
  }
  for (const e of blockerEdges) {
    flowEdges.push({
      id: `b:${e.blocker_id}:${e.blocked_id}`,
      source: `t:${e.blocker_id}`,
      target: `t:${e.blocked_id}`,
      type: "directed",
      class: "blocker",
      data: { color: "var(--destructive)" },
      style: { stroke: "var(--destructive)", strokeOpacity: 0.8, strokeWidth: 1.5, strokeDasharray: "6 4" },
    });
  }

  return { nodes, edges: flowEdges };
});

const sortedUsers = computed(() =>
  [...filtered.value.users].sort((a, b) =>
    `${a.first_name} ${a.last_name}`.localeCompare(`${b.first_name} ${b.last_name}`),
  ),
);

function edgeClass(id: string) {
  if (id.startsWith("s:")) return "subtask";
  if (id.startsWith("b:")) return "blocker";
  return undefined;
}

// Only the affected nodes/edges get a class; CSS dims everything else.
function setHighlight(nodeId: string | null) {
  for (const id of highlightedIds.nodes) {
    const n = findNode(id);
    if (n) n.class = undefined;
  }
  for (const id of highlightedIds.edges) {
    const e = findEdge(id);
    if (e) e.class = edgeClass(id);
  }
  highlightedIds = { nodes: [], edges: [] };
  highlightedNode.value = nodeId;
  if (!nodeId) return;

  const isUser = nodeId.startsWith("u:");
  const id = nodeId.slice(2);
  highlightedIds.nodes.push(nodeId);
  for (const e of filtered.value.edges) {
    if ((isUser ? e.user_id : e.task_id) !== id) continue;
    highlightedIds.nodes.push(isUser ? `t:${e.task_id}` : `u:${e.user_id}`);
    highlightedIds.edges.push(`${e.user_id}:${e.task_id}`);
  }
  if (!isUser) {
    for (const e of filtered.value.subtaskEdges) {
      if (e.parent_id !== id && e.child_id !== id) continue;
      highlightedIds.nodes.push(`t:${e.parent_id === id ? e.child_id : e.parent_id}`);
      highlightedIds.edges.push(`s:${e.parent_id}:${e.child_id}`);
    }
    for (const e of filtered.value.blockerEdges) {
      if (e.blocker_id !== id && e.blocked_id !== id) continue;
      highlightedIds.nodes.push(`t:${e.blocker_id === id ? e.blocked_id : e.blocker_id}`);
      highlightedIds.edges.push(`b:${e.blocker_id}:${e.blocked_id}`);
    }
  }
  for (const id of highlightedIds.nodes) {
    const n = findNode(id);
    if (n) n.class = id === nodeId ? "highlighted highlight-source" : "highlighted";
  }
  for (const id of highlightedIds.edges) {
    const e = findEdge(id);
    if (e) e.class = [edgeClass(id), "highlighted"].filter(Boolean).join(" ");
  }
}

function focusUser(user: GraphUser) {
  userSearchOpen.value = false;
  focusNode(`u:${user.id}`);
}

function focusTask(task: GraphTask) {
  taskSearchOpen.value = false;
  focusNode(`t:${task.id}`);
}

function focusNode(nodeId: string) {
  const node = findNode(nodeId);
  if (!node) return;
  setHighlight(nodeId);
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
  setHighlight(highlightedNode.value === node.id ? null : node.id);
}

function onNodeDoubleClick({ node }: NodeMouseEvent) {
  const path =
    node.type === "task"
      ? `/projects/${(node.data as GraphTask).project_key}/tasks/${(node.data as GraphTask).task_number}`
      : `/profile/${(node.data as GraphUser).id}`;
  navigateTo(path, { open: { target: "_blank" } });
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

        <Popover v-model:open="taskSearchOpen">
          <PopoverTrigger as-child>
            <Button variant="outline" size="sm" class="gap-2 bg-background/90 shadow-sm backdrop-blur" :disabled="!graph">
              <Search class="size-3.5 opacity-60" />
              Find task
            </Button>
          </PopoverTrigger>
          <PopoverContent align="start" class="w-80 p-0">
            <Command>
              <CommandInput
                placeholder="Search by ID or title..."
                @input="taskSearchQuery = ($event.target as HTMLInputElement).value"
              />
              <CommandList>
                <CommandEmpty>No task found.</CommandEmpty>
                <CommandGroup>
                  <CommandItem
                    v-for="t in taskSearchMatches.items"
                    :key="t.id"
                    :value="`${t.project_key}-${t.task_number} ${t.title}`"
                    @select="focusTask(t)"
                  >
                    <span class="shrink-0 font-mono text-xs text-muted-foreground">{{ t.project_key }}-{{ t.task_number }}</span>
                    <span class="truncate">{{ t.title }}</span>
                  </CommandItem>
                </CommandGroup>
                <p v-if="taskSearchMatches.total > taskSearchMatches.items.length" class="px-3 py-2 text-xs text-muted-foreground">
                  Showing {{ taskSearchMatches.items.length }} of {{ taskSearchMatches.total }}. Type to narrow down.
                </p>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>

        <div v-if="graph" class="ml-auto flex flex-col items-end gap-2">
          <span class="rounded-md bg-background/90 px-2 py-1 text-xs text-muted-foreground shadow-sm backdrop-blur">
            {{ filtered.users.length }} users · {{ filtered.tasks.length }} tasks
            <template v-if="filtered.unassignedCount">({{ filtered.unassignedCount }} unassigned)</template>
          </span>
          <div class="space-y-1 rounded-md border bg-background/90 px-3 py-2 text-xs text-muted-foreground shadow-sm backdrop-blur">
            <div class="flex items-center gap-2">
              <svg width="28" height="8" aria-hidden="true">
                <line x1="0" y1="4" x2="28" y2="4" style="stroke: var(--muted-foreground)" stroke-width="2" />
              </svg>
              User → assigned task
            </div>
            <div class="flex items-center gap-2">
              <svg width="28" height="8" aria-hidden="true">
                <line x1="0" y1="4" x2="28" y2="4" style="stroke: var(--graph-subtask)" stroke-width="2" />
                <path d="M10 0 L18 4 L10 8 Z" style="fill: var(--graph-subtask)" />
              </svg>
              Parent → subtask
            </div>
            <div class="flex items-center gap-2">
              <svg width="28" height="8" aria-hidden="true">
                <line x1="0" y1="4" x2="28" y2="4" style="stroke: var(--destructive)" stroke-width="2" stroke-dasharray="4 3" />
                <path d="M10 0 L18 4 L10 8 Z" style="fill: var(--destructive)" />
              </svg>
              Blocker → blocked task
            </div>
          </div>
        </div>
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
          :class="['absolute inset-0', { 'graph-highlight': highlightedNode }]"
          :nodes="flow.nodes"
          :edges="flow.edges"
          :nodes-connectable="false"
          :min-zoom="0.05"
          :zoom-on-double-click="false"
          @node-click="onNodeClick"
          @node-double-click="onNodeDoubleClick"
          @pane-click="setHighlight(null)"
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
          <template #edge-directed="edgeProps">
            <GraphDirectedEdge v-bind="edgeProps" />
          </template>
          <Background :gap="20" />
          <Controls :show-interactive="false" />
        </VueFlow>
      </TooltipProvider>
    </main>
  </div>
</template>

<style>
:root {
  --graph-subtask: #eab308;
}

.vue-flow__controls-button {
  background: var(--card);
  border-color: var(--border);
  color: var(--foreground);
  fill: currentColor;
}

.vue-flow__controls-button:hover {
  background: var(--muted);
}

.vue-flow__node,
.vue-flow__edge {
  transition: opacity 150ms;
}

.graph-highlight .vue-flow__node:not(.highlighted) {
  opacity: 0.2;
}

.graph-highlight .vue-flow__edge:not(.highlighted) {
  opacity: 0.04;
}

.graph-highlight .vue-flow__node.highlighted {
  z-index: 1000 !important;
}

.vue-flow__edge.highlighted path.vue-flow__edge-path {
  stroke: var(--foreground) !important;
  stroke-opacity: 0.35 !important;
  stroke-width: 1.25;
}

.vue-flow__edge.subtask.highlighted path.vue-flow__edge-path {
  stroke: var(--graph-subtask) !important;
  stroke-opacity: 1 !important;
  stroke-width: 2;
}

.vue-flow__edge.blocker.highlighted path.vue-flow__edge-path {
  stroke: var(--destructive) !important;
  stroke-opacity: 1 !important;
  stroke-width: 2;
}

.vue-flow__node-user.highlighted [data-slot="avatar"] {
  box-shadow: 0 0 0 2px var(--background), 0 0 0 3px color-mix(in oklch, var(--foreground) 35%, transparent);
}

.vue-flow__node-task.highlighted > div {
  border-color: color-mix(in oklch, var(--foreground) 30%, transparent);
  box-shadow: 0 2px 8px rgb(0 0 0 / 0.1);
}

.vue-flow__node-user.highlight-source [data-slot="avatar"] {
  box-shadow: 0 0 0 2px var(--background), 0 0 0 4px var(--foreground), 0 6px 16px rgb(0 0 0 / 0.18);
}

.vue-flow__node-task.highlight-source > div {
  border-color: var(--foreground);
  box-shadow: 0 0 0 1px var(--foreground), 0 6px 16px rgb(0 0 0 / 0.18);
}
</style>
