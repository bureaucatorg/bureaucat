<script setup lang="ts">
import { Lock, Users as UsersIcon, Play, Pencil, Trash2, MoreHorizontal, Copy } from "lucide-vue-next";
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

    <Card v-else>
      <CardContent class="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead class="w-10"></TableHead>
              <TableHead>Name</TableHead>
              <TableHead class="hidden sm:table-cell">Filters</TableHead>
              <TableHead class="hidden md:table-cell">Grouping</TableHead>
              <TableHead class="hidden md:table-cell">Created</TableHead>
              <TableHead class="w-28 text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow
              v-for="v in views"
              :key="v.id"
              class="group"
              :class="activeSlug === v.slug ? 'bg-amber-500/5' : ''"
            >
              <TableCell class="pr-0">
                <component
                  :is="v.visibility === 'shared' ? UsersIcon : Lock"
                  class="size-3.5 text-muted-foreground"
                />
              </TableCell>
              <TableCell>
                <button
                  type="button"
                  class="text-left"
                  @click="emit('apply:view', v.slug)"
                >
                  <div class="flex items-center gap-2">
                    <span class="font-medium hover:underline">{{ v.name }}</span>
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
                </button>
              </TableCell>
              <TableCell class="hidden text-xs text-muted-foreground sm:table-cell">
                {{ predicateCount(v) }} predicate{{ predicateCount(v) === 1 ? "" : "s" }}
              </TableCell>
              <TableCell class="hidden text-xs text-muted-foreground md:table-cell">
                {{ v.group_by.replace("_", " ") }}
              </TableCell>
              <TableCell class="hidden text-xs text-muted-foreground md:table-cell">
                {{ formatDate(v.created_at) }}
              </TableCell>
              <TableCell class="text-right">
                <div class="flex items-center justify-end gap-1">
                  <Button
                    size="sm"
                    variant="ghost"
                    class="h-7 px-2 text-xs"
                    @click="emit('apply:view', v.slug)"
                  >
                    <Play class="mr-1 size-3" />
                    Apply
                  </Button>
                  <DropdownMenu v-if="canEdit(v)">
                    <DropdownMenuTrigger as-child>
                      <Button size="sm" variant="ghost" class="h-7 px-2">
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
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
