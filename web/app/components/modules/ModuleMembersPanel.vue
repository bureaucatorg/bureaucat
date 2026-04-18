<script setup lang="ts">
import { UserPlus, X } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { ModuleUserBrief, ProjectMember } from "~/types";

const props = defineProps<{
  projectKey: string;
  moduleId: string;
  leadId?: string;
  members: ModuleUserBrief[];
  isAdmin: boolean;
}>();

const emit = defineEmits<{ changed: [] }>();

const { members: projectMembers, listMembers } = useProjects();
const { addModuleMember, removeModuleMember } = useModules();

const addOpen = ref(false);

async function ensureProjectMembers() {
  if (!projectMembers.value.length) await listMembers(props.projectKey);
}

const candidates = computed<ProjectMember[]>(() => {
  const already = new Set(props.members.map((m) => m.user_id));
  return (projectMembers.value as ProjectMember[]).filter(
    (m) => !already.has(m.user_id)
  );
});

async function handleAdd(userId: string) {
  const result = await addModuleMember(props.projectKey, props.moduleId, userId);
  if (result.success) {
    toast.success("Member added");
    addOpen.value = false;
    emit("changed");
  } else {
    toast.error(result.error || "Failed to add member");
  }
}

async function handleRemove(userId: string) {
  const result = await removeModuleMember(props.projectKey, props.moduleId, userId);
  if (result.success) {
    toast.success("Member removed");
    emit("changed");
  } else {
    toast.error(result.error || "Failed to remove member");
  }
}

function initials(first: string, last: string, username: string): string {
  const f = (first || "").trim()[0] || "";
  const l = (last || "").trim()[0] || "";
  if (f || l) return (f + l).toUpperCase();
  return (username || "?").slice(0, 2).toUpperCase();
}
</script>

<template>
  <section class="space-y-3">
    <div class="flex items-center justify-between">
      <h3 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Members
      </h3>
      <Popover v-if="isAdmin" v-model:open="addOpen">
        <PopoverTrigger as-child>
          <Button size="sm" variant="ghost" @click="ensureProjectMembers">
            <UserPlus class="mr-1 size-3.5" /> Add
          </Button>
        </PopoverTrigger>
        <PopoverContent class="w-64 p-0">
          <div class="max-h-60 overflow-y-auto">
            <button
              v-for="m in candidates"
              :key="m.user_id"
              type="button"
              class="flex w-full items-center gap-2 px-3 py-2 text-sm hover:bg-muted"
              @click="handleAdd(m.user_id)"
            >
              <Avatar class="size-6">
                <AvatarImage v-if="m.avatar_url" :src="m.avatar_url" />
                <AvatarFallback class="text-[10px]" :seed="m.user_id">
                  {{ initials(m.first_name, m.last_name, m.username) }}
                </AvatarFallback>
              </Avatar>
              <span class="truncate text-left">
                <span class="block truncate">{{ m.first_name }} {{ m.last_name }}</span>
                <span class="block truncate text-[10px] text-muted-foreground">@{{ m.username }}</span>
              </span>
            </button>
            <div
              v-if="!candidates.length"
              class="px-3 py-4 text-center text-xs text-muted-foreground"
            >
              All project members are already in this module.
            </div>
          </div>
        </PopoverContent>
      </Popover>
    </div>

    <ul class="space-y-1.5">
      <li
        v-for="m in members"
        :key="m.user_id"
        class="group flex items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-muted/40"
      >
        <Avatar class="size-6 shrink-0">
          <AvatarImage v-if="m.avatar_url" :src="m.avatar_url" />
          <AvatarFallback class="text-[9px]" :seed="m.user_id">
            {{ initials(m.first_name, m.last_name, m.username) }}
          </AvatarFallback>
        </Avatar>
        <span class="min-w-0 flex-1 truncate">
          {{ m.first_name }} {{ m.last_name }}
          <span v-if="leadId === m.user_id" class="ml-1 rounded bg-amber-500/10 px-1.5 py-0.5 text-[10px] font-medium uppercase text-amber-700 dark:text-amber-500">
            Lead
          </span>
        </span>
        <button
          v-if="isAdmin && leadId !== m.user_id"
          type="button"
          class="opacity-0 transition-opacity group-hover:opacity-100"
          :title="`Remove ${m.first_name}`"
          @click="handleRemove(m.user_id)"
        >
          <X class="size-4 text-muted-foreground hover:text-destructive" />
        </button>
      </li>
      <li v-if="!members.length" class="py-4 text-center text-xs text-muted-foreground">
        No members yet.
      </li>
    </ul>
  </section>
</template>
