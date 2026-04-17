<script setup lang="ts">
import { Loader2 } from "lucide-vue-next";
import type { Task, ProjectState, ProjectMember, ProjectLabel, ViewGroupBy } from "~/types";
import { PRIORITY_LABELS, STATE_TYPE_COLORS } from "~/types";

interface BoardColumn {
  id: string;
  label: string;
  color: string;
  tasks: Task[];
  /** Whether cards can be dropped into this column under the current grouping. */
  dropLocked?: boolean;
}

const props = defineProps<{
  tasks: Task[];
  states: ProjectState[];
  members: ProjectMember[];
  labels: ProjectLabel[];
  projectKey: string;
  isMember: boolean;
  groupBy: ViewGroupBy;
  currentUserId?: string;
}>();

const emit = defineEmits<{
  refresh: [];
}>();

const { updateTask, addAssignee, removeAssignee, addLabel, removeLabel } = useTasks();
const updating = ref(false);

const STATE_TYPE_ORDER = ["backlog", "unstarted", "started", "completed", "cancelled"] as const;
const STATE_TYPE_LABELS: Record<string, string> = {
  backlog: "Backlog",
  unstarted: "To Do",
  started: "In Progress",
  completed: "Done",
  cancelled: "Cancelled",
};

const columns = computed<BoardColumn[]>(() => {
  switch (props.groupBy) {
    case "state_type":
      return STATE_TYPE_ORDER
        .filter((t) => props.states.some((s) => s.state_type === t))
        .map((t) => ({
          id: `state_type:${t}`,
          label: STATE_TYPE_LABELS[t] ?? t,
          color: STATE_TYPE_COLORS[t] ?? "#6B7280",
          tasks: props.tasks.filter((task) => task.state_type === t),
        }));

    case "state":
      return props.states.map((s) => ({
        id: `state:${s.id}`,
        label: s.name,
        color: s.color || STATE_TYPE_COLORS[s.state_type] || "#6B7280",
        tasks: props.tasks.filter((task) => task.state_id === s.id),
      }));

    case "priority": {
      const PRIORITY_ORDER = [4, 3, 2, 1, 0];
      return PRIORITY_ORDER.map((p) => ({
        id: `priority:${p}`,
        label: PRIORITY_LABELS[p]?.label ?? String(p),
        color: PRIORITY_LABELS[p]?.color ?? "#6B7280",
        tasks: props.tasks.filter((task) => task.priority === p),
      }));
    }

    case "assignee": {
      const cols: BoardColumn[] = props.members.map((m) => ({
        id: `assignee:${m.user_id}`,
        label: `${m.first_name} ${m.last_name}`.trim() || m.username,
        color: "#3B82F6",
        tasks: props.tasks.filter((t) =>
          (t.assignees ?? []).some((a) => a.user_id === m.user_id)
        ),
      }));
      cols.push({
        id: "assignee:__none__",
        label: "Unassigned",
        color: "#9CA3AF",
        tasks: props.tasks.filter((t) => !t.assignees || t.assignees.length === 0),
      });
      return cols;
    }

    case "label": {
      const cols: BoardColumn[] = props.labels.map((l) => ({
        id: `label:${l.id}`,
        label: l.name,
        color: l.color || "#3B82F6",
        tasks: props.tasks.filter((t) => (t.labels ?? []).some((tl) => tl.id === l.id)),
      }));
      cols.push({
        id: "label:__none__",
        label: "No label",
        color: "#9CA3AF",
        tasks: props.tasks.filter((t) => !t.labels || t.labels.length === 0),
      });
      return cols;
    }

    case "due_bucket": {
      return dueBucketColumns(props.tasks);
    }

    case "start_bucket": {
      return startBucketColumns(props.tasks);
    }

    case "created_bucket": {
      return activityBucketColumns(props.tasks, "created_at");
    }

    case "updated_bucket": {
      return activityBucketColumns(props.tasks, "updated_at");
    }
  }
});

