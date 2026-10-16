<script setup lang="ts">
import { Loader2, RotateCcw, Save, Undo2 } from "lucide-vue-next";
import { toast } from "vue-sonner";
import { useMediaQuery } from "@vueuse/core";
import { VueFlow, type Connection, type Edge, type Node, type NodeDragEvent } from "@vue-flow/core";
import { Background } from "@vue-flow/background";
import { Controls } from "@vue-flow/controls";
import "@vue-flow/core/dist/style.css";
import "@vue-flow/core/dist/theme-default.css";
import "@vue-flow/controls/dist/style.css";
import type {
  ProjectMember,
  ProjectState,
  SaveWorkflowRequest,
  Workflow,
  WorkflowStateConfig,
  WorkflowTransitionConfig,
} from "~/types";

const props = defineProps<{
  projectKey: string;
  states: ProjectState[];
  members: ProjectMember[];
}>();

const { getWorkflow, saveWorkflow, setWorkflowEnabled } = useWorkflow();
const isDesktop = useMediaQuery("(min-width: 768px)");
const editable = computed(() => isDesktop.value);

const STATE_TYPE_ORDER = ["backlog", "unstarted", "started", "completed", "cancelled"];
const COL_GAP = 240;
const ROW_GAP = 100;

const loading = ref(true);
const saving = ref(false);
const toggling = ref(false);
const workflow = ref<Workflow | null>(null);
const draftStates = ref<Record<string, WorkflowStateConfig>>({});
const draftTransitions = ref<WorkflowTransitionConfig[]>([]);
const savedSnapshot = ref("");
const selected = ref<{ kind: "state"; id: string } | { kind: "edge"; from: string; to: string } | null>(null);

const orderedStates = computed(() =>
  [...props.states].sort(
    (a, b) =>
      STATE_TYPE_ORDER.indexOf(a.state_type) - STATE_TYPE_ORDER.indexOf(b.state_type) || a.position - b.position
  )
);

function defaultPositions(): Record<string, { x: number; y: number }> {
  const rows: Record<string, number> = {};
  const out: Record<string, { x: number; y: number }> = {};
  for (const s of orderedStates.value) {
    const col = STATE_TYPE_ORDER.indexOf(s.state_type);
    const row = rows[s.state_type] ?? 0;
    rows[s.state_type] = row + 1;
    out[s.id] = { x: col * COL_GAP, y: row * ROW_GAP };
  }
  return out;
}

function blankState(id: string, pos: { x: number; y: number }, isEntry: boolean): WorkflowStateConfig {
  return { state_id: id, is_entry: isEntry, pos_x: pos.x, pos_y: pos.y, add_assignees: [], remove_assignees: [] };
}

function openTransition(from: string, to: string): WorkflowTransitionConfig {
  return { from_state_id: from, to_state_id: to, allow_assignees: false, allow_creator: false, mover_ids: [] };
}

// Configured workflows keep their graph; new projects start from a two-way chain.
function buildDraft(wf: Workflow) {
  const positions = defaultPositions();
  const states: Record<string, WorkflowStateConfig> = {};
  const configured = new Map(wf.states.map((s) => [s.state_id, s]));
  for (const s of orderedStates.value) {
    const existing = configured.get(s.id);
    states[s.id] = existing
      ? structuredClone(toRaw(existing))
      : blankState(s.id, positions[s.id]!, !wf.configured && s.is_default);
  }
  let transitions: WorkflowTransitionConfig[];
  if (wf.configured) {
    transitions = wf.transitions.filter((t) => states[t.from_state_id] && states[t.to_state_id]).map((t) => structuredClone(toRaw(t)));
  } else {
    transitions = [];
    const chain = orderedStates.value;
    for (let i = 0; i + 1 < chain.length; i++) {
      transitions.push(openTransition(chain[i]!.id, chain[i + 1]!.id), openTransition(chain[i + 1]!.id, chain[i]!.id));
    }
  }
  draftStates.value = states;
  draftTransitions.value = transitions;
  savedSnapshot.value = snapshot();
}

function snapshot() {
  return JSON.stringify(payload());
}

function payload(): SaveWorkflowRequest {
  return { states: Object.values(draftStates.value), transitions: draftTransitions.value };
}

const dirty = computed(() => snapshot() !== savedSnapshot.value);

async function load() {
  loading.value = true;
  const res = await getWorkflow(props.projectKey, true);
  loading.value = false;
  if (!res.success || !res.data) {
    toast.error(res.error || "Failed to load workflow");
    return;
  }
  workflow.value = res.data;
  buildDraft(res.data);
}

onMounted(load);
watch(
  () => props.states.map((s) => s.id).join(),
  () => workflow.value && buildDraft(workflow.value)
);

