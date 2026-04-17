<script setup lang="ts">
import { Loader2, Send } from "lucide-vue-next";
import { toast } from "vue-sonner";

defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  "update:open": [value: boolean];
}>();

// Feedback is always sent to the main Bureaucat instance, regardless of
// where this code is running. Self-hosted admins opt in to expose the button
// via the `send_to_main_enabled` setting; the destination itself is fixed.
const MAIN_FEEDBACK_URL = "https://bureaucat.org/api/v1/feedback";

const { feedbackPublic } = useSettings();
const { getAuthHeader } = useAuth();

const message = ref("");
const submitting = ref(false);

/**
 * Submit the feedback. We intentionally fire the main-instance POST and the
 * local-mirror POST concurrently: the main one is the user's primary intent,
 * the local one is admin bookkeeping. The user sees success as long as *one*
 * of them succeeds so that a flaky bureaucat.org doesn't lose a local copy,
 * and a local DB hiccup doesn't hide that the main send worked.
 */
async function submit() {
  const msg = message.value.trim();
  if (!msg) return;
  submitting.value = true;

  const tasks: Promise<{ ok: boolean; label: string; status?: number }>[] = [];

  tasks.push(
    (async () => {
      try {
        const resp = await fetch(MAIN_FEEDBACK_URL, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          // No credentials — anonymous by design.
          body: JSON.stringify({ message: msg }),
        });
        return { ok: resp.ok || resp.status === 202, label: "main", status: resp.status };
      } catch {
        return { ok: false, label: "main" };
      }
    })()
  );

  if (feedbackPublic.value.store_sent_locally) {
    tasks.push(
      (async () => {
        try {
          const resp = await fetch("/api/v1/me/feedback", {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              ...getAuthHeader(),
            },
            credentials: "include",
            body: JSON.stringify({ message: msg }),
          });
          return { ok: resp.ok || resp.status === 202, label: "local", status: resp.status };
        } catch {
          return { ok: false, label: "local" };
        }
      })()
    );
  }

  const results = await Promise.all(tasks);
  submitting.value = false;

  const mainResult = results.find((r) => r.label === "main");
  const localResult = results.find((r) => r.label === "local");

  if (mainResult?.status === 429) {
    toast.error("Too many submissions — please wait a bit and try again.");
    return;
  }

  if (mainResult?.ok) {
    if (localResult && !localResult.ok) {
      toast.success("Thanks — sent to bureaucat.org (local copy failed).");
    } else {
      toast.success("Thanks for the feedback!");
    }
    message.value = "";
    emit("update:open", false);
    return;
  }

  // Main failed. If the local copy saved, still consider the submission
  // partially successful so the user doesn't retype.
  if (localResult?.ok) {
    toast.success("Saved a local copy — could not reach bureaucat.org.");
    message.value = "";
    emit("update:open", false);
    return;
  }

  toast.error("Could not send feedback — please try again later.");
}
</script>

<template>
  <Dialog :open="open" @update:open="(v: boolean) => emit('update:open', v)">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>Send anonymous feedback</DialogTitle>
        <DialogDescription>
          Your message goes to the Bureaucat team at
          <span class="font-mono">bureaucat.org</span>.
          <template v-if="feedbackPublic.store_sent_locally">
            A copy is also saved on this instance for your admins to review.
          </template>
          Nothing about your identity is attached.
        </DialogDescription>
      </DialogHeader>

      <div class="space-y-2">
        <Label for="feedback-message">Message</Label>
        <Textarea
          id="feedback-message"
          v-model="message"
          rows="6"
          placeholder="What's on your mind? Bugs, ideas, nitpicks — all welcome."
          :disabled="submitting"
          maxlength="5000"
        />
        <p class="text-right text-xs text-muted-foreground">
          {{ message.length }} / 5000
        </p>
      </div>

      <DialogFooter>
        <Button
          variant="outline"
          :disabled="submitting"
          @click="emit('update:open', false)"
        >
          Cancel
        </Button>
        <Button :disabled="!message.trim() || submitting" @click="submit">
          <Loader2 v-if="submitting" class="mr-2 size-4 animate-spin" />
          <Send v-else class="mr-2 size-4" />
          Send
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
