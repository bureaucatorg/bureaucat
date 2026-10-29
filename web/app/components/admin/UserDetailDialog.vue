<script setup lang="ts">
import { Loader2 } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { AdminUserDetail, AdminUserUpdate } from "~/composables/useAdmin";

const props = defineProps<{
  userId: string | null;
}>();

const open = defineModel<boolean>("open", { default: false });

const emit = defineEmits<{
  updated: [user: AdminUserDetail];
}>();

const { user: currentUser } = useAuth();
const { getUserDetail, updateUser } = useAdmin();

const detail = ref<AdminUserDetail | null>(null);
const loading = ref(false);
const saving = ref(false);
const error = ref<string | null>(null);
const form = ref<AdminUserUpdate>({
  first_name: "",
  last_name: "",
  user_type: "user",
  email_notifications: false,
});

const isSelf = computed(() => detail.value?.id === currentUser.value?.id);

const dirty = computed(() => {
  const d = detail.value;
  if (!d) return false;
  return (
    form.value.first_name.trim() !== d.first_name ||
    form.value.last_name.trim() !== d.last_name ||
    form.value.user_type !== d.user_type ||
    form.value.email_notifications !== d.email_notifications
  );
});

const signInMethod = computed(() => {
  const d = detail.value;
  if (!d) return "";
  if (d.auth_provider && d.has_password) return `SSO (${d.auth_provider}) + password`;
  if (d.auth_provider) return `SSO (${d.auth_provider})`;
  return d.has_password ? "Password" : "None";
});

function setDetail(d: AdminUserDetail) {
  detail.value = d;
  form.value = {
    first_name: d.first_name,
    last_name: d.last_name,
    user_type: d.user_type,
    email_notifications: d.email_notifications,
  };
}

async function load() {
  if (!props.userId) return;
  loading.value = true;
  error.value = null;
  detail.value = null;
  const result = await getUserDetail(props.userId);
  if (result.success && result.data) {
    setDetail(result.data);
  } else {
    error.value = result.error || "Failed to load user";
  }
  loading.value = false;
}

async function save() {
  if (!detail.value) return;
  saving.value = true;
  error.value = null;
  const result = await updateUser(detail.value.id, {
    ...form.value,
    first_name: form.value.first_name.trim(),
    last_name: form.value.last_name.trim(),
  });
  saving.value = false;
  if (result.success && result.data) {
    setDetail(result.data);
    emit("updated", result.data);
    toast.success("User updated");
  } else {
    error.value = result.error || "Failed to update user";
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

watch(
  () => [open.value, props.userId] as const,
  ([isOpen]) => {
    if (isOpen) load();
  }
);
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent class="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
      <DialogHeader>
        <DialogTitle>User details</DialogTitle>
        <DialogDescription>
          Username, email and creation date can't be changed.
        </DialogDescription>
      </DialogHeader>

      <div v-if="loading" class="flex items-center justify-center py-12">
        <Loader2 class="size-6 animate-spin text-muted-foreground" />
      </div>

      <div
        v-else-if="!detail && error"
        role="alert"
        class="rounded-md bg-destructive/10 p-3 text-sm text-destructive"
      >
        {{ error }}
      </div>

      <form v-else-if="detail" class="space-y-6" @submit.prevent="save">
        <div class="flex items-center gap-3">
          <Avatar class="size-12">
            <AvatarImage v-if="detail.avatar_url" :src="detail.avatar_url" />
            <AvatarFallback :seed="detail.id">
              {{ detail.first_name[0] }}{{ detail.last_name[0] }}
            </AvatarFallback>
          </Avatar>
          <div class="min-w-0">
            <p class="truncate font-semibold">@{{ detail.username }}</p>
            <p class="truncate text-sm text-muted-foreground">{{ detail.email }}</p>
          </div>
        </div>

        <div v-if="error" role="alert" class="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
          {{ error }}
        </div>

        <dl class="grid grid-cols-1 gap-x-6 gap-y-3 text-sm sm:grid-cols-2">
          <div>
            <dt class="text-xs text-muted-foreground">Sign-in method</dt>
            <dd>{{ signInMethod }}</dd>
          </div>
          <div>
            <dt class="text-xs text-muted-foreground">Active sessions / API tokens</dt>
            <dd>{{ detail.active_sessions }} / {{ detail.active_tokens }}</dd>
          </div>
          <div>
            <dt class="text-xs text-muted-foreground">Created</dt>
            <dd>{{ formatDate(detail.created_at) }}</dd>
          </div>
          <div>
            <dt class="text-xs text-muted-foreground">Last updated</dt>
            <dd>{{ formatDate(detail.updated_at) }}</dd>
          </div>
        </dl>

        <Separator />

        <div class="space-y-4">
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div class="space-y-2">
              <Label for="detail_first_name">First name</Label>
              <Input id="detail_first_name" v-model="form.first_name" maxlength="100" required :disabled="saving" />
            </div>
            <div class="space-y-2">
              <Label for="detail_last_name">Last name</Label>
              <Input id="detail_last_name" v-model="form.last_name" maxlength="100" required :disabled="saving" />
            </div>
          </div>
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div class="space-y-2">
              <Label for="detail_user_type">Role</Label>
              <NativeSelect id="detail_user_type" v-model="form.user_type" :disabled="saving || isSelf">
                <option value="user">User</option>
                <option value="admin">Admin</option>
              </NativeSelect>
              <p v-if="isSelf" class="text-xs text-muted-foreground">You can't change your own role.</p>
            </div>
            <div class="flex items-center justify-between gap-4 rounded-md border px-3 py-2 sm:mt-6 sm:self-start">
              <Label for="detail_email_notifications" class="cursor-pointer">Email notifications</Label>
              <Switch
                id="detail_email_notifications"
                :checked="form.email_notifications"
                :disabled="saving"
                @update:checked="form.email_notifications = $event"
              />
            </div>
          </div>
        </div>

        <Separator />

        <div class="space-y-2">
          <h4 class="text-sm font-medium">
            Workspaces <span class="font-normal text-muted-foreground">({{ detail.workspaces.length }})</span>
          </h4>
          <p v-if="detail.workspaces.length === 0" class="text-sm text-muted-foreground">None</p>
          <div v-else class="flex flex-wrap gap-1.5">
            <Badge v-for="w in detail.workspaces" :key="w.workspace_key" variant="secondary">
              {{ w.name }}
            </Badge>
          </div>
        </div>

        <div class="space-y-2">
          <h4 class="text-sm font-medium">
            Projects <span class="font-normal text-muted-foreground">({{ detail.projects.length }})</span>
          </h4>
          <p v-if="detail.projects.length === 0" class="text-sm text-muted-foreground">None</p>
          <ul v-else class="max-h-56 divide-y overflow-y-auto rounded-md border">
            <li
              v-for="p in detail.projects"
              :key="p.project_key"
              class="flex items-center gap-2 px-3 py-2 text-sm"
            >
              <span class="shrink-0 font-mono text-xs text-amber-600 dark:text-amber-500">{{ p.project_key }}</span>
              <span class="min-w-0 flex-1 truncate">{{ p.name }}</span>
              <span v-if="p.workspace_name" class="hidden truncate text-xs text-muted-foreground sm:inline">
                {{ p.workspace_name }}
              </span>
              <Badge :variant="p.role === 'admin' ? 'default' : 'outline'" class="shrink-0">{{ p.role }}</Badge>
            </li>
          </ul>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" :disabled="saving" @click="open = false">Close</Button>
          <Button type="submit" :disabled="saving || !dirty">
            <Loader2 v-if="saving" class="mr-2 size-4 animate-spin" />
            Save changes
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