const stateById = computed(() => new Map(props.states.map((s) => [s.id, s])));

const warnings = computed(() => {
  const out: Record<string, string | undefined> = {};
  const exits = new Set(draftTransitions.value.map((t) => t.from_state_id));
  const entries = new Set(draftTransitions.value.map((t) => t.to_state_id));
  for (const s of props.states) {
    const cfg = draftStates.value[s.id];
    if (!cfg) continue;
    const terminal = s.state_type === "completed" || s.state_type === "cancelled";
    if (!cfg.is_entry && !entries.has(s.id)) out[s.id] = "Unreachable: no transition leads here";
    else if (!exits.has(s.id) && !terminal) out[s.id] = "No exits: tasks here can only be moved by admins";
  }
  return out;
});

const nodes = computed<Node[]>(() =>
  Object.values(draftStates.value)
    .filter((cfg) => stateById.value.has(cfg.state_id))
    .map((cfg) => ({
      id: cfg.state_id,
      type: "state",
      position: { x: cfg.pos_x, y: cfg.pos_y },
      draggable: editable.value,
      selectable: true,
      data: {
        state: stateById.value.get(cfg.state_id),
        config: cfg,
        warning: warnings.value[cfg.state_id],
        selected: selected.value?.kind === "state" && selected.value.id === cfg.state_id,
      },
    }))
);

const edges = computed<Edge[]>(() =>
  draftTransitions.value.map((t) => {
    const isSelected =
      selected.value?.kind === "edge" && selected.value.from === t.from_state_id && selected.value.to === t.to_state_id;
    return {
      id: `${t.from_state_id}->${t.to_state_id}`,
      source: t.from_state_id,
      target: t.to_state_id,
      type: "transition",
      data: {
        restricted: t.mover_ids.length > 0 || t.allow_assignees || t.allow_creator,
        selected: isSelected,
      },
    };
  })
);

function onConnect(c: Connection) {
  if (!editable.value || !c.source || !c.target || c.source === c.target) return;
  if (draftTransitions.value.some((t) => t.from_state_id === c.source && t.to_state_id === c.target)) return;
  draftTransitions.value.push(openTransition(c.source, c.target));
  selected.value = { kind: "edge", from: c.source, to: c.target };
}

function onNodeDragStop(e: NodeDragEvent) {
  for (const n of e.nodes) {
    const cfg = draftStates.value[n.id];
    if (cfg) {
      cfg.pos_x = Math.round(n.position.x);
      cfg.pos_y = Math.round(n.position.y);
    }
  }
}

function onEdgeClick({ edge }: { edge: Edge }) {
  selected.value = { kind: "edge", from: edge.source, to: edge.target };
}

const selectedState = computed(() =>
  selected.value?.kind === "state" ? draftStates.value[selected.value.id] : undefined
);
const selectedTransition = computed(() => {
  const sel = selected.value;
  if (sel?.kind !== "edge") return undefined;
  return draftTransitions.value.find((t) => t.from_state_id === sel.from && t.to_state_id === sel.to);
});

function removeSelectedTransition() {
  const t = selectedTransition.value;
  if (!t || !editable.value) return;
  draftTransitions.value = draftTransitions.value.filter((x) => x !== t);
  selected.value = null;
}

function onKeydown(e: KeyboardEvent) {
  const target = e.target as HTMLElement | null;
  if (target?.closest("input, textarea, [role='combobox'], [role='dialog']")) return;
  if (e.key === "Delete" || e.key === "Backspace") removeSelectedTransition();
}

function resetLayout() {
  const positions = defaultPositions();
  for (const cfg of Object.values(draftStates.value)) {
    const p = positions[cfg.state_id];
    if (p) {
      cfg.pos_x = p.x;
      cfg.pos_y = p.y;
    }
  }
}

function discard() {
  if (workflow.value) buildDraft(workflow.value);
  selected.value = null;
}

async function save() {
  saving.value = true;
  const res = await saveWorkflow(props.projectKey, payload());
  saving.value = false;
  if (!res.success || !res.data) {
    toast.error(res.error || "Failed to save workflow");
    return;
  }
  workflow.value = res.data;
  savedSnapshot.value = snapshot();
  toast.success("Workflow saved");
}

async function toggleEnabled(enabled: boolean) {
  toggling.value = true;
  const res = await setWorkflowEnabled(props.projectKey, enabled);
  toggling.value = false;
  if (!res.success || !res.data) {
    toast.error(res.error || "Failed to update workflow");
    return;
  }
  workflow.value = { ...workflow.value!, enabled: res.data.enabled };
  toast.success(enabled ? "Workflow rules are now enforced" : "Workflow rules turned off");
}

