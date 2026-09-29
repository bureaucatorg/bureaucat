<script setup lang="ts">
import { Bell, BellOff } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { TaskFollower } from "~/types";

const props = defineProps<{
  followers: TaskFollower[];
  projectKey: string;
  taskNum: number;
}>();

const emit = defineEmits<{
  refresh: [];
}>();

const { user } = useAuth();
const { setFollowing } = useTasks();

const loading = ref(false);

const isFollowing = computed(() =>
  props.followers.some((f) => f.user_id === user.value?.id),
);

async function toggle() {
  loading.value = true;
  const follow = !isFollowing.value;
  const result = await setFollowing(props.projectKey, props.taskNum, follow);
  loading.value = false;

  if (result.success) {
    toast.success(follow ? "Following task" : "Unfollowed task");
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to update follow status");
  }
}
</script>

<template>
  <div class="space-y-2">
    <div class="flex items-center justify-between gap-2">
      <p class="text-xs text-muted-foreground">Followers</p>
      <Button
        variant="outline"
        size="sm"
        class="h-7 px-2 text-xs"
        :disabled="loading"
        @click="toggle"
      >
        <BellOff v-if="isFollowing" class="mr-1 size-3.5" />
        <Bell v-else class="mr-1 size-3.5" />
        {{ isFollowing ? "Unfollow" : "Follow" }}
      </Button>
    </div>

    <div class="flex flex-wrap items-center gap-2">
      <NuxtLink
        v-for="follower in followers"
        :key="follower.user_id"
        :to="`/profile/${follower.user_id}`"
        class="flex items-center gap-1.5 rounded-md border bg-muted/50 py-1 pl-1 pr-2.5 transition-opacity hover:opacity-80"
      >
        <Avatar class="size-5">
          <AvatarImage v-if="follower.avatar_url" :src="follower.avatar_url" />
          <AvatarFallback class="text-[10px]" :seed="follower.user_id">
            {{ follower.first_name[0] }}{{ follower.last_name[0] }}
          </AvatarFallback>
        </Avatar>
        <span class="text-sm">
          {{ follower.first_name }} {{ follower.last_name }}
        </span>
      </NuxtLink>
      <span v-if="followers.length === 0" class="text-sm text-muted-foreground">
        No followers
      </span>
    </div>
  </div>
</template>
