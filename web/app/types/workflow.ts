export interface WorkflowStateConfig {
  state_id: string;
  is_entry: boolean;
  pos_x: number;
  pos_y: number;
  add_assignees: string[];
  remove_assignees: string[];
}

export interface WorkflowTransitionConfig {
  from_state_id: string;
  to_state_id: string;
  allow_assignees: boolean;
  allow_creator: boolean;
  mover_ids: string[];
}

export interface Workflow {
  configured: boolean;
  enabled: boolean;
  bypass: boolean;
  states: WorkflowStateConfig[];
  transitions: WorkflowTransitionConfig[];
}

export interface SaveWorkflowRequest {
  states: WorkflowStateConfig[];
  transitions: WorkflowTransitionConfig[];
}

// The task fields the workflow rules look at.
export interface WorkflowTaskRef {
  project_key: string;
  state_id: string;
  created_by?: string;
  assignees?: { user_id: string }[];
}
