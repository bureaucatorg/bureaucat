<script setup lang="ts">
import { ChevronLeft, Check } from "lucide-vue-next";
import type { FilterField, FilterOp, Predicate, ProjectState, ProjectMember, ProjectLabel, FilterValue } from "~/types";
import { FILTER_CATALOG, findFieldDef, findOpDef } from "./filterCatalog";
import FilterValuePicker from "./FilterValuePicker.vue";

const props = withDefaults(
  defineProps<{
    /** Predicate to edit; undefined for a fresh predicate. */
    initial?: Predicate;
    states: ProjectState[];
    labels: ProjectLabel[];
    members: ProjectMember[];
    /** Lock the field so only op/value can change. */
    lockField?: boolean;
  }>(),
  { initial: undefined, lockField: false }
);

const emit = defineEmits<{
  confirm: [predicate: Predicate];
  cancel: [];
}>();

type Step = "field" | "op" | "value";

const step = ref<Step>(props.initial ? "value" : "field");
const field = ref<FilterField | null>(props.initial?.field ?? null);
const op = ref<FilterOp | null>(props.initial?.op ?? null);
const value = ref<FilterValue | undefined>(props.initial?.value);

function chooseField(f: FilterField) {
  field.value = f;
  const def = findFieldDef(f);
  if (def && def.ops.length === 1) {
    op.value = def.ops[0].op;
    value.value = undefined;
    step.value = findOpDef(f, op.value)?.valueKind === "none" ? "value" : "value";
  } else {
    op.value = null;
    value.value = undefined;
    step.value = "op";
  }
}

function chooseOp(o: FilterOp) {
  op.value = o;
  const opDef = findOpDef(field.value!, o);
  if (opDef?.valueKind === "none") {
    value.value = undefined;
    confirm();
    return;
  }
  value.value = undefined;
  step.value = "value";
}

function back() {
  if (step.value === "value") {
    if (props.lockField) {
      // Go back to op if field is locked.
      step.value = "op";
    } else {
      step.value = "op";
    }
  } else if (step.value === "op") {
    if (props.lockField) return; // nowhere to go back to
    step.value = "field";
  }
}

function confirm() {
  if (!field.value || !op.value) return;
  const opDef = findOpDef(field.value, op.value);
  if (!opDef) return;
  if (opDef.valueKind !== "none" && value.value === undefined) return;
  emit("confirm", { field: field.value, op: op.value, value: value.value });
}

const currentFieldDef = computed(() => (field.value ? findFieldDef(field.value) : null));
const currentOpDef = computed(() =>
  field.value && op.value ? findOpDef(field.value, op.value) : null
);

const isDateValue = computed(() => {
  if (!currentOpDef.value) return false;
  return currentOpDef.value.valueKind === "date" || currentOpDef.value.valueKind === "date-range";
});
const canConfirm = computed(() => {
  if (!field.value || !op.value) return false;
  const def = findOpDef(field.value, op.value);
  if (!def) return false;
  if (def.valueKind === "none") return true;
  if (value.value === undefined) return false;
  if (Array.isArray(value.value) && value.value.length === 0) return false;
  if (typeof value.value === "string" && value.value === "") return false;
  return true;
});
</script>

<template>
  <div class="overflow-hidden" :class="step === 'value' && isDateValue ? 'w-80' : 'w-72'">
    <!-- step header with back button when applicable -->
    <header class="flex items-center justify-between gap-2 border-b bg-muted/30 px-3 py-2">
      <button
        v-if="step !== 'field' && !(lockField && step === 'op')"
        type="button"
        class="flex items-center gap-1 text-xs text-muted-foreground transition-colors hover:text-foreground"
        @click="back"
      >
        <ChevronLeft class="size-3.5" />
        Back
      </button>
      <span class="ml-auto truncate text-xs font-medium text-muted-foreground">
        <template v-if="step === 'field'">Filter by&hellip;</template>
        <template v-else-if="step === 'op' && currentFieldDef">
          {{ currentFieldDef.label }}
        </template>
        <template v-else-if="step === 'value' && currentFieldDef && currentOpDef">
          {{ currentFieldDef.label }} <span class="text-foreground">{{ currentOpDef.label }}</span>
        </template>
      </span>
    </header>

    <!-- STEP 1: field -->
    <div v-if="step === 'field'" class="max-h-72 overflow-y-auto p-1.5">
      <button
        v-for="def in FILTER_CATALOG"
        :key="def.field"
        type="button"
        class="flex w-full items-center gap-2.5 rounded-md px-2.5 py-2 text-sm transition-colors hover:bg-accent hover:text-accent-foreground"
        @click="chooseField(def.field)"
      >
        <div class="flex size-6 items-center justify-center rounded bg-muted/60">
          <component :is="def.icon" class="size-3.5 text-muted-foreground" />
        </div>
        {{ def.label }}
      </button>
    </div>

    <!-- STEP 2: op -->
    <div v-else-if="step === 'op' && currentFieldDef" class="max-h-72 overflow-y-auto p-1.5">
      <button
        v-for="o in currentFieldDef.ops"
        :key="o.op"
        type="button"
        class="block w-full rounded-md px-2.5 py-2 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground"
        @click="chooseOp(o.op)"
      >
        {{ o.label }}
      </button>
    </div>

    <!-- STEP 3: value -->
    <div v-else-if="step === 'value' && field && op">
      <FilterValuePicker
        :field="field"
        :op="op"
        :value="value"
        :states="states"
        :labels="labels"
        :members="members"
        @update:value="(v) => (value = v)"
      />
      <div class="flex items-center justify-end gap-2 border-t px-3 py-2.5">
        <Button type="button" variant="ghost" size="sm" @click="emit('cancel')">Cancel</Button>
        <Button type="button" size="sm" :disabled="!canConfirm" @click="confirm">
          <Check class="mr-1 size-3.5" />
          Apply
        </Button>
      </div>
    </div>
  </div>
</template>
