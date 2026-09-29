<script setup lang="ts">
import { Users, Plus, Trash2, Loader2, ChevronLeft, ChevronRight, Shield, ShieldOff, KeyRound, Search, X, Merge, Check } from "lucide-vue-next";
import type { MergeUserResult } from "~/composables/useAdmin";

definePageMeta({
  middleware: ["admin"],
});

useSeoMeta({ title: "Manage Users" });

const { listUsers, createUser, deleteUser, updateUserRole, resetUserPassword, mergeUser } = useAdmin();

interface User {
  id: string;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  user_type: string;
  created_at: string;
}

// State
const users = ref<User[]>([]);
const loading = ref(true);
const page = ref(1);
const perPage = ref(20);
const total = ref(0);
const totalPages = ref(0);
const error = ref<string | null>(null);
const searchQuery = ref("");
let searchDebounce: ReturnType<typeof setTimeout> | null = null;

// Create dialog state
const showCreateDialog = ref(false);
const createLoading = ref(false);
const createError = ref<string | null>(null);
const createForm = ref({
  username: "",
  email: "",
  password: "",
  first_name: "",
  last_name: "",
  user_type: "user",
});

// Delete dialog state
const showDeleteDialog = ref(false);
const deleteLoading = ref(false);
const userToDelete = ref<User | null>(null);

// Role toggle dialog state
const showRoleDialog = ref(false);
const roleLoading = ref(false);
const userToToggleRole = ref<User | null>(null);

// Password reset dialog state
const showPasswordDialog = ref(false);
const passwordLoading = ref(false);
const passwordError = ref<string | null>(null);
const userToResetPassword = ref<User | null>(null);
const newPassword = ref("");

// Merge dialog state
const showMergeDialog = ref(false);
const mergeLoading = ref(false);
const mergeError = ref<string | null>(null);
const mergeSource = ref<User | null>(null);
const mergeTarget = ref<User | null>(null);
const mergeQuery = ref("");
const mergeCandidates = ref<User[]>([]);
const mergeResult = ref<MergeUserResult | null>(null);
let mergeSearchDebounce: ReturnType<typeof setTimeout> | null = null;

async function fetchUsers() {
  loading.value = true;
  error.value = null;
  const result = await listUsers(page.value, perPage.value, searchQuery.value);
  if (result.success && result.data) {
    users.value = result.data.users || [];
    total.value = result.data.total;
    totalPages.value = result.data.total_pages;
  } else {
    error.value = result.error || "Failed to fetch users";
  }
  loading.value = false;
}

function resetCreateForm() {
  createForm.value = {
    username: "",
    email: "",
    password: "",
    first_name: "",
    last_name: "",
    user_type: "user",
  };
  createError.value = null;
}

async function handleCreateUser() {
  createLoading.value = true;
  createError.value = null;
  const result = await createUser(createForm.value);
  createLoading.value = false;

  if (result.success) {
    showCreateDialog.value = false;
    resetCreateForm();
    await fetchUsers();
  } else {
    createError.value = result.error || "Failed to create user";
  }
}

function confirmDelete(user: User) {
  userToDelete.value = user;
  showDeleteDialog.value = true;
}

async function handleDeleteUser() {
  if (!userToDelete.value) return;

  deleteLoading.value = true;
  const result = await deleteUser(userToDelete.value.id);
  deleteLoading.value = false;

  if (result.success) {
    showDeleteDialog.value = false;
    userToDelete.value = null;
    await fetchUsers();
  } else {
    error.value = result.error || "Failed to delete user";
  }
}

function confirmToggleRole(user: User) {
  userToToggleRole.value = user;
  showRoleDialog.value = true;
}

async function handleToggleRole() {
  if (!userToToggleRole.value) return;

  roleLoading.value = true;
  const newType = userToToggleRole.value.user_type === "admin" ? "user" : "admin";
  const result = await updateUserRole(userToToggleRole.value.id, newType);
  roleLoading.value = false;

  if (result.success) {
    showRoleDialog.value = false;
    userToToggleRole.value = null;
    await fetchUsers();
  } else {
    error.value = result.error || "Failed to update role";
    showRoleDialog.value = false;
  }
}

function openPasswordReset(user: User) {
  userToResetPassword.value = user;
  newPassword.value = "";
  passwordError.value = null;
  showPasswordDialog.value = true;
}

