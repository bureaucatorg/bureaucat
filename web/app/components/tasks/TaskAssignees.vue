<script setup lang="ts">
import { Plus, X, Loader2 } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { TaskAssignee, ProjectMember } from "~/types";

const props = defineProps<{
  assignees: TaskAssignee[];
  projectKey: string;
  taskNum: number;
  members: ProjectMember[];
  isMember: boolean;
}>();

const emit = defineEmits<{
  refresh: [];
}>();

const { addAssignee, removeAssignee } = useTasks();

const loading = ref<string | null>(null);
const showPopover = ref(false);

// Members not already assigned
const availableMembers = computed(() => {
  const assignedIds = new Set(props.assignees.map((a) => a.user_id));
  return props.members.filter((m) => !assignedIds.has(m.user_id));
});

function memberSearchText(m: ProjectMember) {
  return `${m.first_name} ${m.last_name} ${m.username}`;
}

async function handleAdd(userId: string) {
  loading.value = userId;
  const result = await addAssignee(props.projectKey, props.taskNum, userId);
  loading.value = null;

  if (result.success) {
    toast.success("Assignee added");
    showPopover.value = false;
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to add assignee");
  }
}

async function handleRemove(userId: string) {
  loading.value = userId;
  const result = await removeAssignee(props.projectKey, props.taskNum, userId);
  loading.value = null;

  if (result.success) {
    toast.success("Assignee removed");
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to remove assignee");
  }
}
</script>

<template>
  <div class="space-y-2">
    <p class="text-xs text-muted-foreground">Assignees</p>

    <div class="flex flex-wrap items-center gap-2">
      <div
        v-for="assignee in assignees"
        :key="assignee.id"
        class="group relative flex items-center gap-1.5 rounded-md border bg-muted/50 py-1 pl-1 pr-2.5"
      >
        <NuxtLink :to="`/profile/${assignee.user_id}`" class="flex items-center gap-1.5 hover:opacity-80 transition-opacity">
          <Avatar class="size-6">
            <AvatarImage v-if="assignee.avatar_url" :src="assignee.avatar_url" />
            <AvatarFallback class="text-xs" :seed="assignee.user_id">
              {{ assignee.first_name[0] }}{{ assignee.last_name[0] }}
            </AvatarFallback>
          </Avatar>
          <span class="text-sm">
            {{ assignee.first_name }} {{ assignee.last_name }}
          </span>
        </NuxtLink>
        <button
          v-if="isMember"
          type="button"
          :aria-label="`Remove ${assignee.first_name} ${assignee.last_name}`"
          class="absolute -top-1.5 -right-1.5 flex size-4 items-center justify-center rounded-full bg-foreground text-background opacity-0 shadow-sm transition-opacity group-hover:opacity-100 focus:opacity-100 focus-visible:ring-2 focus-visible:ring-ring outline-none"
          :disabled="loading === assignee.user_id"
          @click="handleRemove(assignee.user_id)"
        >
          <Loader2
            v-if="loading === assignee.user_id"
            class="size-2.5 animate-spin"
          />
          <X v-else class="size-2.5" />
        </button>
      </div>

      <!-- Add button -->
      <SearchableSelect
        v-if="isMember && availableMembers.length > 0"
        v-model:open="showPopover"
        :items="availableMembers"
        :get-search-text="memberSearchText"
        :get-key="(m) => m.user_id"
        placeholder="Search members..."
        empty-text="No members found"
        @select="(m) => handleAdd(m.user_id)"
      >
        <template #trigger>
          <Button variant="outline" size="sm" class="h-8 gap-1.5">
            <Plus class="size-3.5" />
            Add
          </Button>
        </template>
        <template #option="{ item: member }">
          <Avatar class="size-6">
            <AvatarImage v-if="member.avatar_url" :src="member.avatar_url" />
            <AvatarFallback class="text-xs" :seed="member.user_id">
              {{ member.first_name[0] }}{{ member.last_name[0] }}
            </AvatarFallback>
          </Avatar>
          {{ member.first_name }} {{ member.last_name }}
        </template>
      </SearchableSelect>

      <!-- Empty state -->
      <span
        v-if="assignees.length === 0 && !isMember"
        class="text-sm text-muted-foreground"
      >
        No assignees
      </span>
    </div>
  </div>
</template>
