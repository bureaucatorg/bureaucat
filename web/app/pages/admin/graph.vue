<script setup lang="ts">
import { Check, ChevronsUpDown, Loader2, Search } from "lucide-vue-next";
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
import type { GraphTask, GraphUser, TaskGraph } from "~/composables/useAdmin";
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
const { getTaskGraph } = useAdmin();
const { fitView, findNode, setCenter } = useVueFlow();

const loading = ref(true);
const graph = ref<TaskGraph | null>(null);
const workspaceOpen = ref(false);
const projectsOpen = ref(false);
const stateTypesOpen = ref(false);
const userSearchOpen = ref(false);

onMounted(async () => {
  const result = await getTaskGraph();
  if (result.success && result.data) {
    graph.value = result.data;
  } else {
    toast.error(result.error || "Failed to load graph");
  }
  loading.value = false;
});

const selectedWorkspace = computed(() =>
  typeof route.query.workspace === "string" ? route.query.workspace : "",
);
function listParam(name: string): string[] {
  const value = route.query[name];
  return typeof value === "string" && value ? value.split(",") : [];
}

const selectedProjects = computed(() => listParam("projects"));
const selectedStateTypes = computed(() => listParam("state_types"));

function updateQuery(patch: { workspace?: string; projects?: string[]; stateTypes?: string[] }) {
  const query: LocationQueryRaw = { ...route.query };
  if (patch.workspace !== undefined) query.workspace = patch.workspace || undefined;
  if (patch.projects) query.projects = patch.projects.length ? patch.projects.join(",") : undefined;
  // An empty value means "all"; a missing one triggers the default redirect.
  if (patch.stateTypes) query.state_types = patch.stateTypes.join(",");
  router.replace({ query });
}

function toggle(list: string[], value: string): string[] {
  return list.includes(value) ? list.filter((v) => v !== value) : [...list, value];
}

function selectWorkspace(key: string) {
  workspaceOpen.value = false;
  if (key !== selectedWorkspace.value) updateQuery({ workspace: key, projects: [] });
}

const workspaces = computed(() => {
  const map = new Map<string, string>();
  for (const t of graph.value?.tasks ?? []) map.set(t.workspace_key, t.workspace_name);
  return [...map].map(([key, name]) => ({ key, name })).sort((a, b) => a.name.localeCompare(b.name));
});

const projects = computed(() => {
  const map = new Map<string, string>();
  for (const t of graph.value?.tasks ?? []) {
    if (!selectedWorkspace.value || t.workspace_key === selectedWorkspace.value) map.set(t.project_key, t.project_name);
  }
  return [...map].map(([key, name]) => ({ key, name })).sort((a, b) => a.key.localeCompare(b.key));
});

const scopedTasks = computed(() => {
  const projectSet = new Set(selectedProjects.value);
  return (graph.value?.tasks ?? []).filter(
    (t) =>
      (!selectedWorkspace.value || t.workspace_key === selectedWorkspace.value) &&
      (projectSet.size === 0 || projectSet.has(t.project_key)),
  );
});

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

const filtered = computed(() => {
  const typeSet = new Set(selectedStateTypes.value);
  const tasks = scopedTasks.value.filter((t) => typeSet.size === 0 || typeSet.has(t.state_type));
  const taskIds = new Set(tasks.map((t) => t.id));
  const edges = (graph.value?.edges ?? []).filter((e) => taskIds.has(e.task_id));
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

function onNodeClick({ node }: NodeMouseEvent) {
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
      <div v-if="loading" class="flex h-full items-center justify-center">
        <Loader2 class="size-6 animate-spin text-muted-foreground" />
      </div>

      <template v-else>
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
                <CommandInput placeholder="Search workspaces..." />
                <CommandList>
                  <CommandEmpty>No workspace found.</CommandEmpty>
                  <CommandGroup>
                    <CommandItem value="all-workspaces" @select="selectWorkspace('')">
                      <Check :class="['size-4', selectedWorkspace ? 'opacity-0' : 'opacity-100']" />
                      All workspaces
                    </CommandItem>
                    <CommandItem v-for="w in workspaces" :key="w.key" :value="`${w.name} ${w.key}`" @select="selectWorkspace(w.key)">
                      <Check :class="['size-4', selectedWorkspace === w.key ? 'opacity-100' : 'opacity-0']" />
                      <span class="truncate">{{ w.name }}</span>
                    </CommandItem>
                  </CommandGroup>
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
                <CommandInput placeholder="Search projects..." />
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
                    <CommandItem v-for="p in projects" :key="p.key" :value="`${p.key} ${p.name}`" @select="updateQuery({ projects: toggle(selectedProjects, p.key) })">
                      <Check :class="['size-4', selectedProjects.includes(p.key) ? 'opacity-100' : 'opacity-0']" />
                      <span class="font-mono text-xs">{{ p.key }}</span>
                      <span class="truncate text-muted-foreground">{{ p.name }}</span>
                    </CommandItem>
                  </CommandGroup>
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

          <Popover v-model:open="userSearchOpen">
            <PopoverTrigger as-child>
              <Button variant="outline" size="sm" class="gap-2 bg-background/90 shadow-sm backdrop-blur">
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

          <span class="ml-auto rounded-md bg-background/90 px-2 py-1 text-xs text-muted-foreground shadow-sm backdrop-blur">
            {{ filtered.users.length }} users · {{ filtered.tasks.length }} tasks
          </span>
        </div>

        <div
          v-if="flow.nodes.length === 0"
          class="flex h-full items-center justify-center text-sm text-muted-foreground"
        >
          No tasks match these filters
        </div>

        <TooltipProvider v-else :delay-duration="150">
          <VueFlow
            class="absolute inset-0"
            :nodes="flow.nodes"
            :edges="flow.edges"
            :nodes-connectable="false"
            :min-zoom="0.05"
            @node-click="onNodeClick"
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
      </template>
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
