<script setup lang="ts">
import { Users, Plus, Trash2, Loader2, ChevronLeft, ChevronRight } from "lucide-vue-next";

definePageMeta({
  middleware: ["admin"],
});

useSeoMeta({ title: "Manage Users" });

const { listUsers, createUser, deleteUser } = useAdmin();

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

async function fetchUsers() {
  loading.value = true;
  error.value = null;
  const result = await listUsers(page.value, perPage.value);
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

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
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

    <main class="flex-1">
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

        <div v-if="error" class="mb-4 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
          {{ error }}
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
                  <TableHead class="w-[100px]">Actions</TableHead>
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
                    <Button
                      variant="ghost"
                      size="icon"
                      class="text-destructive hover:text-destructive"
                      @click="confirmDelete(user)"
                    >
                      <Trash2 class="size-4" />
                    </Button>
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
              <Button variant="outline" size="sm" :disabled="page === 1" @click="prevPage">
                <ChevronLeft class="size-4" />
              </Button>
              <span class="text-sm">Page {{ page }} of {{ totalPages || 1 }}</span>
              <Button variant="outline" size="sm" :disabled="page >= totalPages" @click="nextPage">
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
              <div v-if="createError" class="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
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
