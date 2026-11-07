<script setup lang="ts">
import { Plus, Building2, Pencil, Trash2, Users, Loader2, X } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { Workspace, WorkspaceMember } from "~/types";

// Minimal shape for the add-member picker (from the admin user directory).
interface DirectoryUser {
  id: string;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
}

definePageMeta({
  middleware: ["auth", "admin"],
});

useSeoMeta({ title: "Workspaces" });

const {
  workspaces,
  loading,
  listWorkspaces,
  updateWorkspace,
  deleteWorkspace,
  listMembers,
  members,
  addMember,
  removeMember,
} = useWorkspaces();
const { listUsers } = useAdmin();

const showCreate = ref(false);

onMounted(() => {
  listWorkspaces();
});
onDataChanged(listWorkspaces);

// --- Edit ---------------------------------------------------------------
const editing = ref<Workspace | null>(null);
const editForm = ref({ name: "", description: "" });
const editLoading = ref(false);

function openEdit(ws: Workspace) {
  editing.value = ws;
  editForm.value = { name: ws.name, description: ws.description || "" };
}

async function saveEdit() {
  if (!editing.value) return;
  editLoading.value = true;
  const result = await updateWorkspace(editing.value.workspace_key, {
    name: editForm.value.name,
    description: editForm.value.description || undefined,
  });
  editLoading.value = false;
  if (result.success) {
    toast.success("Workspace updated");
    editing.value = null;
  } else {
    toast.error(result.error || "Failed to update workspace");
  }
}

// --- Delete -------------------------------------------------------------
const deleting = ref<Workspace | null>(null);
const deleteLoading = ref(false);

async function confirmDelete() {
  if (!deleting.value) return;
  deleteLoading.value = true;
  const result = await deleteWorkspace(deleting.value.workspace_key);
  deleteLoading.value = false;
  if (result.success) {
    toast.success("Workspace deleted");
    deleting.value = null;
  } else {
    toast.error(result.error || "Failed to delete workspace");
  }
}

// --- Members ------------------------------------------------------------
const managing = ref<Workspace | null>(null);
const addOpen = ref(false);
const allUsers = ref<DirectoryUser[]>([]);

// Users not already members — the pool the add picker searches over.
const availableUsers = computed(() => {
  const existing = new Set(members.value.map((m) => m.user_id));
  return allUsers.value.filter((u) => !existing.has(u.id));
});

function userSearchText(u: DirectoryUser) {
  return `${u.first_name} ${u.last_name} ${u.username} ${u.email}`;
}

// Load the full user directory once so the SearchableSelect picker can filter
// client-side (same UX as the task-assignee picker). Pages through the admin
// user list; capped as a safety valve.
async function loadAllUsers() {
  const collected: DirectoryUser[] = [];
  for (let page = 1; page <= 50; page++) {
    const res = await listUsers(page, 100);
    if (!res.success || !res.data) break;
    collected.push(...res.data.users);
    if (page >= res.data.total_pages) break;
  }
  allUsers.value = collected;
}

async function openMembers(ws: Workspace) {
  managing.value = ws;
  addOpen.value = false;
  await Promise.all([listMembers(ws.workspace_key), loadAllUsers()]);
}

async function handleAddMember(userId: string) {
  if (!managing.value) return;
  const result = await addMember(managing.value.workspace_key, { user_id: userId });
  if (result.success) {
    await listMembers(managing.value.workspace_key);
    toast.success("Member added");
  } else {
    toast.error(result.error || "Failed to add member");
  }
}

