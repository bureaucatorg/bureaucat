<script setup lang="ts">
import { Check, ChevronsUpDown, Plus, Settings2, Building2 } from "lucide-vue-next";
import type { Workspace } from "~/types";

const { user } = useAuth();
const { workspaces, currentWorkspace, setCurrentWorkspace } = useWorkspaces();
const route = useRoute();
const router = useRouter();

const isAdmin = computed(() => user.value?.user_type === "admin");
const showCreate = ref(false);

// One-time onboarding coach-mark: a persistent hint pointing at the switcher,
// shown until the user opens it once (dismissal persisted so it never returns).
const HINT_KEY = "bureaucat.workspaceSwitcherHintSeen";
const showHint = ref(false);

onMounted(() => {
  if (typeof window !== "undefined" && !localStorage.getItem(HINT_KEY)) {
    showHint.value = true;
  }
});

function dismissHint() {
  if (!showHint.value) return;
  showHint.value = false;
  if (typeof window !== "undefined") localStorage.setItem(HINT_KEY, "1");
}

const triggerLabel = computed(() => {
  const ws = currentWorkspace.value;
  return ws ? ws.workspace_key.slice(0, 2) : "";
});

async function selectWorkspace(ws: Workspace) {
  if (ws.id === currentWorkspace.value?.id) return;
  setCurrentWorkspace(ws);
  // A project detail page belongs to the previous workspace, so bounce back to
  // the project list. Other pages (dashboard, settings) stay put.
  if (route.path.startsWith("/projects/")) {
    await router.push("/projects");
  }
}
</script>

<template>
  <div class="relative">
    <!-- One-time coach-mark, dismissed after the switcher is opened once. -->
    <div
      v-if="showHint"
      class="pointer-events-none absolute left-full top-1/2 z-50 ml-2 flex -translate-y-1/2 items-center"
    >
      <span
        class="size-0 border-y-[6px] border-r-[6px] border-y-transparent border-r-foreground"
      />
      <div
        class="whitespace-nowrap rounded-lg bg-foreground px-3 py-2 text-xs font-medium text-background shadow-lg"
      >
        Switch Workspace
      </div>
    </div>

    <DropdownMenu>
      <DropdownMenuTrigger as-child>
        <button
          type="button"
          :title="currentWorkspace ? currentWorkspace.name : 'Workspaces'"
          aria-label="Switch workspace"
          class="group relative flex size-9 items-center justify-center rounded-md bg-amber-500/10 text-xs font-semibold text-amber-700 outline-none transition-colors hover:bg-amber-500/20 focus-visible:ring-2 focus-visible:ring-ring dark:text-amber-400"
          @click="dismissHint"
        >
          <span v-if="triggerLabel" class="font-mono">{{ triggerLabel }}</span>
          <Building2 v-else class="size-4.5" />
          <ChevronsUpDown
            class="absolute -bottom-0.5 -right-0.5 size-2.5 text-muted-foreground/70"
          />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent side="right" align="start" class="w-56">
        <DropdownMenuLabel class="text-xs text-muted-foreground">
          Workspaces
        </DropdownMenuLabel>
        <DropdownMenuSeparator />

        <div v-if="workspaces.length === 0" class="px-2 py-1.5 text-sm text-muted-foreground">
          No workspaces yet
        </div>

        <DropdownMenuItem
          v-for="ws in workspaces"
          :key="ws.id"
          class="cursor-pointer"
          @click="selectWorkspace(ws)"
        >
          <span
            class="mr-2 flex size-6 shrink-0 items-center justify-center rounded bg-muted font-mono text-[10px] font-semibold"
          >
            {{ ws.workspace_key.slice(0, 2) }}
          </span>
          <span class="flex-1 truncate">{{ ws.name }}</span>
          <Check
            v-if="ws.id === currentWorkspace?.id"
            class="ml-2 size-4 text-amber-600 dark:text-amber-500"
          />
        </DropdownMenuItem>

        <template v-if="isAdmin">
          <DropdownMenuSeparator />
          <DropdownMenuItem class="cursor-pointer" @click="showCreate = true">
            <Plus class="mr-2 size-4" />
            <span>Create workspace</span>
          </DropdownMenuItem>
          <DropdownMenuItem as-child>
            <NuxtLink to="/workspaces" class="flex cursor-pointer items-center">
              <Settings2 class="mr-2 size-4" />
              <span>Manage workspaces</span>
            </NuxtLink>
          </DropdownMenuItem>
        </template>
      </DropdownMenuContent>
    </DropdownMenu>

    <CreateWorkspaceDialog v-model:open="showCreate" />
  </div>
</template>
