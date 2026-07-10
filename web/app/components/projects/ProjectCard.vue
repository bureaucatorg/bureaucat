<script setup lang="ts">
import { FolderKanban, Users, Building2 } from "lucide-vue-next";
import type { Project } from "~/types";

const props = defineProps<{
  project: Project;
  // When true, show which workspace the project belongs to (used when the
  // list spans all workspaces).
  showWorkspace?: boolean;
}>();

const { workspaces } = useWorkspaces();

const workspaceName = computed(
  () => workspaces.value.find((w) => w.id === props.project.workspace_id)?.name ?? ""
);

const roleBadgeVariant = computed(() => {
  switch (props.project.role) {
    case "admin":
      return "default";
    case "member":
      return "secondary";
    default:
      return "outline";
  }
});
</script>

<template>
  <NuxtLink :to="`/projects/${project.project_key}`">
    <Card
      class="group h-full cursor-pointer border-border/50 bg-background/50 transition-all hover:border-amber-500/30 hover:shadow-lg hover:shadow-amber-500/5"
    >
      <CardHeader class="pb-3">
        <div class="flex items-start justify-between">
          <div
            class="flex size-10 items-center justify-center rounded-lg bg-muted transition-colors group-hover:bg-amber-500/10"
          >
            <FolderKanban
              class="size-5 text-muted-foreground transition-colors group-hover:text-amber-600 dark:group-hover:text-amber-500"
            />
          </div>
          <Badge :variant="roleBadgeVariant" class="text-xs capitalize">
            {{ project.role }}
          </Badge>
        </div>
        <CardTitle class="mt-3 text-base font-semibold">{{ project.name }}</CardTitle>
        <span
          class="font-mono text-xs text-muted-foreground transition-colors group-hover:text-amber-600 dark:group-hover:text-amber-500"
        >
          {{ project.project_key }}
        </span>
        <div
          v-if="showWorkspace && workspaceName"
          class="mt-1 inline-flex w-fit items-center gap-1 rounded-md bg-muted px-1.5 py-0.5 text-xs text-muted-foreground"
        >
          <Building2 class="size-3 shrink-0" />
          {{ workspaceName }}
        </div>
      </CardHeader>
      <CardContent class="pt-0">
        <p
          v-if="project.description"
          class="line-clamp-2 text-sm leading-relaxed text-muted-foreground"
        >
          {{ project.description }}
        </p>
        <p v-else class="text-sm italic text-muted-foreground/50">No description</p>
      </CardContent>
    </Card>
  </NuxtLink>
</template>
