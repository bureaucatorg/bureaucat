export interface Project {
  id: string;
  project_key: string;
  name: string;
  description?: string;
  icon_url?: string;
  cover_url?: string;
  role: string;
  disabled: boolean;
  require_template: boolean;
  workspace_id: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface PaginatedProjectsResponse {
  projects: Project[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface ProjectMember {
  id: string;
  user_id: string;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  role: string;
  avatar_url?: string;
  joined_at: string;
}

// A project member who is not a member of a given destination workspace.
// Returned by the move-project impact preview.
export interface MoveImpactMember {
  user_id: string;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  avatar_url?: string;
}

export interface ProjectState {
  id: string;
  state_type: StateType;
  name: string;
  color: string;
  position: number;
  is_default: boolean;
  created_at: string;
}

export type StateType = "backlog" | "unstarted" | "started" | "completed" | "cancelled";

export interface ProjectLabel {
  id: string;
  name: string;
  color: string;
  created_at: string;
}

export interface CreateProjectRequest {
  project_key: string;
  name: string;
  description?: string;
  icon_id?: string;
  cover_id?: string;
  workspace_id?: string;
}

export interface UpdateProjectRequest {
  name?: string;
  description?: string;
  icon_id?: string;
  cover_id?: string;
  require_template?: boolean;
}

export interface AddMemberRequest {
  user_id: string;
  role: string;
}

export interface UpdateMemberRequest {
  role: string;
}

export interface CreateStateRequest {
  state_type: StateType;
  name: string;
  color: string;
  position?: number;
}

export interface UpdateStateRequest {
  name?: string;
  color?: string;
  position?: number;
}

export interface CreateLabelRequest {
  name: string;
  color: string;
}

export interface UpdateLabelRequest {
  name?: string;
  color?: string;
}

export type TemplateFieldType = "text" | "textarea" | "number" | "select";

export interface TemplateField {
  key: string;
  label: string;
  type: TemplateFieldType;
  description?: string;
  required: boolean;
  options?: string[];
  multiple?: boolean;
  min?: number;
  max?: number;
}

export type TemplateFieldValue = string | number | string[] | undefined;

export interface TaskTemplate {
  id: string;
  name: string;
  title: string;
  description: string;
  fields: TemplateField[];
  default_priority: number | null;
  default_labels: string[];
  default_assignees: string[];
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface CreateTemplateRequest {
  name: string;
  title: string;
  description: string;
  fields: TemplateField[];
  default_priority: number | null;
  default_labels: string[];
  default_assignees: string[];
}

export type UpdateTemplateRequest = Partial<CreateTemplateRequest>;

export interface EditableTemplateField {
  uid: string;
  key: string;
  label: string;
  type: TemplateFieldType;
  description: string;
  required: boolean;
  optionsText: string;
  multiple: boolean;
  min: string;
  max: string;
}

