import type { ProjectMember, ProjectState, WorkflowStateConfig, WorkflowTransitionConfig } from "~/types";

export interface WorkflowChange {
  kind: "add" | "remove" | "change";
  subject: string;
  detail: string;
}

export interface WorkflowPreview {
  mode: "changes" | "new" | "summary";
  changes: WorkflowChange[];
}

// Keyed by action id: once approved the "before" graph is gone, so keep what the approver saw.
export const workflowPreviewCache = new Map<string, WorkflowPreview>();

interface Graph {
  states: WorkflowStateConfig[];
  transitions: WorkflowTransitionConfig[];
}

const arr = <T>(v: unknown): T[] => (Array.isArray(v) ? v : []);

function joinList(items: string[], conj: "and" | "or") {
  if (items.length <= 1) return items.join("");
  return `${items.slice(0, -1).join(", ")} ${conj} ${items.at(-1)}`;
}

// A null `before` describes `after` on its own instead of diffing.
export function diffWorkflow(
  before: Graph | null,
  after: Graph,
  states: ProjectState[],
  members: ProjectMember[]
): WorkflowChange[] {
  const stateNames = new Map(states.map((s) => [s.id, s.name]));
  const userNames = new Map(
    members.map((m) => [m.user_id, `${m.first_name} ${m.last_name}`.trim() || m.username])
  );
  const stateName = (id: string) => stateNames.get(id) ?? "Unknown state";
  const users = (ids: string[]) => ids.map((id) => userNames.get(id) ?? "an unknown user").sort();
  const arrow = (t: WorkflowTransitionConfig) => `${stateName(t.from_state_id)} → ${stateName(t.to_state_id)}`;
  const movers = (t: WorkflowTransitionConfig) => {
    const who = users(arr<string>(t.mover_ids));
    if (t.allow_assignees) who.push("the task's assignees");
    if (t.allow_creator) who.push("the task's creator");
    return who.length ? joinList(who, "or") : "anyone";
  };

  const out: WorkflowChange[] = [];
  const prevStates = new Map(arr<WorkflowStateConfig>(before?.states).map((s) => [s.state_id, s]));
  const nextStates = new Map(arr<WorkflowStateConfig>(after.states).map((s) => [s.state_id, s]));

  for (const [id, s] of nextStates) {
    const prev = prevStates.get(id);
    const name = stateName(id);
    if (before && !prev) out.push({ kind: "add", subject: name, detail: "added to the workflow" });
    if (s.is_entry && !prev?.is_entry) out.push({ kind: "add", subject: name, detail: "new tasks can start here" });
    if (!s.is_entry && prev?.is_entry) out.push({ kind: "remove", subject: name, detail: "new tasks can no longer start here" });
    for (const [field, verb] of [["add_assignees", "assigns"], ["remove_assignees", "unassigns"]] as const) {
      const was = new Set(arr<string>(prev?.[field]));
      const now = new Set(arr<string>(s[field]));
      const added = users([...now].filter((u) => !was.has(u)));
      const dropped = users([...was].filter((u) => !now.has(u)));
      if (added.length) out.push({ kind: "add", subject: name, detail: `on enter, ${verb} ${joinList(added, "and")}` });
      if (dropped.length) out.push({ kind: "remove", subject: name, detail: `on enter, no longer ${verb} ${joinList(dropped, "and")}` });
    }
  }
  for (const id of prevStates.keys()) {
    if (!nextStates.has(id)) out.push({ kind: "remove", subject: stateName(id), detail: "removed from the workflow" });
  }

  const key = (t: WorkflowTransitionConfig) => `${t.from_state_id}>${t.to_state_id}`;
  const prevTransitions = new Map(arr<WorkflowTransitionConfig>(before?.transitions).map((t) => [key(t), t]));
  const nextTransitions = new Map(arr<WorkflowTransitionConfig>(after.transitions).map((t) => [key(t), t]));
  for (const [k, t] of prevTransitions) {
    if (!nextTransitions.has(k)) out.push({ kind: "remove", subject: arrow(t), detail: "no longer allowed" });
  }
  for (const [k, t] of nextTransitions) {
    const prev = prevTransitions.get(k);
    if (!prev) out.push({ kind: "add", subject: arrow(t), detail: `allowed for ${movers(t)}` });
    else if (movers(prev) !== movers(t)) {
      out.push({ kind: "change", subject: arrow(t), detail: `now allowed for ${movers(t)} (was ${movers(prev)})` });
    }
  }
  return out;
}
