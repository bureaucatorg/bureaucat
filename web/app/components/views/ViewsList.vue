<script setup lang="ts">
import { Lock, Users as UsersIcon, Play, Pencil, Trash2, MoreHorizontal, Copy, Filter, Layers, Calendar } from "lucide-vue-next";
import type { ProjectView } from "~/types";

const props = defineProps<{
  projectKey: string;
  views: ProjectView[];
  activeSlug: string | null;
  currentUserId?: string;
  isAdmin: boolean;
}>();

const emit = defineEmits<{
  "apply:view": [slug: string];
  "rename:view": [view: ProjectView];
  "refresh": [];
}>();

const { updateView, deleteView } = useViews();

async function toggleVisibility(view: ProjectView) {
  const next = view.visibility === "shared" ? "private" : "shared";
  await updateView(props.projectKey, view.slug, { visibility: next });
  emit("refresh");
}

async function handleDelete(view: ProjectView) {
  if (!confirm(`Delete view "${view.name}"?`)) return;
  await deleteView(props.projectKey, view.slug);
  emit("refresh");
}

function isOwner(v: ProjectView): boolean {
  return props.currentUserId !== undefined && v.owner_id === props.currentUserId;
}

function canEdit(v: ProjectView): boolean {
  return isOwner(v) || (v.visibility === "shared" && props.isAdmin);
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function predicateCount(v: ProjectView): number {
  return (v.filter_tree?.children ?? []).filter((c) => c.predicate).length;
}

function groupByLabel(groupBy: string): string {
  if (!groupBy || groupBy === "none") return "None";
  return groupBy.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}
</script>

<template>
  <div>
    <div
      v-if="views.length === 0"
      class="flex flex-col items-center justify-center rounded-lg border border-dashed py-16"
    >
      <UsersIcon class="size-8 text-muted-foreground" />
      <h3 class="mt-4 font-semibold">No views yet</h3>
      <p class="mt-1 max-w-sm text-center text-sm text-muted-foreground">
        Save a combination of filters as a view to jump back to it quickly. Share
        a view with the project so teammates can see the same slice of work.
      </p>
    </div>

    <div v-else class="space-y-2">
      <div
        v-for="v in views"
        :key="v.id"
        class="group relative rounded-lg border bg-card transition-colors hover:border-border/80 hover:bg-accent/30"
        :class="activeSlug === v.slug ? 'border-amber-500/40 bg-amber-500/5' : ''"
      >
        <div class="flex items-center gap-4 px-4 py-3">
          <!-- Visibility icon -->
          <div class="flex size-8 shrink-0 items-center justify-center rounded-md bg-muted/50">
            <component
              :is="v.visibility === 'shared' ? UsersIcon : Lock"
              class="size-3.5 text-muted-foreground"
            />
          </div>

          <!-- Name & description -->
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <button
                type="button"
                class="text-left font-medium hover:underline"
                @click="emit('apply:view', v.slug)"
              >
                {{ v.name }}
              </button>
              <span
                v-if="activeSlug === v.slug"
                class="rounded-full bg-amber-500/15 px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wider text-amber-700 dark:text-amber-300"
              >
                active
              </span>
            </div>
            <p
              v-if="v.description"
              class="mt-0.5 line-clamp-1 text-xs text-muted-foreground"
            >
              {{ v.description }}
            </p>
          </div>

          <!-- Meta pills -->
          <div class="hidden items-center gap-3 sm:flex">
            <div class="flex items-center gap-1.5 text-xs text-muted-foreground" :title="`${predicateCount(v)} filter${predicateCount(v) === 1 ? '' : 's'}`">
              <Filter class="size-3" />
              <span>{{ predicateCount(v) }}</span>
            </div>
            <div v-if="v.group_by && v.group_by !== 'none'" class="flex items-center gap-1.5 text-xs text-muted-foreground" :title="`Grouped by ${groupByLabel(v.group_by)}`">
              <Layers class="size-3" />
              <span>{{ groupByLabel(v.group_by) }}</span>
            </div>
            <div class="hidden items-center gap-1.5 text-xs text-muted-foreground md:flex" :title="formatDate(v.created_at)">
              <Calendar class="size-3" />
              <span>{{ formatDate(v.created_at) }}</span>
            </div>
          </div>

          <!-- Actions -->
          <div class="flex items-center gap-1">
            <Button
              size="sm"
              variant="outline"
              class="h-7 gap-1.5 px-3 text-xs"
              @click="emit('apply:view', v.slug)"
            >
              <Play class="size-3" />
              Apply
            </Button>
            <DropdownMenu v-if="canEdit(v)">
              <DropdownMenuTrigger as-child>
                <Button size="sm" variant="ghost" class="h-7 w-7 p-0 opacity-0 transition-opacity group-hover:opacity-100">
                  <MoreHorizontal class="size-3.5" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" class="w-44">
                <DropdownMenuItem @click="emit('rename:view', v)">
                  <Pencil class="mr-2 size-3.5" /> Rename
                </DropdownMenuItem>
                <DropdownMenuItem v-if="isOwner(v)" @click="toggleVisibility(v)">
                  <Copy class="mr-2 size-3.5" />
                  {{ v.visibility === "shared" ? "Make private" : "Share with project" }}
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem class="text-destructive" @click="handleDelete(v)">
                  <Trash2 class="mr-2 size-3.5" /> Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
