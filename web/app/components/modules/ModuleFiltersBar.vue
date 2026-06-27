<script setup lang="ts">
import { ChevronsUpDown } from "lucide-vue-next";
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

// reka-ui's Select reserves the empty string for "no selection", so the
// "All statuses" choice uses a sentinel that maps back to undefined.
const ALL_STATUSES = "__all__";

const statusValue = computed({
  get: () => model.value.status ?? ALL_STATUSES,
  set: (v: string) => {
    model.value = {
      ...model.value,
      status: v === ALL_STATUSES ? undefined : (v as ModuleStatus),
    };
  },
});

const sortValue = computed({
  get: () => `${model.value.sort_by ?? "created_at"}:${model.value.sort_dir ?? "desc"}`,
  set: (v: string) => {
    const [by, dir] = v.split(":");
    model.value = {
      ...model.value,
      sort_by: by as ModuleListFilters["sort_by"],
      sort_dir: dir as ModuleListFilters["sort_dir"],
    };
  },
});

// Lead picker uses the same searchable popover as the task assignee selector.
// A null-id option clears the filter ("Any lead").
interface LeadOption {
  id: string | null;
  label: string;
  member?: ProjectMember;
}

const leadOpen = ref(false);

const leadOptions = computed<LeadOption[]>(() => [
  { id: null, label: "Any lead" },
  ...(members.value as ProjectMember[]).map((m) => ({
    id: m.user_id,
    label: `${m.first_name} ${m.last_name}`,
    member: m,
  })),
]);

const selectedLead = computed(() =>
  (members.value as ProjectMember[]).find((m) => m.user_id === model.value.lead_id)
);

function leadSearchText(o: LeadOption) {
  return o.member ? `${o.member.first_name} ${o.member.last_name} ${o.member.username}` : o.label;
}

function selectLead(o: LeadOption) {
  model.value = { ...model.value, lead_id: o.id ?? undefined };
}
</script>

<template>
  <div class="flex flex-wrap items-center gap-2">
    <Select v-model="statusValue">
      <SelectTrigger class="h-9 w-[10rem]">
        <SelectValue placeholder="All statuses" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem :value="ALL_STATUSES">All statuses</SelectItem>
        <SelectItem v-for="s in MODULE_STATUSES" :key="s" :value="s" class="capitalize">
          {{ s.replace("_", " ") }}
        </SelectItem>
      </SelectContent>
    </Select>

    <SearchableSelect
      v-model:open="leadOpen"
      :items="leadOptions"
      :get-search-text="leadSearchText"
      :get-key="(o) => o.id ?? '__any__'"
      placeholder="Search members..."
      empty-text="No members found"
      @select="selectLead"
    >
      <template #trigger>
        <Button
          variant="outline"
          role="combobox"
          :aria-expanded="leadOpen"
          class="h-9 w-[10rem] justify-between gap-1 px-3 font-normal"
        >
          <span class="flex min-w-0 items-center gap-1.5">
            <Avatar v-if="selectedLead" class="size-5">
              <AvatarImage v-if="selectedLead.avatar_url" :src="selectedLead.avatar_url" />
              <AvatarFallback class="text-[10px]" :seed="selectedLead.user_id">
                {{ selectedLead.first_name[0] }}{{ selectedLead.last_name[0] }}
              </AvatarFallback>
            </Avatar>
            <span class="truncate">
              {{ selectedLead ? `${selectedLead.first_name} ${selectedLead.last_name}` : "Any lead" }}
            </span>
          </span>
          <ChevronsUpDown class="size-4 shrink-0 opacity-50" />
        </Button>
      </template>
      <template #option="{ item: option }">
        <Avatar v-if="option.member" class="size-6">
          <AvatarImage v-if="option.member.avatar_url" :src="option.member.avatar_url" />
          <AvatarFallback class="text-xs" :seed="option.member.user_id">
            {{ option.member.first_name[0] }}{{ option.member.last_name[0] }}
          </AvatarFallback>
        </Avatar>
        <span :class="{ 'text-muted-foreground': !option.member }">{{ option.label }}</span>
      </template>
    </SearchableSelect>

    <Select v-model="sortValue">
      <SelectTrigger class="h-9 w-[12rem]">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="created_at:desc">Newest first</SelectItem>
        <SelectItem value="created_at:asc">Oldest first</SelectItem>
        <SelectItem value="end_date:asc">End date · soonest</SelectItem>
        <SelectItem value="end_date:desc">End date · latest</SelectItem>
        <SelectItem value="progress:desc">Progress · highest</SelectItem>
        <SelectItem value="progress:asc">Progress · lowest</SelectItem>
      </SelectContent>
    </Select>
  </div>
</template>
