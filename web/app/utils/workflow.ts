import type { Workflow, WorkflowTaskRef, WorkflowTransitionConfig } from "~/types";

function isActive(wf: Workflow | undefined): wf is Workflow {
  return !!wf && wf.enabled && !wf.bypass;
}

// Mirrors internal/workflow.CheckTransition. Unknown assignees count as allowed;
// the server stays authoritative.
export function canUseTransition(t: WorkflowTransitionConfig, task: WorkflowTaskRef, userId?: string) {
  if (t.mover_ids.length === 0 && !t.allow_assignees && !t.allow_creator) return true;
  if (userId && t.mover_ids.includes(userId)) return true;
  if (t.allow_creator && userId && task.created_by === userId) return true;
  if (t.allow_assignees) {
    if (!task.assignees) return true;
    return !!userId && task.assignees.some((a) => a.user_id === userId);
  }
  return false;
}

// Target states the user may move the task to, or null when unrestricted.
export function allowedTargetStateIds(
  wf: Workflow | undefined,
  task: WorkflowTaskRef,
  userId?: string
): Set<string> | null {
  if (!isActive(wf)) return null;
  const allowed = new Set<string>([task.state_id]);
  for (const t of wf.transitions) {
    if (t.from_state_id === task.state_id && canUseTransition(t, task, userId)) {
      allowed.add(t.to_state_id);
    }
  }
  return allowed;
}

// States a new task may start in, or null when unrestricted.
export function entryStateIds(wf: Workflow | undefined): Set<string> | null {
  if (!isActive(wf)) return null;
  return new Set(wf.states.filter((s) => s.is_entry).map((s) => s.state_id));
}