function dueBucketColumns(tasks: Task[]): BoardColumn[] {
  const now = new Date();
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
  const startOfTomorrow = startOfToday + 24 * 60 * 60 * 1000;
  const weekday = (now.getDay() + 6) % 7; // Mon=0..Sun=6
  const startOfWeek = startOfToday - weekday * 24 * 60 * 60 * 1000;
  const startOfNextWeek = startOfWeek + 7 * 24 * 60 * 60 * 1000;
  const startOfWeekAfter = startOfNextWeek + 7 * 24 * 60 * 60 * 1000;

  const buckets: Record<string, Task[]> = {
    overdue: [], today: [], this_week: [], next_week: [], later: [], none: [],
  };

  for (const t of tasks) {
    if (!t.due_date) {
      buckets.none.push(t);
      continue;
    }
    const ts = new Date(t.due_date).getTime();
    if (ts < startOfToday && t.state_type !== "completed" && t.state_type !== "cancelled") {
      buckets.overdue.push(t);
    } else if (ts < startOfTomorrow) {
      buckets.today.push(t);
    } else if (ts < startOfNextWeek) {
      buckets.this_week.push(t);
    } else if (ts < startOfWeekAfter) {
      buckets.next_week.push(t);
    } else {
      buckets.later.push(t);
    }
  }

  return [
    { id: "due:overdue", label: "Overdue", color: "#EF4444", tasks: buckets.overdue, dropLocked: true },
    { id: "due:today", label: "Today", color: "#F97316", tasks: buckets.today, dropLocked: true },
    { id: "due:this_week", label: "This week", color: "#EAB308", tasks: buckets.this_week, dropLocked: true },
    { id: "due:next_week", label: "Next week", color: "#3B82F6", tasks: buckets.next_week, dropLocked: true },
    { id: "due:later", label: "Later", color: "#6B7280", tasks: buckets.later, dropLocked: true },
    { id: "due:none", label: "No due date", color: "#9CA3AF", tasks: buckets.none, dropLocked: true },
  ];
}

function startBucketColumns(tasks: Task[]): BoardColumn[] {
  const now = new Date();
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
  const startOfTomorrow = startOfToday + 24 * 60 * 60 * 1000;
  const weekday = (now.getDay() + 6) % 7;
  const startOfWeek = startOfToday - weekday * 24 * 60 * 60 * 1000;
  const startOfNextWeek = startOfWeek + 7 * 24 * 60 * 60 * 1000;
  const startOfWeekAfter = startOfNextWeek + 7 * 24 * 60 * 60 * 1000;

  const buckets: Record<string, Task[]> = {
    started: [], today: [], this_week: [], next_week: [], later: [], none: [],
  };

  for (const t of tasks) {
    if (!t.start_date) {
      buckets.none.push(t);
      continue;
    }
    const ts = new Date(t.start_date).getTime();
    if (ts < startOfToday) {
      buckets.started.push(t);
    } else if (ts < startOfTomorrow) {
      buckets.today.push(t);
    } else if (ts < startOfNextWeek) {
      buckets.this_week.push(t);
    } else if (ts < startOfWeekAfter) {
      buckets.next_week.push(t);
    } else {
      buckets.later.push(t);
    }
  }

  return [
    { id: "start:started", label: "Already started", color: "#10B981", tasks: buckets.started, dropLocked: true },
    { id: "start:today", label: "Today", color: "#F97316", tasks: buckets.today, dropLocked: true },
    { id: "start:this_week", label: "This week", color: "#EAB308", tasks: buckets.this_week, dropLocked: true },
    { id: "start:next_week", label: "Next week", color: "#3B82F6", tasks: buckets.next_week, dropLocked: true },
    { id: "start:later", label: "Later", color: "#6B7280", tasks: buckets.later, dropLocked: true },
    { id: "start:none", label: "No start date", color: "#9CA3AF", tasks: buckets.none, dropLocked: true },
  ];
}