async function handleRemoveMember(member: WorkspaceMember) {
  if (!managing.value) return;
  const result = await removeMember(managing.value.workspace_key, member.user_id);
  if (result.success) {
    await listMembers(managing.value.workspace_key);
    toast.success("Member removed");
  } else {
    toast.error(result.error || "Failed to remove member");
  }
}
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main id="main-content" class="flex-1">
      <div class="mx-auto max-w-5xl px-6 py-8">
        <div class="mb-8 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 class="flex items-center gap-3 text-3xl font-bold tracking-tight">
              <Building2 class="size-8" />
              Workspaces
            </h1>
            <p class="mt-2 text-muted-foreground">
              Group projects into workspaces and manage who can see them.
            </p>
          </div>
          <Button @click="showCreate = true">
            <Plus class="mr-2 size-4" />
            Create Workspace
          </Button>
        </div>

        <div v-if="loading && workspaces.length === 0" class="flex items-center justify-center py-12">
          <Loader2 class="size-8 animate-spin text-muted-foreground" />
        </div>

        <div
          v-else-if="workspaces.length === 0"
          class="flex flex-col items-center justify-center rounded-lg border border-dashed py-16"
        >
          <div class="flex size-16 items-center justify-center rounded-full bg-muted">
            <Building2 class="size-8 text-muted-foreground" />
          </div>
          <h3 class="mt-4 text-lg font-semibold">No workspaces yet</h3>
          <Button class="mt-4" @click="showCreate = true">
            <Plus class="mr-2 size-4" />
            Create Workspace
          </Button>
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="ws in workspaces"
            :key="ws.id"
            class="flex items-center gap-4 rounded-lg border p-4"
          >
            <span
              class="flex size-10 shrink-0 items-center justify-center rounded-md bg-amber-500/10 font-mono text-sm font-semibold text-amber-700 dark:text-amber-400"
            >
              {{ ws.workspace_key.slice(0, 2) }}
            </span>
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <h3 class="truncate font-semibold">{{ ws.name }}</h3>
                <span class="rounded bg-muted px-1.5 py-0.5 font-mono text-xs text-muted-foreground">
                  {{ ws.workspace_key }}
                </span>
              </div>
              <p v-if="ws.description" class="truncate text-sm text-muted-foreground">
                {{ ws.description }}
              </p>
            </div>
            <div class="flex items-center gap-1">
              <Button variant="ghost" size="sm" @click="openMembers(ws)">
                <Users class="mr-1 size-4" />
                Members
              </Button>
              <Button variant="ghost" size="icon" aria-label="Edit" @click="openEdit(ws)">
                <Pencil class="size-4" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                aria-label="Delete"
                class="text-destructive hover:text-destructive"
                @click="deleting = ws"
              >
                <Trash2 class="size-4" />
              </Button>
            </div>
          </div>
        </div>
      </div>
    </main>

    <CreateWorkspaceDialog v-model:open="showCreate" @created="listWorkspaces" />

    <!-- Edit dialog -->
    <Dialog :open="!!editing" @update:open="(v) => { if (!v) editing = null; }">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Edit Workspace</DialogTitle>
        </DialogHeader>
        <form class="space-y-4" @submit.prevent="saveEdit">
          <div class="space-y-2">
            <Label for="edit_name">Name</Label>
            <Input id="edit_name" v-model="editForm.name" required :disabled="editLoading" />
          </div>
          <div class="space-y-2">
            <Label for="edit_desc">Description</Label>
            <Textarea id="edit_desc" v-model="editForm.description" rows="3" :disabled="editLoading" />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" :disabled="editLoading" @click="editing = null">
              Cancel
            </Button>
            <Button type="submit" :disabled="editLoading || !editForm.name">
              <Loader2 v-if="editLoading" class="mr-2 size-4 animate-spin" />
              Save
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <!-- Delete confirm -->
    <Dialog :open="!!deleting" @update:open="(v) => { if (!v) deleting = null; }">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Delete workspace?</DialogTitle>
          <DialogDescription>
            "{{ deleting?.name }}" and its association with projects will be removed. This cannot be undone.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button type="button" variant="outline" :disabled="deleteLoading" @click="deleting = null">
            Cancel
          </Button>
          <Button variant="destructive" :disabled="deleteLoading" @click="confirmDelete">
            <Loader2 v-if="deleteLoading" class="mr-2 size-4 animate-spin" />
            Delete
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Members dialog -->
    <Dialog :open="!!managing" @update:open="(v) => { if (!v) managing = null; }">
      <DialogContent class="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Members — {{ managing?.name }}</DialogTitle>
          <DialogDescription>
            Members can see this workspace and its projects.
          </DialogDescription>
        </DialogHeader>

        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <p class="text-xs font-medium text-muted-foreground">
              {{ members.length }} member{{ members.length === 1 ? "" : "s" }}
            </p>

            <!-- Add member — same searchable picker as task assignees -->
            <SearchableSelect
              v-model:open="addOpen"
              :items="availableUsers"
              :get-search-text="userSearchText"
              :get-key="(u) => u.id"
              :close-on-select="false"
              align="end"
              content-class="w-72"
              placeholder="Search users to add..."
              empty-text="No users found"
              @select="(u) => handleAddMember(u.id)"
            >
              <template #trigger>
                <Button variant="outline" size="sm" class="h-8 gap-1.5">
                  <Plus class="size-3.5" />
                  Add member
                </Button>
              </template>
              <template #option="{ item: u }">
                <Avatar class="size-6">
                  <AvatarFallback class="text-xs" :seed="u.id">
                    {{ u.first_name?.[0] }}{{ u.last_name?.[0] }}
                  </AvatarFallback>
                </Avatar>
                <span class="truncate">
                  {{ u.first_name }} {{ u.last_name }}
                  <span class="text-muted-foreground">@{{ u.username }}</span>
                </span>
              </template>
            </SearchableSelect>
          </div>

          <!-- Member list — avatar chips (task-assignee pattern), scrollable so
               a large membership never pushes the dialog past the viewport. -->
          <div class="flex max-h-[45vh] flex-wrap content-start gap-2 overflow-y-auto pr-1">
            <div
              v-for="m in members"
              :key="m.id"
              class="group relative flex items-center gap-1.5 rounded-md border bg-muted/50 py-1 pl-1 pr-2.5"
            >
              <NuxtLink
                :to="`/profile/${m.user_id}`"
                class="flex items-center gap-1.5 transition-opacity hover:opacity-80"
              >
                <Avatar class="size-6">
                  <AvatarImage v-if="m.avatar_url" :src="m.avatar_url" />
                  <AvatarFallback class="text-xs" :seed="m.user_id">
                    {{ m.first_name?.[0] }}{{ m.last_name?.[0] }}
                  </AvatarFallback>
                </Avatar>
                <span class="text-sm">{{ m.first_name }} {{ m.last_name }}</span>
              </NuxtLink>
              <button
                type="button"
                :aria-label="`Remove ${m.first_name} ${m.last_name}`"
                class="absolute -right-1.5 -top-1.5 flex size-4 items-center justify-center rounded-full bg-foreground text-background opacity-0 shadow-sm outline-none transition-opacity focus:opacity-100 focus-visible:ring-2 focus-visible:ring-ring group-hover:opacity-100"
                @click="handleRemoveMember(m)"
              >
                <X class="size-2.5" />
              </button>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>
