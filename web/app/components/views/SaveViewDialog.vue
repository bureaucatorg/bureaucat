<script setup lang="ts">
import type { ViewVisibility, FilterTree, ViewGroupBy, SortKey, SortDir } from "~/types";

const props = defineProps<{
  open: boolean;
  projectKey: string;
  /** When provided, pre-fill form for editing; otherwise create-mode. */
  initial?: {
    slug: string;
    name: string;
    description?: string;
    visibility: ViewVisibility;
  };
  /** Filters to snapshot into the view (create-mode only). */
  currentTree: FilterTree;
  currentGroupBy: ViewGroupBy;
  currentSortBy: SortKey;
  currentSortDir: SortDir;
}>();

const emit = defineEmits<{
  "update:open": [open: boolean];
  saved: [slug: string];
}>();

const { createView, updateView } = useViews();

const name = ref("");
const description = ref("");
const visibility = ref<ViewVisibility>("private");
const saving = ref(false);
const error = ref<string | null>(null);

watch(
  () => props.open,
  (open) => {
    if (open) {
      name.value = props.initial?.name ?? suggestName(props.currentTree);
      description.value = props.initial?.description ?? "";
      visibility.value = props.initial?.visibility ?? "private";
      error.value = null;
    }
  }
);

function suggestName(tree: FilterTree): string {
  if (tree.children.length === 0) return "Untitled view";
  return `New view`;
}

async function save() {
  if (!name.value.trim()) {
    error.value = "Name is required";
    return;
  }
  saving.value = true;
  try {
    if (props.initial) {
      const result = await updateView(props.projectKey, props.initial.slug, {
        name: name.value.trim(),
        description: description.value.trim() || null,
        visibility: visibility.value,
      });
      if (!result.success) {
        error.value = result.error || "Failed to update view";
        return;
      }
      emit("saved", props.initial.slug);
    } else {
      const result = await createView(props.projectKey, {
        name: name.value.trim(),
        description: description.value.trim() || undefined,
        visibility: visibility.value,
        filter_tree: props.currentTree,
        group_by: props.currentGroupBy,
        sort_by: props.currentSortBy,
        sort_dir: props.currentSortDir,
      });
      if (!result.success || !result.data) {
        error.value = result.error || "Failed to create view";
        return;
      }
      emit("saved", result.data.slug);
    }
    emit("update:open", false);
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="(v) => emit('update:open', v)">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>{{ initial ? "Edit view" : "Save as view" }}</DialogTitle>
        <DialogDescription>
          {{
            initial
              ? "Rename this view or change who can see it."
              : "Snapshot the current filters into a named view you can return to."
          }}
        </DialogDescription>
      </DialogHeader>

      <form class="space-y-4" @submit.prevent="save">
        <div class="space-y-1.5">
          <Label for="view-name">Name</Label>
          <Input id="view-name" v-model="name" placeholder="e.g. Overdue for me" />
        </div>
        <div class="space-y-1.5">
          <Label for="view-desc">Description</Label>
          <Textarea id="view-desc" v-model="description" placeholder="Optional" rows="2" />
        </div>
        <div class="space-y-1.5">
          <Label>Visibility</Label>
          <div class="flex gap-2">
            <button
              type="button"
              class="flex-1 rounded-md border px-3 py-2 text-left text-sm transition-colors"
              :class="visibility === 'private' ? 'border-primary bg-primary/5' : 'hover:border-muted-foreground/50'"
              @click="visibility = 'private'"
            >
              <div class="font-medium">Private</div>
              <div class="text-xs text-muted-foreground">Only you</div>
            </button>
            <button
              type="button"
              class="flex-1 rounded-md border px-3 py-2 text-left text-sm transition-colors"
              :class="visibility === 'shared' ? 'border-primary bg-primary/5' : 'hover:border-muted-foreground/50'"
              @click="visibility = 'shared'"
            >
              <div class="font-medium">Shared</div>
              <div class="text-xs text-muted-foreground">Everyone in the project</div>
            </button>
          </div>
        </div>
        <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">
            Cancel
          </Button>
          <Button type="submit" :disabled="saving">
            {{ initial ? "Save" : "Create view" }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
