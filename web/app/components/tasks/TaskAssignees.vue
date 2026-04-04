<script setup lang="ts">
import { Plus, X, Loader2, Search } from "lucide-vue-next";
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
const searchQuery = ref("");

// Members not already assigned
const availableMembers = computed(() => {
  const assignedIds = new Set(props.assignees.map((a) => a.user_id));
  return props.members.filter((m) => !assignedIds.has(m.user_id));
});

const filteredMembers = computed(() => {
  const q = searchQuery.value.toLowerCase().trim();
  if (!q) return availableMembers.value;
  return availableMembers.value.filter(
    (m) =>
      m.first_name.toLowerCase().includes(q) ||
      m.last_name.toLowerCase().includes(q) ||
      m.username.toLowerCase().includes(q)
  );
});

watch(showPopover, (open) => {
  if (!open) searchQuery.value = "";
});

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
            <AvatarFallback class="text-xs">
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
      <Popover v-if="isMember && availableMembers.length > 0" v-model:open="showPopover">
        <PopoverTrigger as-child>
          <Button variant="outline" size="sm" class="h-8 gap-1.5">
            <Plus class="size-3.5" />
            Add
          </Button>
        </PopoverTrigger>
        <PopoverContent align="start" class="w-56 p-0">
          <div class="border-b px-3 py-2">
            <div class="relative">
              <Search class="absolute left-2 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
              <Input
                v-model="searchQuery"
                placeholder="Search members..."
                class="h-8 pl-7 text-sm"
              />
            </div>
          </div>
          <div class="max-h-48 overflow-y-auto">
            <div class="py-1">
              <button
                v-for="member in filteredMembers"
                :key="member.user_id"
                type="button"
                class="flex w-full items-center gap-2 px-3 py-1.5 text-sm hover:bg-accent disabled:opacity-50"
                :disabled="loading === member.user_id"
                @click="handleAdd(member.user_id)"
              >
                <Avatar class="size-6">
                  <AvatarImage v-if="member.avatar_url" :src="member.avatar_url" />
                  <AvatarFallback class="text-xs">
                    {{ member.first_name[0] }}{{ member.last_name[0] }}
                  </AvatarFallback>
                </Avatar>
                {{ member.first_name }} {{ member.last_name }}
              </button>
              <p
                v-if="filteredMembers.length === 0"
                class="px-3 py-2 text-center text-sm text-muted-foreground"
              >
                No members found
              </p>
            </div>
          </div>
        </PopoverContent>
      </Popover>

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
