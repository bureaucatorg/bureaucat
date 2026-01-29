<script setup lang="ts">
import { FolderKanban, Users } from "lucide-vue-next";
import type { Project } from "~/types";

const props = defineProps<{
  project: Project;
}>();

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
        <div class="flex items-center gap-2">
          <span
            class="font-mono text-xs text-muted-foreground transition-colors group-hover:text-amber-600 dark:group-hover:text-amber-500"
          >
            {{ project.project_key }}
          </span>
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
