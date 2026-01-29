<script setup lang="ts">
import { Loader2, Search, UserPlus } from "lucide-vue-next";
import { toast } from "vue-sonner";

const props = defineProps<{
  projectKey: string;
  existingMemberIds: string[];
}>();

const open = defineModel<boolean>("open", { default: false });

const emit = defineEmits<{
  added: [];
}>();

const { addMember } = useProjects();
const { listUsers } = useAdmin();

const loading = ref(false);
const searchLoading = ref(false);
const error = ref<string | null>(null);
const searchQuery = ref("");
const searchResults = ref<Array<{
  id: string;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
}>>([]);
const selectedUser = ref<string | null>(null);
const selectedRole = ref("member");

const roles = [
  { value: "guest", label: "Guest", description: "Read-only access" },
  { value: "member", label: "Member", description: "Can create and edit tasks" },
  { value: "admin", label: "Admin", description: "Full project control" },
];

async function searchUsers() {
  if (searchQuery.value.length < 2) {
    searchResults.value = [];
    return;
  }

  searchLoading.value = true;
  const result = await listUsers(1, 20);
  searchLoading.value = false;

  if (result.success && result.data) {
    // Filter out existing members and search by name/email
    const query = searchQuery.value.toLowerCase();
    searchResults.value = result.data.users.filter(
      (u) =>
        !props.existingMemberIds.includes(u.id) &&
        (u.username.toLowerCase().includes(query) ||
          u.email.toLowerCase().includes(query) ||
          u.first_name.toLowerCase().includes(query) ||
          u.last_name.toLowerCase().includes(query))
    );
  }
}

function resetForm() {
  searchQuery.value = "";
  searchResults.value = [];
  selectedUser.value = null;
  selectedRole.value = "member";
  error.value = null;
}

watch(open, (isOpen) => {
  if (isOpen) {
    resetForm();
  }
});

let searchTimeout: ReturnType<typeof setTimeout>;
watch(searchQuery, () => {
  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(searchUsers, 300);
});

async function handleSubmit() {
  if (!selectedUser.value) return;

  loading.value = true;
  error.value = null;

  const result = await addMember(props.projectKey, {
    user_id: selectedUser.value,
    role: selectedRole.value,
  });

  loading.value = false;

  if (result.success) {
    toast.success("Member added successfully");
    open.value = false;
    emit("added");
  } else {
    error.value = result.error || "Failed to add member";
  }
}

const selectedUserInfo = computed(() =>
  searchResults.value.find((u) => u.id === selectedUser.value)
);
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>Add Member</DialogTitle>
        <DialogDescription>
          Search for a user to add to this project.
        </DialogDescription>
      </DialogHeader>
      <form class="space-y-4" @submit.prevent="handleSubmit">
        <div
          v-if="error"
          class="rounded-md bg-destructive/10 p-3 text-sm text-destructive"
        >
          {{ error }}
        </div>

        <div class="space-y-2">
          <Label for="search">Search Users</Label>
          <div class="relative">
            <Search class="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              id="search"
              v-model="searchQuery"
              placeholder="Search by name or email..."
              class="pl-9"
              :disabled="loading"
            />
          </div>
        </div>

        <!-- Search results -->
        <div
          v-if="searchResults.length > 0 || searchLoading"
          class="max-h-48 space-y-1 overflow-auto rounded-lg border p-2"
        >
          <div
            v-if="searchLoading"
            class="flex items-center justify-center py-4"
          >
            <Loader2 class="size-4 animate-spin text-muted-foreground" />
          </div>
          <button
            v-for="u in searchResults"
            v-else
            :key="u.id"
            type="button"
            class="flex w-full items-center gap-3 rounded-md p-2 text-left transition-colors hover:bg-muted"
            :class="{ 'bg-muted': selectedUser === u.id }"
            @click="selectedUser = u.id"
          >
            <Avatar class="size-8">
              <AvatarFallback class="text-xs">
                {{ u.first_name[0] }}{{ u.last_name[0] }}
              </AvatarFallback>
            </Avatar>
            <div class="min-w-0 flex-1">
              <p class="truncate text-sm font-medium">
                {{ u.first_name }} {{ u.last_name }}
              </p>
              <p class="truncate text-xs text-muted-foreground">
                @{{ u.username }} · {{ u.email }}
              </p>
            </div>
          </button>
        </div>

        <!-- Selected user -->
        <div
          v-if="selectedUserInfo"
          class="flex items-center gap-3 rounded-lg border bg-muted/50 p-3"
        >
          <Avatar>
            <AvatarFallback>
              {{ selectedUserInfo.first_name[0] }}{{ selectedUserInfo.last_name[0] }}
            </AvatarFallback>
          </Avatar>
          <div>
            <p class="font-medium">
              {{ selectedUserInfo.first_name }} {{ selectedUserInfo.last_name }}
            </p>
            <p class="text-sm text-muted-foreground">
              @{{ selectedUserInfo.username }}
            </p>
          </div>
        </div>

        <div class="space-y-2">
          <Label>Role</Label>
          <div class="space-y-2">
            <label
              v-for="role in roles"
              :key="role.value"
              class="flex cursor-pointer items-center gap-3 rounded-lg border p-3 transition-colors hover:bg-muted/50"
              :class="{ 'border-primary bg-primary/5': selectedRole === role.value }"
            >
              <input
                v-model="selectedRole"
                type="radio"
                :value="role.value"
                class="sr-only"
                :disabled="loading"
              />
              <div
                class="size-4 rounded-full border-2"
                :class="{
                  'border-primary bg-primary': selectedRole === role.value,
                  'border-muted-foreground': selectedRole !== role.value,
                }"
              />
              <div>
                <p class="font-medium">{{ role.label }}</p>
                <p class="text-sm text-muted-foreground">{{ role.description }}</p>
              </div>
            </label>
          </div>
        </div>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            :disabled="loading"
            @click="open = false"
          >
            Cancel
          </Button>
          <Button type="submit" :disabled="loading || !selectedUser">
            <Loader2 v-if="loading" class="mr-2 size-4 animate-spin" />
            <UserPlus v-else class="mr-2 size-4" />
            Add Member
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
