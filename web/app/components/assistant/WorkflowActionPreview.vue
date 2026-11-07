<script setup lang="ts">
import { ArrowRightLeft, Loader2, Minus, Plus } from "lucide-vue-next";
import type { AssistantAction } from "~/composables/useAssistant";
import type { ProjectMember, ProjectState, SaveWorkflowRequest, Workflow } from "~/types";
import { diffWorkflow, workflowPreviewCache, type WorkflowPreview } from "~/utils/workflowDiff";

const props = defineProps<{ action: AssistantAction }>();
const { getAuthHeader } = useAuth();

const VISIBLE = 8;
const loading = ref(true);
const preview = ref<WorkflowPreview | null>(null);
const expanded = ref(false);

const projectKey = computed(() => String(props.action.args.project_key ?? ""));
const shown = computed(() =>
  expanded.value ? preview.value?.changes ?? [] : (preview.value?.changes ?? []).slice(0, VISIBLE)
);
const heading = computed(() => {
  switch (preview.value?.mode) {
    case "changes":
      return `Changes to the ${projectKey.value} workflow`;
    case "new":
      return `New workflow for ${projectKey.value}`;
    default:
      return `${projectKey.value} workflow as saved`;
  }
});

async function getJSON<T>(path: string): Promise<T | null> {
  try {
    const res = await fetch(`/api/v1/projects/${encodeURIComponent(projectKey.value)}${path}`, {
      headers: getAuthHeader(),
    });
    return res.ok ? await res.json() : null;
  } catch {
    return null;
  }
}

async function build() {
  const cached = workflowPreviewCache.get(props.action.id);
  if (cached) {
    preview.value = cached;
    loading.value = false;
    return;
  }
  const pending = props.action.status === "pending";
  const [states, members, current] = await Promise.all([
    getJSON<ProjectState[]>("/states"),
    getJSON<ProjectMember[]>("/members"),
    pending ? getJSON<Workflow>("/workflow") : Promise.resolve(null),
  ]);
  loading.value = false;
  if (!states || (pending && !current)) return;
  const before = current?.configured ? current : null;
  const proposed = props.action.args as unknown as SaveWorkflowRequest;
  const result: WorkflowPreview = {
    mode: !pending ? "summary" : before ? "changes" : "new",
    changes: diffWorkflow(before, proposed, states, members ?? []),
  };
  preview.value = result;
  if (pending) workflowPreviewCache.set(props.action.id, result);
}

onMounted(build);
</script>

<template>
  <div class="mt-2 text-xs">
    <p v-if="loading" class="flex items-center gap-1.5 text-muted-foreground">
      <Loader2 class="size-3 animate-spin" /> Working out what changes…
    </p>
    <slot v-else-if="!preview" />
    <template v-else>
      <p class="mb-1.5 text-muted-foreground">{{ heading }}</p>
      <p v-if="preview.changes.length === 0" class="text-muted-foreground">
        Nothing changes apart from the layout.
      </p>
      <ul v-else class="space-y-1">
        <li v-for="(c, i) in shown" :key="i" class="flex items-start gap-1.5">
          <Plus v-if="c.kind === 'add'" class="mt-0.5 size-3 shrink-0 text-emerald-500" />
          <Minus v-else-if="c.kind === 'remove'" class="mt-0.5 size-3 shrink-0 text-destructive" />
          <ArrowRightLeft v-else class="mt-0.5 size-3 shrink-0 text-amber-500" />
          <span class="min-w-0 break-words">
            <span class="font-medium">{{ c.subject }}</span>
            <span class="text-muted-foreground">: {{ c.detail }}</span>
          </span>
        </li>
      </ul>
      <button
        v-if="preview.changes.length > VISIBLE"
        type="button"
        class="mt-1.5 text-muted-foreground underline-offset-2 hover:text-foreground hover:underline"
        @click="expanded = !expanded"
      >
        {{ expanded ? "Show less" : `Show ${preview.changes.length - VISIBLE} more` }}
      </button>
    </template>
  </div>
</template>
