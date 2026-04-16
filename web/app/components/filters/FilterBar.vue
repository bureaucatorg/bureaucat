<script setup lang="ts">
import { Search, Plus, X } from "lucide-vue-next";
import type {
  FilterTree,
  FilterNode,
  Predicate,
  ProjectState,
  ProjectMember,
  ProjectLabel,
  SortKey,
  SortDir,
  ViewGroupBy,
} from "~/types";
import FilterChip from "./FilterChip.vue";
import FilterPredicateEditor from "./FilterPredicateEditor.vue";
import SortMenu from "./SortMenu.vue";
import BoardGroupBySelect from "~/components/board/BoardGroupBySelect.vue";

const props = defineProps<{
  tree: FilterTree;
  searchQuery: string;
  sortBy: SortKey;
  sortDir: SortDir;
  groupBy: ViewGroupBy;
  states: ProjectState[];
  labels: ProjectLabel[];
  members: ProjectMember[];
  /** When true, show the Group-by control (board tab only). */
  showGroupBy?: boolean;
}>();

const emit = defineEmits<{
  "update:tree": [tree: FilterTree];
  "update:searchQuery": [value: string];
  "update:sortBy": [value: SortKey];
  "update:sortDir": [value: SortDir];
  "update:groupBy": [value: ViewGroupBy];
  reset: [];
}>();

const addingFilter = ref(false);

function hasAnyFilter(): boolean {
  return props.tree.children.length > 0 || props.searchQuery.length > 0;
}

function updateChildAt(index: number, node: FilterNode) {
  const next = [...props.tree.children];
  next[index] = node;
  emit("update:tree", { children: next });
}

function removeChildAt(index: number) {
  const next = props.tree.children.filter((_, i) => i !== index);
  emit("update:tree", { children: next });
}

function addTopLevelPredicate(p: Predicate) {
  emit("update:tree", { children: [...props.tree.children, { predicate: p }] });
  addingFilter.value = false;
}

function updatePredicate(index: number, p: Predicate) {
  updateChildAt(index, { predicate: p });
}
</script>

<template>
  <div class="space-y-2">
    <!-- Top row: search, add-filter, group-by, sort, reset -->
    <div class="flex flex-wrap items-center gap-2">
      <div class="relative flex-1 sm:max-w-xs">
        <Search class="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          :model-value="searchQuery"
          placeholder="Search tasks…"
          class="pl-9"
          @update:model-value="(v) => emit('update:searchQuery', String(v ?? ''))"
        />
      </div>

      <Popover v-model:open="addingFilter">
        <PopoverTrigger as-child>
          <Button variant="outline" size="sm" class="gap-1.5">
            <Plus class="size-3.5" />
            Filter
          </Button>
        </PopoverTrigger>
        <PopoverContent align="start" class="p-0">
          <FilterPredicateEditor
            :states="states"
            :labels="labels"
            :members="members"
            @confirm="addTopLevelPredicate"
            @cancel="addingFilter = false"
          />
        </PopoverContent>
      </Popover>

      <BoardGroupBySelect
        v-if="showGroupBy"
        :model-value="groupBy"
        @update:model-value="(v) => emit('update:groupBy', v)"
      />

      <SortMenu
        :sort-by="sortBy"
        :sort-dir="sortDir"
        @update:sort-by="(v) => emit('update:sortBy', v)"
        @update:sort-dir="(v) => emit('update:sortDir', v)"
      />

      <Button
        v-if="hasAnyFilter()"
        variant="ghost"
        size="sm"
        @click="emit('reset')"
      >
        <X class="mr-1 size-3.5" />
        Clear
      </Button>
    </div>

    <!-- Active filter chips (implicit AND between all chips) -->
    <div
      v-if="tree.children.length > 0"
      class="flex flex-wrap items-center gap-1.5"
    >
      <template v-for="(child, i) in tree.children" :key="i">
        <FilterChip
          v-if="child.predicate"
          :predicate="child.predicate"
          :states="states"
          :labels="labels"
          :members="members"
          @update="(p) => updatePredicate(i, p)"
          @remove="removeChildAt(i)"
        />
      </template>
    </div>
  </div>
</template>
