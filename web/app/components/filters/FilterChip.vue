<script setup lang="ts">
import { X } from "lucide-vue-next";
import type {
  Predicate,
  ProjectState,
  ProjectMember,
  ProjectLabel,
  FilterValue,
} from "~/types";
import { findFieldDef, findOpDef, stateTypeLabel } from "./filterCatalog";
import FilterPredicateEditor from "./FilterPredicateEditor.vue";

const props = defineProps<{
  predicate: Predicate;
  states: ProjectState[];
  labels: ProjectLabel[];
  members: ProjectMember[];
  currentUserId?: string;
}>();

const emit = defineEmits<{
  update: [predicate: Predicate];
  remove: [];
}>();

const open = ref(false);

const fieldDef = computed(() => findFieldDef(props.predicate.field));
const opDef = computed(() => findOpDef(props.predicate.field, props.predicate.op));

// Resolve a value into a short display string.
const valueLabel = computed<string>(() => {
  const v = props.predicate.value;
  const field = props.predicate.field;
  if (v === undefined || v === null) return "";
  if (typeof v === "string") return v;
  if (typeof v === "number") return formatNumberValue(field, v);
  if (Array.isArray(v)) return formatArray(field, v as (string | number)[]);
  if (typeof v === "object" && "from" in v) {
    return `${v.from} → ${v.to}`;
  }
  return String(v);
});

function formatArray(field: string, items: (string | number)[]): string {
  if (items.length === 0) return "—";
  const names = items.slice(0, 2).map((it) => formatSingle(field, it));
  if (items.length > 2) names.push(`+${items.length - 2}`);
  return names.join(", ");
}

function formatSingle(field: string, item: string | number): string {
  if (typeof item === "number") return formatNumberValue(field, item);
  if (item === "@me") return "Me";
  if (field === "state") {
    const s = props.states.find((x) => x.id === item);
    return s?.name ?? String(item).slice(0, 6);
  }
  if (field === "state_type") return stateTypeLabel(String(item));
  if (field === "labels") {
    const l = props.labels.find((x) => x.id === item);
    return l?.name ?? String(item).slice(0, 6);
  }
  if (field === "assignees" || field === "created_by") {
    const m = props.members.find((x) => x.user_id === item);
    if (!m) return String(item).slice(0, 6);
    return `${m.first_name} ${m.last_name}`.trim() || m.username;
  }
  return String(item);
}

function formatNumberValue(field: string, n: number): string {
  if (field === "priority") {
    const map = ["None", "Low", "Med", "High", "Urgent"];
    return map[n] ?? String(n);
  }
  return String(n);
}

function onConfirm(p: Predicate) {
  open.value = false;
  emit("update", p);
}

function onCancel() {
  open.value = false;
}

function handleValueUpdate(v: FilterValue | undefined) {
  // unused placeholder to keep linter quiet on optional field access
  void v;
}
void handleValueUpdate;
</script>

<template>
  <Popover v-model:open="open">
    <PopoverTrigger as-child>
      <button
        type="button"
        class="group inline-flex items-center gap-1.5 rounded-full border border-border bg-background px-2.5 py-1 text-xs font-medium transition-colors hover:border-primary/50 hover:bg-accent"
      >
        <component v-if="fieldDef" :is="fieldDef.icon" class="size-3 text-muted-foreground" />
        <span class="text-muted-foreground">{{ fieldDef?.label ?? predicate.field }}</span>
        <span class="text-foreground">{{ opDef?.label ?? predicate.op }}</span>
        <span v-if="valueLabel" class="text-foreground">{{ valueLabel }}</span>
        <button
          type="button"
          class="ml-0.5 flex items-center justify-center rounded-full p-0.5 text-muted-foreground opacity-70 transition-all hover:bg-destructive/10 hover:text-destructive group-hover:opacity-100"
          aria-label="Remove filter"
          @click.stop="emit('remove')"
        >
          <X class="size-3" />
        </button>
      </button>
    </PopoverTrigger>
    <PopoverContent align="start" class="p-0">
      <FilterPredicateEditor
        :initial="predicate"
        :states="states"
        :labels="labels"
        :members="members"
        lock-field
        @confirm="onConfirm"
        @cancel="onCancel"
      />
    </PopoverContent>
  </Popover>
</template>