function onBeforeUnload(e: BeforeUnloadEvent) {
  if (dirty.value) e.preventDefault();
}
onMounted(() => window.addEventListener("beforeunload", onBeforeUnload));
onBeforeUnmount(() => window.removeEventListener("beforeunload", onBeforeUnload));
const confirmLeave = () => !dirty.value || window.confirm("Discard unsaved workflow changes?");
onBeforeRouteLeave(confirmLeave);
onBeforeRouteUpdate((to) => to.query.tab === "workflow" || confirmLeave());
</script>

<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="space-y-1">
        <h2 class="text-lg font-semibold">Workflow</h2>
        <p class="max-w-xl text-sm text-muted-foreground">
          When on, tasks can only move along the arrows below, and each arrow can be limited to certain people. States can
          also assign or unassign people when a task enters them. Project admins can always make any move.
        </p>
      </div>
      <div v-if="editable && workflow" class="flex items-center gap-2">
        <Label for="workflow-enabled" class="text-sm">Enforce</Label>
        <Switch
          id="workflow-enabled"
          :checked="workflow.enabled"
          :disabled="toggling || !workflow.configured || dirty"
          @update:checked="toggleEnabled"
        />
      </div>
    </div>

    <div v-if="loading" class="flex h-64 items-center justify-center">
      <Loader2 class="size-6 animate-spin text-muted-foreground" />
    </div>

    <template v-else-if="workflow">
      <div v-if="editable" class="flex flex-wrap items-center gap-2">
        <Button size="sm" class="gap-1.5" :disabled="saving || (!dirty && workflow.configured)" @click="save">
          <Loader2 v-if="saving" class="size-3.5 animate-spin" />
          <Save v-else class="size-3.5" />
          Save
        </Button>
        <Button size="sm" variant="outline" class="gap-1.5" :disabled="saving || !dirty || !workflow.configured" @click="discard">
          <Undo2 class="size-3.5" />
          Discard
        </Button>
        <Button size="sm" variant="ghost" class="gap-1.5" @click="resetLayout">
          <RotateCcw class="size-3.5" />
          Reset layout
        </Button>
        <span v-if="!workflow.configured" class="text-xs text-muted-foreground">
          Suggested starting point. Save it before you can turn enforcement on.
        </span>
        <span v-else-if="dirty" class="text-xs text-muted-foreground">Unsaved changes</span>
        <span v-else-if="!workflow.enabled" class="text-xs text-muted-foreground">Saved, not enforced</span>
      </div>
      <p v-else class="text-sm text-muted-foreground">Open this page on a larger screen to edit the workflow.</p>

      <div class="flex flex-col gap-4 lg:flex-row">
        <div
          class="relative h-[60vh] min-h-96 flex-1 overflow-hidden rounded-lg border bg-muted/20 outline-none"
          tabindex="0"
          @keydown="onKeydown"
        >
          <VueFlow
            class="absolute inset-0"
            :nodes="nodes"
            :edges="edges"
            :nodes-connectable="editable"
            :edges-updatable="false"
            :delete-key-code="null"
            :connection-radius="80"
            :min-zoom="0.2"
            :zoom-on-double-click="false"
            fit-view-on-init
            @connect="onConnect"
            @node-drag-stop="onNodeDragStop"
            @node-click="({ node }) => (selected = { kind: 'state', id: node.id })"
            @edge-click="onEdgeClick"
            @pane-click="selected = null"
          >
            <template #node-state="{ data }">
              <WorkflowStateNode :data="data" :connectable="editable" />
            </template>
            <template #edge-transition="edgeProps">
              <WorkflowTransitionEdge v-bind="edgeProps" />
            </template>
            <Background :gap="20" />
            <Controls :show-interactive="false" />
          </VueFlow>
          <p
            v-if="editable"
            class="pointer-events-none absolute bottom-2 right-3 text-[11px] text-muted-foreground"
          >
            Drag from a state's dot to another state to add a transition
          </p>
        </div>

        <aside class="w-full shrink-0 rounded-lg border p-4 lg:w-80">
          <WorkflowStatePanel
            v-if="selectedState"
            :config="selectedState"
            :state="stateById.get(selectedState.state_id)"
            :members="members"
            :warning="warnings[selectedState.state_id]"
            :editable="editable"
          />
          <WorkflowTransitionPanel
            v-else-if="selectedTransition"
            :transition="selectedTransition"
            :from="stateById.get(selectedTransition.from_state_id)"
            :to="stateById.get(selectedTransition.to_state_id)"
            :members="members"
            :editable="editable"
            @remove="removeSelectedTransition"
          />
          <p v-else class="text-sm text-muted-foreground">
            Select a state to choose who gets assigned or unassigned when a task enters it, or select an arrow to choose
            who can make that move.
          </p>
        </aside>
      </div>
    </template>
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
