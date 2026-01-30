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
const showDropdown = ref(false);

// Members not already assigned
const availableMembers = computed(() => {
  const assignedIds = new Set(props.assignees.map((a) => a.user_id));
  return props.members.filter((m) => !assignedIds.has(m.user_id));
});

async function handleAdd(userId: string) {
  loading.value = userId;
  const result = await addAssignee(props.projectKey, props.taskNum, userId);
  loading.value = null;

  if (result.success) {
    toast.success("Assignee added");
    showDropdown.value = false;
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
        <Avatar class="size-6">
          <AvatarFallback class="text-xs">
            {{ assignee.first_name[0] }}{{ assignee.last_name[0] }}
          </AvatarFallback>
        </Avatar>
        <span class="text-sm">
          {{ assignee.first_name }} {{ assignee.last_name }}
        </span>
        <button
          v-if="isMember"
          type="button"
          class="absolute -top-1.5 -right-1.5 flex size-4 items-center justify-center rounded-full bg-foreground text-background opacity-0 shadow-sm transition-opacity group-hover:opacity-100"
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
      <DropdownMenu v-if="isMember && availableMembers.length > 0" v-model:open="showDropdown">
        <DropdownMenuTrigger as-child>
          <Button variant="outline" size="sm" class="h-8 gap-1.5">
            <Plus class="size-3.5" />
            Add
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" class="w-56">
          <DropdownMenuLabel>Add assignee</DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            v-for="member in availableMembers"
            :key="member.user_id"
            :disabled="loading === member.user_id"
            @click="handleAdd(member.user_id)"
          >
            <Avatar class="mr-2 size-6">
              <AvatarFallback class="text-xs">
                {{ member.first_name[0] }}{{ member.last_name[0] }}
              </AvatarFallback>
            </Avatar>
            {{ member.first_name }} {{ member.last_name }}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

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
