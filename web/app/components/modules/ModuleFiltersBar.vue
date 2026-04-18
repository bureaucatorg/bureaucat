<script setup lang="ts">
import type { ModuleListFilters, ModuleStatus, ProjectMember } from "~/types";
import { MODULE_STATUSES } from "~/types";

const props = defineProps<{
  projectKey: string;
}>();

const model = defineModel<ModuleListFilters>({ required: true });

const { members, listMembers } = useProjects();

onMounted(async () => {
  if (!members.value.length) await listMembers(props.projectKey);
});

function updateStatus(v: string) {
  model.value = {
    ...model.value,
    status: (v || undefined) as ModuleStatus | undefined,
  };
}
function updateLead(v: string) {
  model.value = { ...model.value, lead_id: v || undefined };
}
function updateSort(v: string) {
  const [by, dir] = v.split(":");
  model.value = {
    ...model.value,
    sort_by: by as ModuleListFilters["sort_by"],
    sort_dir: dir as ModuleListFilters["sort_dir"],
  };
}

const currentSort = computed(
  () => `${model.value.sort_by ?? "created_at"}:${model.value.sort_dir ?? "desc"}`
);
</script>

<template>
  <div class="flex flex-wrap items-center gap-2">
    <select
      :value="model.status ?? ''"
      class="h-9 rounded-md border border-input bg-transparent px-3 text-sm"
      @change="updateStatus(($event.target as HTMLSelectElement).value)"
    >
      <option value="">All statuses</option>
      <option v-for="s in MODULE_STATUSES" :key="s" :value="s">
        {{ s.replace("_", " ") }}
      </option>
    </select>

    <select
      :value="model.lead_id ?? ''"
      class="h-9 rounded-md border border-input bg-transparent px-3 text-sm"
      @change="updateLead(($event.target as HTMLSelectElement).value)"
    >
      <option value="">Any lead</option>
      <option v-for="m in (members as ProjectMember[])" :key="m.user_id" :value="m.user_id">
        {{ m.first_name }} {{ m.last_name }}
      </option>
    </select>

    <select
      :value="currentSort"
      class="h-9 rounded-md border border-input bg-transparent px-3 text-sm"
      @change="updateSort(($event.target as HTMLSelectElement).value)"
    >
      <option value="created_at:desc">Newest first</option>
      <option value="created_at:asc">Oldest first</option>
      <option value="end_date:asc">End date · soonest</option>
      <option value="end_date:desc">End date · latest</option>
      <option value="progress:desc">Progress · highest</option>
      <option value="progress:asc">Progress · lowest</option>
    </select>
  </div>
</template>
