<script setup lang="ts">
import { History, Loader2, ShieldCheck, ShieldAlert } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { ActivityLogEntry } from "~/types";

const props = defineProps<{
  activities: ActivityLogEntry[];
  projectKey: string;
  taskNum: number;
  loading?: boolean;
}>();

const { verifyActivity } = useActivity();

const verifying = ref(false);
const verificationResult = ref<{ valid: boolean; message: string } | null>(null);

async function handleVerify() {
  verifying.value = true;
  verificationResult.value = null;

  const result = await verifyActivity(props.projectKey, props.taskNum);

  verifying.value = false;

  if (result.success && result.data) {
    verificationResult.value = result.data;
    if (result.data.valid) {
      toast.success("Activity log verified successfully");
    } else {
      toast.error("Activity log integrity compromised");
    }
  } else {
    toast.error(result.error || "Failed to verify activity");
  }
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h3 class="flex items-center gap-2 font-semibold">
        <History class="size-4" />
        Activity
      </h3>
      <Button variant="outline" size="sm" :disabled="verifying" @click="handleVerify">
        <Loader2 v-if="verifying" class="mr-1.5 size-3.5 animate-spin" />
        <ShieldCheck v-else class="mr-1.5 size-3.5" />
        Verify Log
      </Button>
    </div>

    <!-- Verification result -->
    <Alert
      v-if="verificationResult"
      :variant="verificationResult.valid ? 'default' : 'destructive'"
    >
      <component
        :is="verificationResult.valid ? ShieldCheck : ShieldAlert"
        class="size-4"
      />
      <AlertTitle>
        {{ verificationResult.valid ? "Verified" : "Integrity Issue" }}
      </AlertTitle>
      <AlertDescription>
        {{ verificationResult.message }}
      </AlertDescription>
    </Alert>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-8">
      <Loader2 class="size-5 animate-spin text-muted-foreground" />
    </div>

    <!-- Activities -->
    <div v-else-if="activities.length > 0" class="space-y-0">
      <ActivityItem
        v-for="(activity, index) in activities"
        :key="activity.id"
        :activity="activity"
        :is-last="index === activities.length - 1"
      />
    </div>

    <!-- Empty state -->
    <div
      v-else
      class="rounded-lg border border-dashed py-8 text-center text-sm text-muted-foreground"
    >
      No activity yet
    </div>
  </div>
</template>
