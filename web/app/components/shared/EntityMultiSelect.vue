<script setup lang="ts">
import { Search } from "lucide-vue-next";

interface EntityItem {
  [key: string]: unknown;
}

const props = withDefaults(
  defineProps<{
    items: EntityItem[];
    modelValue: string[];
    itemKey?: string;
    placeholder?: string;
    emptyMessage?: string;
    /** When false, selecting an item replaces the selection. */
    multi?: boolean;
    /** Optional predicate for filtering items by search string. */
    filter?: (item: EntityItem, query: string) => boolean;
    /** Limit list height; defaults to a sensible medium. */
    maxHeightClass?: string;
  }>(),
  {
    itemKey: "id",
    placeholder: "Search…",
    emptyMessage: "No matches",
    multi: true,
    filter: undefined,
    maxHeightClass: "max-h-60",
  }
);

const emit = defineEmits<{
  "update:modelValue": [ids: string[]];
  select: [id: string];
}>();

const search = ref("");
const highlighted = ref(-1);

const filtered = computed(() => {
  const q = search.value.trim().toLowerCase();
  if (!q) return props.items;
  if (props.filter) return props.items.filter((i) => props.filter!(i, q));
  return props.items.filter((item) =>
    Object.values(item).some(
      (v) => typeof v === "string" && v.toLowerCase().includes(q)
    )
  );
});

function idOf(item: EntityItem): string {
  return String(item[props.itemKey] ?? "");
}

function isSelected(item: EntityItem): boolean {
  return props.modelValue.includes(idOf(item));
}

function toggle(item: EntityItem) {
  const id = idOf(item);
  emit("select", id);
  if (!props.multi) {
    emit("update:modelValue", isSelected(item) ? [] : [id]);
    return;
  }
  if (isSelected(item)) {
    emit(
      "update:modelValue",
      props.modelValue.filter((x) => x !== id)
    );
  } else {
    emit("update:modelValue", [...props.modelValue, id]);
  }
}

function onKeydown(event: KeyboardEvent) {
  const items = filtered.value;
  if (event.key === "ArrowDown") {
    event.preventDefault();
    highlighted.value = Math.min(highlighted.value + 1, items.length - 1);
  } else if (event.key === "ArrowUp") {
    event.preventDefault();
    highlighted.value = Math.max(highlighted.value - 1, 0);
  } else if (event.key === "Enter") {
    event.preventDefault();
    const item = items[highlighted.value];
    if (item) toggle(item);
  }
}

watch(search, () => {
  highlighted.value = -1;
});
</script>

<template>
  <div class="w-full">
    <div class="p-2">
      <div class="relative">
        <Search
          class="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground"
        />
        <Input
          v-model="search"
          :placeholder="placeholder"
          class="h-8 pl-8 text-sm"
          @keydown="onKeydown"
        />
      </div>
    </div>
    <Separator />
    <div :class="['overflow-y-auto p-1', maxHeightClass]">
      <button
        v-for="(item, idx) in filtered"
        :key="idOf(item) || idx"
        type="button"
        class="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm transition-colors"
        :class="[
          highlighted === idx ? 'bg-accent text-accent-foreground' : 'hover:bg-accent hover:text-accent-foreground',
          isSelected(item) ? 'font-medium' : ''
        ]"
        @click="toggle(item)"
        @mouseenter="highlighted = idx"
      >
        <span
          v-if="multi"
          class="flex size-4 shrink-0 items-center justify-center rounded border"
          :class="isSelected(item) ? 'border-primary bg-primary text-primary-foreground' : 'border-muted-foreground/50'"
        >
          <svg
            v-if="isSelected(item)"
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="3"
            class="size-3"
          >
            <polyline points="20 6 9 17 4 12" />
          </svg>
        </span>
        <slot name="option" :item="item" :selected="isSelected(item)">
          <span>{{ idOf(item) }}</span>
        </slot>
      </button>
      <p
        v-if="filtered.length === 0"
        class="px-2 py-3 text-center text-xs text-muted-foreground"
      >
        {{ emptyMessage }}
      </p>
    </div>
  </div>
</template>