function activityBucketColumns(tasks: Task[], field: "created_at" | "updated_at"): BoardColumn[] {
  const now = new Date();
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
  const startOfTomorrow = startOfToday + 24 * 60 * 60 * 1000;
  const weekday = (now.getDay() + 6) % 7;
  const startOfWeek = startOfToday - weekday * 24 * 60 * 60 * 1000;
  const startOfMonth = new Date(now.getFullYear(), now.getMonth(), 1).getTime();

  const buckets: Record<string, Task[]> = {
    today: [], yesterday: [], this_week: [], this_month: [], older: [],
  };

  for (const t of tasks) {
    const raw = t[field];
    if (!raw) {
      buckets.older.push(t);
      continue;
    }
    const ts = new Date(raw).getTime();
    if (ts >= startOfToday && ts < startOfTomorrow) {
      buckets.today.push(t);
    } else if (ts >= startOfToday - 24 * 60 * 60 * 1000 && ts < startOfToday) {
      buckets.yesterday.push(t);
    } else if (ts >= startOfWeek) {
      buckets.this_week.push(t);
    } else if (ts >= startOfMonth) {
      buckets.this_month.push(t);
    } else {
      buckets.older.push(t);
    }
  }

  const prefix = field === "created_at" ? "created" : "updated";
  return [
    { id: `${prefix}:today`, label: "Today", color: "#F97316", tasks: buckets.today, dropLocked: true },
    { id: `${prefix}:yesterday`, label: "Yesterday", color: "#EAB308", tasks: buckets.yesterday, dropLocked: true },
    { id: `${prefix}:this_week`, label: "This week", color: "#3B82F6", tasks: buckets.this_week, dropLocked: true },
    { id: `${prefix}:this_month`, label: "This month", color: "#8B5CF6", tasks: buckets.this_month, dropLocked: true },
    { id: `${prefix}:older`, label: "Older", color: "#6B7280", tasks: buckets.older, dropLocked: true },
  ];
}

// ---- drop handler: route to the right mutation based on groupBy ----

async function handleDrop(task: Task, fromColumnId: string, toColumnId: string) {
  if (fromColumnId === toColumnId || !props.isMember || updating.value) return;
  updating.value = true;
  try {
    const [, toValue] = toColumnId.split(":");
    switch (props.groupBy) {
      case "state_type": {
        const targetState = props.states.find((s) => s.state_type === toValue);
        if (targetState) {
          await updateTask(props.projectKey, task.task_number, { state_id: targetState.id });
        }
        break;
      }
      case "state": {
        await updateTask(props.projectKey, task.task_number, { state_id: toValue });
        break;
      }
      case "priority": {
        const p = parseInt(toValue, 10);
        if (!isNaN(p)) {
          await updateTask(props.projectKey, task.task_number, { priority: p });
        }
        break;
      }
      case "assignee": {
        // Leaving the "from" column and joining the "to" column. If from is a
        // real user, remove that assignment; if to is a real user, add it.
        // Dropping into Unassigned removes *all* assignees on the task.
        const fromValue = fromColumnId.split(":")[1];
        if (toValue === "__none__") {
          for (const a of task.assignees ?? []) {
            await removeAssignee(props.projectKey, task.task_number, a.user_id);
          }
        } else {
          if (fromValue && fromValue !== "__none__") {
            await removeAssignee(props.projectKey, task.task_number, fromValue);
          }
          await addAssignee(props.projectKey, task.task_number, toValue);
        }
        break;
      }
      case "label": {
        const fromValue = fromColumnId.split(":")[1];
        if (toValue === "__none__") {
          for (const l of task.labels ?? []) {
            await removeLabel(props.projectKey, task.task_number, l.id);
          }
        } else {
          if (fromValue && fromValue !== "__none__") {
            await removeLabel(props.projectKey, task.task_number, fromValue);
          }
          await addLabel(props.projectKey, task.task_number, toValue);
        }
        break;
      }
      case "due_bucket":
      case "start_bucket":
      case "created_bucket":
      case "updated_bucket":
        // Locked — drop not allowed.
        return;
    }
    emit("refresh");
  } finally {
    updating.value = false;
  }
}
</script>

<template>
  <div class="relative">
    <div
      v-if="updating"
      class="absolute inset-0 z-10 flex items-center justify-center bg-background/50"
    >
      <Loader2 class="size-6 animate-spin" />
    </div>

    <div class="flex gap-4 overflow-x-auto pb-4">
      <KanbanColumn
        v-for="column in columns"
        :key="column.id"
        :column-id="column.id"
        :label="column.label"
        :color="column.color"
        :tasks="column.tasks"
        :project-key="projectKey"
        :is-member="isMember"
        :drop-locked="column.dropLocked"
        @drop="handleDrop"
      />
    </div>
  </div>
</template>
