<script setup lang="ts">
import { FolderPlus, Loader2, X } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { NewUserDefaultProject } from "~/composables/useSettings";
import type { PaginatedProjectsResponse, Project } from "~/types/project";

const { fetchNewUserDefaults, updateNewUserDefaults } = useSettings();
const { getAuthHeader } = useAuth();

const projects = ref<NewUserDefaultProject[]>([]);
const loaded = ref(false);
const saving = ref(false);
const searchLoading = ref(false);
const searchResults = ref<Project[]>([]);

const availableProjects = computed(() => {
  const chosen = new Set(projects.value.map((p) => p.project_id));
  return searchResults.value.filter((p) => !chosen.has(p.id));
});

let searchTimeout: ReturnType<typeof setTimeout>;
function onSearch(query: string) {
  clearTimeout(searchTimeout);
  searchLoading.value = true;
  searchTimeout = setTimeout(async () => {
    try {
      const response = await fetch(
        `/api/v1/projects?page=1&per_page=20&search=${encodeURIComponent(query.trim())}`,
        { headers: getAuthHeader() }
      );
      const data: PaginatedProjectsResponse = await response.json();
      searchResults.value = response.ok ? data.projects || [] : [];
    } catch {
      searchResults.value = [];
    } finally {
      searchLoading.value = false;
    }
  }, 300);
}

function addProject(p: Project) {
  projects.value = [
    ...projects.value,
    { project_id: p.id, role: "member", project_key: p.project_key, project_name: p.name },
  ];
}

function removeProject(id: string) {
  projects.value = projects.value.filter((p) => p.project_id !== id);
}

async function handleSave() {
  saving.value = true;
  const result = await updateNewUserDefaults({
    projects: projects.value.map((p) => ({ project_id: p.project_id, role: p.role })),
  });
  saving.value = false;
  if (result.success && result.data) {
    projects.value = result.data.projects;
    toast.success("New user defaults saved");
  } else {
    toast.error(result.error || "Failed to save new user defaults");
  }
}

onMounted(async () => {
  const result = await fetchNewUserDefaults();
  if (result.success && result.data) projects.value = result.data.projects;
  loaded.value = true;
});
</script>

<template>
  <Card>
    <CardContent class="space-y-4 pt-6">
      <div class="flex items-center gap-3">
        <FolderPlus class="size-5 shrink-0 text-muted-foreground" />
        <div>
          <p class="font-medium">New User Defaults</p>
          <p class="text-sm text-muted-foreground">
            Projects every new user joins, along with their workspaces. Existing users are not changed.
          </p>
        </div>
      </div>

      <div class="rounded-md border px-3 py-2">
        <TokenSelect
          :selected="[]"
          :available="availableProjects"
          :get-key="(p) => p.id"
          :get-search-text="(p) => `${p.project_key} ${p.name}`"
          server-search
          :loading="searchLoading"
          placeholder="Search projects to add..."
          empty-text="No projects found"
          :disabled="saving || !loaded"
          @search="onSearch"
          @add="addProject"
        >
          <template #option="{ item: p }">
            <span class="font-mono text-xs text-muted-foreground">{{ p.project_key }}</span>
            <span class="truncate text-sm">{{ p.name }}</span>
          </template>
        </TokenSelect>
      </div>

      <ul v-if="projects.length" class="divide-y rounded-md border">
        <li
          v-for="p in projects"
          :key="p.project_id"
          class="flex flex-wrap items-center gap-x-3 gap-y-2 px-3 py-2"
        >
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm">
              <span class="font-mono text-xs text-muted-foreground">{{ p.project_key }}</span>
              {{ p.project_name }}
            </p>
            <p v-if="p.workspace_name" class="truncate text-xs text-muted-foreground">
              {{ p.workspace_name }}
            </p>
          </div>
          <NativeSelect v-model="p.role" class="w-28" :disabled="saving" aria-label="Role">
            <option value="member">Member</option>
            <option value="guest">Guest</option>
          </NativeSelect>
          <Button
            variant="ghost"
            size="icon"
            :aria-label="`Remove ${p.project_key}`"
            :disabled="saving"
            @click="removeProject(p.project_id)"
          >
            <X class="size-4" />
          </Button>
        </li>
      </ul>
      <p v-else-if="loaded" class="text-sm text-muted-foreground">
        No default projects. New users start with no workspaces or projects.
      </p>

      <div class="flex justify-end">
        <Button :disabled="saving || !loaded" @click="handleSave">
          <Loader2 v-if="saving" class="mr-2 size-4 animate-spin" />
          Save Defaults
        </Button>
      </div>
    </CardContent>
  </Card>
</template>
