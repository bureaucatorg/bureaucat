export interface Comment {
  id: string;
  task_id: string;
  content: string;
  version: number;
  created_by: string;
  username: string;
  first_name: string;
  last_name: string;
  created_at: string;
  updated_at: string;
}

export interface CreateCommentRequest {
  content: string;
}

export interface UpdateCommentRequest {
  content: string;
}
