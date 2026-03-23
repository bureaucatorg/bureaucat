<script setup lang="ts">
import { MoreHorizontal, Trash2, Shield, User, UserCog, Loader2 } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { ProjectMember } from "~/types";

const props = defineProps<{
  members: ProjectMember[];
  projectKey: string;
  currentUserRole: string;
  loading?: boolean;
}>();

const emit = defineEmits<{
  refresh: [];
}>();

const { updateMemberRole, removeMember } = useProjects();
const { user } = useAuth();

const processingId = ref<string | null>(null);

const isAdmin = computed(() => props.currentUserRole === "admin");

function getRoleIcon(role: string) {
  switch (role) {
    case "admin":
      return Shield;
    case "member":
      return User;
    default:
      return UserCog;
  }
}

async function handleRoleChange(member: ProjectMember, newRole: string) {
  if (!isAdmin.value || processingId.value) return;

  processingId.value = member.user_id;
  const result = await updateMemberRole(props.projectKey, member.user_id, { role: newRole });
  processingId.value = null;

  if (result.success) {
    toast.success(`Role updated to ${newRole}`);
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to update role");
  }
}

async function handleRemove(member: ProjectMember) {
  if (!isAdmin.value || processingId.value) return;
  if (member.user_id === user.value?.id) {
    toast.error("Cannot remove yourself");
    return;
  }

  processingId.value = member.user_id;
  const result = await removeMember(props.projectKey, member.user_id);
  processingId.value = null;

  if (result.success) {
    toast.success("Member removed");
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to remove member");
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}
</script>

<template>
  <div>
    <div v-if="loading" class="flex items-center justify-center py-8">
      <Loader2 class="size-6 animate-spin text-muted-foreground" />
    </div>

    <Table v-else>
      <TableHeader>
        <TableRow>
          <TableHead>Member</TableHead>
          <TableHead>Role</TableHead>
          <TableHead>Joined</TableHead>
          <TableHead v-if="isAdmin" class="w-[100px]">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-for="member in members" :key="member.id">
          <TableCell>
            <NuxtLink :to="`/profile/${member.user_id}`" class="flex items-center gap-3 hover:opacity-80 transition-opacity">
              <Avatar>
                <AvatarFallback>
                  {{ member.first_name[0] }}{{ member.last_name[0] }}
                </AvatarFallback>
              </Avatar>
              <div>
                <p class="font-medium">
                  {{ member.first_name }} {{ member.last_name }}
                </p>
                <p class="text-sm text-muted-foreground">@{{ member.username }}</p>
              </div>
            </NuxtLink>
          </TableCell>
          <TableCell>
            <div class="flex items-center gap-2">
              <component :is="getRoleIcon(member.role)" class="size-4 text-muted-foreground" />
              <Badge
                :variant="member.role === 'admin' ? 'default' : 'secondary'"
                class="capitalize"
              >
                {{ member.role }}
              </Badge>
            </div>
          </TableCell>
          <TableCell class="text-muted-foreground">
            {{ formatDate(member.joined_at) }}
          </TableCell>
          <TableCell v-if="isAdmin">
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label="Member actions"
                  :disabled="processingId === member.user_id"
                >
                  <Loader2
                    v-if="processingId === member.user_id"
                    class="size-4 animate-spin"
                  />
                  <MoreHorizontal v-else class="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuLabel>Change Role</DropdownMenuLabel>
                <DropdownMenuItem
                  :disabled="member.role === 'admin'"
                  @click="handleRoleChange(member, 'admin')"
                >
                  <Shield class="mr-2 size-4" />
                  Admin
                </DropdownMenuItem>
                <DropdownMenuItem
                  :disabled="member.role === 'member'"
                  @click="handleRoleChange(member, 'member')"
                >
                  <User class="mr-2 size-4" />
                  Member
                </DropdownMenuItem>
                <DropdownMenuItem
                  :disabled="member.role === 'guest'"
                  @click="handleRoleChange(member, 'guest')"
                >
                  <UserCog class="mr-2 size-4" />
                  Guest
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  class="text-destructive focus:text-destructive"
                  :disabled="member.user_id === user?.id"
                  @click="handleRemove(member)"
                >
                  <Trash2 class="mr-2 size-4" />
                  Remove
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  </div>
</template>