async function handleResetPassword() {
  if (!userToResetPassword.value) return;

  passwordLoading.value = true;
  passwordError.value = null;
  const result = await resetUserPassword(userToResetPassword.value.id, newPassword.value);
  passwordLoading.value = false;

  if (result.success) {
    showPasswordDialog.value = false;
    userToResetPassword.value = null;
    newPassword.value = "";
  } else {
    passwordError.value = result.error || "Failed to reset password";
  }
}

function openMerge(user: User) {
  mergeSource.value = user;
  mergeTarget.value = null;
  mergeQuery.value = "";
  mergeCandidates.value = [];
  mergeResult.value = null;
  mergeError.value = null;
  showMergeDialog.value = true;
}

function onMergeSearchInput() {
  if (mergeSearchDebounce) clearTimeout(mergeSearchDebounce);
  mergeSearchDebounce = setTimeout(async () => {
    if (!mergeQuery.value.trim()) {
      mergeCandidates.value = [];
      return;
    }
    const result = await listUsers(1, 10, mergeQuery.value);
    if (result.success && result.data) {
      mergeCandidates.value = (result.data.users || []).filter((u) => u.id !== mergeSource.value?.id);
    }
  }, 300);
}

async function handleMerge() {
  if (!mergeSource.value || !mergeTarget.value) return;

  mergeLoading.value = true;
  mergeError.value = null;
  const result = await mergeUser(mergeSource.value.id, mergeTarget.value.id);
  mergeLoading.value = false;

  if (result.success && result.data) {
    mergeResult.value = result.data;
  } else {
    mergeError.value = result.error || "Failed to merge user";
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

// Search
function onSearchInput() {
  if (searchDebounce) clearTimeout(searchDebounce);
  searchDebounce = setTimeout(() => {
    page.value = 1;
    fetchUsers();
  }, 300);
}

function clearSearch() {
  searchQuery.value = "";
  page.value = 1;
  fetchUsers();
}

// Pagination
function prevPage() {
  if (page.value > 1) {
    page.value--;
    fetchUsers();
  }
}

function nextPage() {
  if (page.value < totalPages.value) {
    page.value++;
    fetchUsers();
  }
}

onMounted(() => {
  fetchUsers();
});
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main id="main-content" class="flex-1">
      <div class="mx-auto max-w-6xl px-6 py-12">
        <div class="mb-8 flex items-center justify-between">
          <div>
            <h1 class="flex items-center gap-2 text-3xl font-bold tracking-tight">
              <Users class="size-8" />
              User Management
            </h1>
            <p class="mt-2 text-muted-foreground">
              Manage all users in the system
            </p>
          </div>
          <Button @click="showCreateDialog = true; resetCreateForm()">
            <Plus class="mr-2 size-4" />
            Create User
          </Button>
        </div>

        <div v-if="error" role="alert" class="mb-4 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
          {{ error }}
        </div>

        <div class="mb-4 relative">
          <Search class="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            v-model="searchQuery"
            placeholder="Search by username, email, or name..."
            class="pl-9 pr-9"
            @input="onSearchInput"
          />
          <button
            v-if="searchQuery"
            aria-label="Clear search"
            class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 rounded-sm outline-none"
            @click="clearSearch"
          >
            <X class="size-4" />
          </button>
        </div>

        <Card>
          <CardContent class="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Username</TableHead>
                  <TableHead>Email</TableHead>
                  <TableHead>Name</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead class="w-[140px]">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-if="loading">
                  <TableCell colspan="6" class="py-8 text-center">
                    <Loader2 class="mx-auto size-6 animate-spin" />
                  </TableCell>
                </TableRow>
                <TableRow v-else-if="users.length === 0">
                  <TableCell colspan="6" class="py-8 text-center text-muted-foreground">
                    No users found
                  </TableCell>
                </TableRow>
                <TableRow v-for="user in users" :key="user.id">
                  <TableCell class="font-medium">{{ user.username }}</TableCell>
                  <TableCell>{{ user.email }}</TableCell>
                  <TableCell>{{ user.first_name }} {{ user.last_name }}</TableCell>
                  <TableCell>
                    <Badge :variant="user.user_type === 'admin' ? 'default' : 'secondary'">
                      {{ user.user_type }}
                    </Badge>
                  </TableCell>
                  <TableCell>{{ formatDate(user.created_at) }}</TableCell>
                  <TableCell>
                    <div class="flex items-center gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        :aria-label="user.user_type === 'admin' ? 'Demote to user' : 'Promote to admin'"
                        :title="user.user_type === 'admin' ? 'Demote to user' : 'Promote to admin'"
                        @click="confirmToggleRole(user)"
                      >
                        <Shield v-if="user.user_type !== 'admin'" class="size-4" />
                        <ShieldOff v-else class="size-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        aria-label="Reset password"
                        title="Reset password"
                        @click="openPasswordReset(user)"
                      >
                        <KeyRound class="size-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        aria-label="Merge into another user"
                        title="Merge into another user"
                        @click="openMerge(user)"
                      >
                        <Merge class="size-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        aria-label="Delete user"
                        class="text-destructive hover:text-destructive"
                        @click="confirmDelete(user)"
                      >
                        <Trash2 class="size-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </CardContent>
          <CardFooter class="flex items-center justify-between border-t px-6 py-4">
            <p class="text-sm text-muted-foreground">
              Showing {{ users.length }} of {{ total }} users
            </p>
            <div class="flex items-center gap-2">
              <Button variant="outline" size="sm" aria-label="Previous page" :disabled="page === 1" @click="prevPage">
                <ChevronLeft class="size-4" />
              </Button>
              <span class="text-sm">Page {{ page }} of {{ totalPages || 1 }}</span>
              <Button variant="outline" size="sm" aria-label="Next page" :disabled="page >= totalPages" @click="nextPage">
                <ChevronRight class="size-4" />
              </Button>
            </div>
          </CardFooter>
        </Card>

        <!-- Create User Dialog -->
        <Dialog v-model:open="showCreateDialog">
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New User</DialogTitle>
              <DialogDescription>
                Fill in the details to create a new user account.
              </DialogDescription>
            </DialogHeader>
            <form @submit.prevent="handleCreateUser" class="space-y-4">
              <div v-if="createError" role="alert" class="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                {{ createError }}
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div class="space-y-2">
                  <Label for="first_name">First Name</Label>
                  <Input id="first_name" v-model="createForm.first_name" required :disabled="createLoading" />
                </div>
                <div class="space-y-2">
                  <Label for="last_name">Last Name</Label>
                  <Input id="last_name" v-model="createForm.last_name" required :disabled="createLoading" />
                </div>
              </div>
              <div class="space-y-2">
                <Label for="username">Username</Label>
                <Input id="username" v-model="createForm.username" required :disabled="createLoading" />
              </div>
              <div class="space-y-2">
                <Label for="email">Email</Label>
                <Input id="email" type="email" v-model="createForm.email" required :disabled="createLoading" />
              </div>
              <div class="space-y-2">
                <Label for="password">Password</Label>
                <Input id="password" type="password" v-model="createForm.password" required :disabled="createLoading" />
              </div>
              <div class="space-y-2">
                <Label for="user_type">User Type</Label>
                <NativeSelect id="user_type" v-model="createForm.user_type" :disabled="createLoading">
                  <option value="user">User</option>
                  <option value="admin">Admin</option>
                </NativeSelect>
              </div>
              <DialogFooter>
                <Button type="button" variant="outline" @click="showCreateDialog = false" :disabled="createLoading">
                  Cancel
                </Button>
                <Button type="submit" :disabled="createLoading">
                  <Loader2 v-if="createLoading" class="mr-2 size-4 animate-spin" />
                  Create
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>

        <!-- Toggle Role Dialog -->
        <Dialog v-model:open="showRoleDialog">
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Change User Role</DialogTitle>
              <DialogDescription>
                <template v-if="userToToggleRole?.user_type === 'admin'">
                  Demote "{{ userToToggleRole?.username }}" from admin to regular user?
                  They will lose access to admin features.
                </template>
                <template v-else>
                  Promote "{{ userToToggleRole?.username }}" to admin?
                  They will gain access to all admin features.
                </template>
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button variant="outline" @click="showRoleDialog = false" :disabled="roleLoading">
                Cancel
              </Button>
              <Button :disabled="roleLoading" @click="handleToggleRole">
                <Loader2 v-if="roleLoading" class="mr-2 size-4 animate-spin" />
                {{ userToToggleRole?.user_type === 'admin' ? 'Demote to User' : 'Promote to Admin' }}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <!-- Reset Password Dialog -->
        <Dialog v-model:open="showPasswordDialog">
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Reset Password</DialogTitle>
              <DialogDescription>
                Set a new password for "{{ userToResetPassword?.username }}".
                This will revoke all their active sessions.
              </DialogDescription>
            </DialogHeader>
            <form @submit.prevent="handleResetPassword" class="space-y-4">
              <div v-if="passwordError" role="alert" class="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                {{ passwordError }}
              </div>
              <div class="space-y-2">
                <Label for="new_password">New Password</Label>
                <Input
                  id="new_password"
                  type="password"
                  v-model="newPassword"
                  required
                  minlength="8"
                  placeholder="Minimum 8 characters"
                  :disabled="passwordLoading"
                />
              </div>
              <DialogFooter>
                <Button type="button" variant="outline" @click="showPasswordDialog = false" :disabled="passwordLoading">
                  Cancel
                </Button>
                <Button type="submit" :disabled="passwordLoading">
                  <Loader2 v-if="passwordLoading" class="mr-2 size-4 animate-spin" />
                  Reset Password
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>

        <!-- Merge User Dialog -->
        <Dialog v-model:open="showMergeDialog">
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Merge User</DialogTitle>
              <DialogDescription>
                Give another user everything "{{ mergeSource?.username }}" has: workspaces, projects
                (higher role wins), task assignments, modules and private views. They will also follow
                tasks "{{ mergeSource?.username }}" created, follows or commented on.
                "{{ mergeSource?.username }}" is left unchanged. No notifications are sent.
              </DialogDescription>
            </DialogHeader>

            <div v-if="mergeResult" class="space-y-2 text-sm">
              <p class="font-medium">
                Merged "{{ mergeSource?.username }}" into "{{ mergeTarget?.username }}".
              </p>
              <ul class="space-y-1 text-muted-foreground">
                <li>Workspaces added: {{ mergeResult.workspaces }}</li>
                <li>Projects added or upgraded: {{ mergeResult.projects }}</li>
                <li>Task assignments added: {{ mergeResult.assignments }}</li>
                <li>Modules added: {{ mergeResult.modules }}</li>
                <li>Tasks followed: {{ mergeResult.followed_tasks }}</li>
                <li>Views copied: {{ mergeResult.views }}</li>
              </ul>
            </div>

            <div v-else class="space-y-3">
              <div v-if="mergeError" role="alert" class="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                {{ mergeError }}
              </div>
              <div class="space-y-2">
                <Label for="merge_target">Merge into</Label>
                <Input
                  id="merge_target"
                  v-model="mergeQuery"
                  placeholder="Search by username, email, or name..."
                  autocomplete="off"
                  :disabled="mergeLoading"
                  @input="onMergeSearchInput"
                />
              </div>
              <div v-if="mergeCandidates.length" class="max-h-56 overflow-y-auto rounded-md border">
                <button
                  v-for="candidate in mergeCandidates"
                  :key="candidate.id"
                  type="button"
                  class="flex w-full items-center justify-between gap-2 px-3 py-2 text-left text-sm hover:bg-accent focus-visible:bg-accent outline-none"
                  :aria-pressed="mergeTarget?.id === candidate.id"
                  @click="mergeTarget = candidate"
                >
                  <span class="min-w-0 truncate">
                    <span class="font-medium">{{ candidate.username }}</span>
                    <span class="text-muted-foreground"> · {{ candidate.email }}</span>
                  </span>
                  <Check v-if="mergeTarget?.id === candidate.id" class="size-4 shrink-0" />
                </button>
              </div>
              <p v-else-if="mergeQuery.trim()" class="text-sm text-muted-foreground">No matching users</p>
            </div>

            <DialogFooter>
              <template v-if="mergeResult">
                <Button @click="showMergeDialog = false">Done</Button>
              </template>
              <template v-else>
                <Button variant="outline" @click="showMergeDialog = false" :disabled="mergeLoading">
                  Cancel
                </Button>
                <Button :disabled="mergeLoading || !mergeTarget" @click="handleMerge">
                  <Loader2 v-if="mergeLoading" class="mr-2 size-4 animate-spin" />
                  {{ mergeTarget ? `Merge into ${mergeTarget.username}` : 'Merge' }}
                </Button>
              </template>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <!-- Delete Confirmation Dialog -->
        <Dialog v-model:open="showDeleteDialog">
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Delete User</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete the user "{{ userToDelete?.username }}"?
                This action cannot be undone.
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button variant="outline" @click="showDeleteDialog = false" :disabled="deleteLoading">
                Cancel
              </Button>
              <Button variant="destructive" :disabled="deleteLoading" @click="handleDeleteUser">
                <Loader2 v-if="deleteLoading" class="mr-2 size-4 animate-spin" />
                Delete
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </main>
  </div>
</template>
