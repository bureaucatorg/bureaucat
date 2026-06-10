<script setup lang="ts" generic="T">
import { Search } from "lucide-vue-next";
import { cn } from "@/lib/utils";

// A searchable popover dropdown with built-in keyboard navigation
// (Up/Down to highlight, Enter to select, Escape to close), search filtering,
// scroll-into-view, and search-reset on close.
//
// Callers supply the trigger via the `#trigger` slot and each row's contents
// via the `#option` slot (receives `{ item, active }`). Selection is emitted
// through `@select`; the caller decides what to do with the chosen item.

const props = withDefaults(
  defineProps<{
    items: T[];
    // Text used to match the search query against each item.
    getSearchText: (item: T) => string;
    // Stable key for each row (defaults to the array index).
    getKey?: (item: T) => string | number;
    placeholder?: string;
    emptyText?: string;
    // When true (default) the popover closes after a selection. Set false for
    // multi-select pickers that should stay open to add several items.
    closeOnSelect?: boolean;
    align?: "start" | "center" | "end";
    contentClass?: string;
  }>(),
  {
    placeholder: "Search...",
    emptyText: "No results found",
    closeOnSelect: true,
    align: "start",
    contentClass: "w-56",
  }
);

const open = defineModel<boolean>("open", { default: false });

const emit = defineEmits<{
  select: [item: T];
}>();

const search = ref("");
const highlighted = ref(0);
const listRef = ref<HTMLElement | null>(null);

const filtered = computed(() => {
  const q = search.value.toLowerCase().trim();
  if (!q) return props.items;
  return props.items.filter((it) =>
    props.getSearchText(it).toLowerCase().includes(q)
  );
});

function keyOf(item: T, index: number): string | number {
  return props.getKey ? props.getKey(item) : index;
}

watch(open, (isOpen) => {
  if (isOpen) {
    highlighted.value = 0;
  } else {
    search.value = "";
  }
});

// Keep the highlight in range as the filtered list shrinks while typing.
watch(filtered, (list) => {
  if (highlighted.value >= list.length) {
    highlighted.value = Math.max(0, list.length - 1);
  }
});

function scrollHighlightedIntoView() {
  nextTick(() => {
    const el = listRef.value?.querySelector<HTMLElement>(
      `[data-index="${highlighted.value}"]`
    );
    el?.scrollIntoView({ block: "nearest" });
  });
}

function choose(item: T) {
  emit("select", item);
  if (props.closeOnSelect) open.value = false;
}

function onKeydown(event: KeyboardEvent) {
  const count = filtered.value.length;
  if (event.key === "ArrowDown") {
    event.preventDefault();
    if (count === 0) return;
    highlighted.value = (highlighted.value + 1) % count;
    scrollHighlightedIntoView();
  } else if (event.key === "ArrowUp") {
    event.preventDefault();
    if (count === 0) return;
    highlighted.value = (highlighted.value - 1 + count) % count;
    scrollHighlightedIntoView();
  } else if (event.key === "Enter") {
    event.preventDefault();
    const item = filtered.value[highlighted.value];
    if (item !== undefined) choose(item);
  } else if (event.key === "Escape") {
    open.value = false;
  }
}
</script>

<template>
  <Popover v-model:open="open">
    <PopoverTrigger as-child>
      <slot name="trigger" :open="open" />
    </PopoverTrigger>
    <PopoverContent :align="align" :class="cn('p-0', contentClass)">
      <div class="border-b px-3 py-2">
        <div class="relative">
          <Search class="absolute left-2 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            v-model="search"
            :placeholder="placeholder"
            class="h-8 pl-7 text-sm"
            autofocus
            @keydown="onKeydown"
          />
        </div>
      </div>
      <div ref="listRef" class="max-h-60 overflow-y-auto p-1">
        <button
          v-for="(item, idx) in filtered"
          :key="keyOf(item, idx)"
          type="button"
          :data-index="idx"
          class="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-left text-sm"
          :class="idx === highlighted ? 'bg-accent' : 'hover:bg-accent'"
          @click="choose(item)"
          @mouseenter="highlighted = idx"
        >
          <slot name="option" :item="item" :active="idx === highlighted" />
        </button>
        <p
          v-if="filtered.length === 0"
          class="px-3 py-6 text-center text-sm text-muted-foreground"
        >
          {{ emptyText }}
        </p>
      </div>
    </PopoverContent>
  </Popover>
</template>
