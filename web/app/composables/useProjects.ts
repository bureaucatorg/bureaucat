import type {
  Project,
  PaginatedProjectsResponse,
  ProjectMember,
  ProjectState,
  ProjectLabel,
  TaskTemplate,
  CreateProjectRequest,
  UpdateProjectRequest,
  AddMemberRequest,
  UpdateMemberRequest,
  CreateStateRequest,
  UpdateStateRequest,
  CreateLabelRequest,
  UpdateLabelRequest,
  CreateTemplateRequest,
  UpdateTemplateRequest,
} from "~/types";

interface ProjectsState {
  projects: Project[];
  currentProject: Project | null;
  members: ProjectMember[];
  states: ProjectState[];
  labels: ProjectLabel[];
  templates: TaskTemplate[];
  loading: boolean;
  total: number;
  page: number;
  perPage: number;
  totalPages: number;
}

const state = reactive<ProjectsState>({
  projects: [],
  currentProject: null,
  members: [],
  states: [],
  labels: [],
  templates: [],
  loading: false,
  total: 0,
  page: 1,
  perPage: 20,
  totalPages: 0,
});

export function useProjects() {
  const { getAuthHeader } = useAuth();

  // Projects CRUD
  async function listProjects(
    page = 1,
    perPage = 20
  ): Promise<{ success: boolean; data?: PaginatedProjectsResponse; error?: string }> {
    try {
      state.loading = true;
      const response = await fetch(
        `/api/v1/projects?page=${page}&per_page=${perPage}`,
        { headers: getAuthHeader() }
      );

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to fetch projects" };
      }

      const data: PaginatedProjectsResponse = await response.json();
      state.projects = data.projects || [];
      state.total = data.total;
      state.page = data.page;
      state.perPage = data.per_page;
      state.totalPages = data.total_pages;
      return { success: true, data };
    } catch {
      return { success: false, error: "Network error" };
    } finally {
      state.loading = false;
    }
  }

  async function createProject(
    data: CreateProjectRequest
  ): Promise<{ success: boolean; data?: Project; error?: string }> {
    try {
      const response = await fetch("/api/v1/projects", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to create project" };
      }

      const project: Project = await response.json();
      return { success: true, data: project };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function getProject(
    projectKey: string
  ): Promise<{ success: boolean; data?: Project; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}`, {
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to fetch project" };
      }

      const project: Project = await response.json();
      state.currentProject = project;
      return { success: true, data: project };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function updateProject(
    projectKey: string,
    data: UpdateProjectRequest
  ): Promise<{ success: boolean; data?: Project; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to update project" };
      }

      const project: Project = await response.json();
      state.currentProject = project;
      return { success: true, data: project };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function deleteProject(
    projectKey: string
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}`, {
        method: "DELETE",
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to delete project" };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  // Members
  async function listMembers(
    projectKey: string
  ): Promise<{ success: boolean; data?: ProjectMember[]; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/members`, {
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to fetch members" };
      }

      const members: ProjectMember[] = await response.json();
      state.members = members;
      return { success: true, data: members };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function addMember(
    projectKey: string,
    data: AddMemberRequest
  ): Promise<{ success: boolean; data?: ProjectMember; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/members`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to add member" };
      }

      const member: ProjectMember = await response.json();
      return { success: true, data: member };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function updateMemberRole(
    projectKey: string,
    userId: string,
    data: UpdateMemberRequest
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/members/${userId}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to update member role" };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function removeMember(
    projectKey: string,
    userId: string
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/members/${userId}`, {
        method: "DELETE",
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to remove member" };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  // States
  async function listStates(
    projectKey: string
  ): Promise<{ success: boolean; data?: ProjectState[]; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/states`, {
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to fetch states" };
      }

      const states: ProjectState[] = await response.json();
      state.states = states;
      return { success: true, data: states };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function createState(
    projectKey: string,
    data: CreateStateRequest
  ): Promise<{ success: boolean; data?: ProjectState; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/states`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to create state" };
      }

      const stateData: ProjectState = await response.json();
      return { success: true, data: stateData };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function updateState(
    projectKey: string,
    stateId: string,
    data: UpdateStateRequest
  ): Promise<{ success: boolean; data?: ProjectState; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/states/${stateId}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to update state" };
      }

      const stateData: ProjectState = await response.json();
      return { success: true, data: stateData };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function deleteState(
    projectKey: string,
    stateId: string
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/states/${stateId}`, {
        method: "DELETE",
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to delete state" };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  // Labels
  async function listLabels(
    projectKey: string
  ): Promise<{ success: boolean; data?: ProjectLabel[]; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/labels`, {
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to fetch labels" };
      }

      const labels: ProjectLabel[] = await response.json();
      state.labels = labels;
      return { success: true, data: labels };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function createLabel(
    projectKey: string,
    data: CreateLabelRequest
  ): Promise<{ success: boolean; data?: ProjectLabel; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/labels`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to create label" };
      }

      const label: ProjectLabel = await response.json();
      return { success: true, data: label };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function updateLabel(
    projectKey: string,
    labelId: string,
    data: UpdateLabelRequest
  ): Promise<{ success: boolean; data?: ProjectLabel; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/labels/${labelId}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to update label" };
      }

      const label: ProjectLabel = await response.json();
      return { success: true, data: label };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function deleteLabel(
    projectKey: string,
    labelId: string
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/labels/${labelId}`, {
        method: "DELETE",
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to delete label" };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  // Templates
  async function listTemplates(
    projectKey: string
  ): Promise<{ success: boolean; data?: TaskTemplate[]; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/templates`, {
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to fetch templates" };
      }

      const templates: TaskTemplate[] = await response.json();
      state.templates = templates;
      return { success: true, data: templates };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function createTemplate(
    projectKey: string,
    data: CreateTemplateRequest
  ): Promise<{ success: boolean; data?: TaskTemplate; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/templates`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to create template" };
      }

      const template: TaskTemplate = await response.json();
      return { success: true, data: template };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function updateTemplate(
    projectKey: string,
    templateId: string,
    data: UpdateTemplateRequest
  ): Promise<{ success: boolean; data?: TaskTemplate; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/templates/${templateId}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          ...getAuthHeader(),
        },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to update template" };
      }

      const template: TaskTemplate = await response.json();
      return { success: true, data: template };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  async function deleteTemplate(
    projectKey: string,
    templateId: string
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await fetch(`/api/v1/projects/${projectKey}/templates/${templateId}`, {
        method: "DELETE",
        headers: getAuthHeader(),
      });

      if (!response.ok) {
        const error = await response.json();
        return { success: false, error: error.message || "Failed to delete template" };
      }

      return { success: true };
    } catch {
      return { success: false, error: "Network error" };
    }
  }

  function clearCurrentProject() {
    state.currentProject = null;
    state.members = [];
    state.states = [];
    state.labels = [];
    state.templates = [];
  }

  return {
    // State (readonly)
    projects: computed(() => state.projects),
    currentProject: computed(() => state.currentProject),
    members: computed(() => state.members),
    states: computed(() => state.states),
    labels: computed(() => state.labels),
    templates: computed(() => state.templates),
    loading: computed(() => state.loading),
    total: computed(() => state.total),
    page: computed(() => state.page),
    perPage: computed(() => state.perPage),
    totalPages: computed(() => state.totalPages),

    // Projects
    listProjects,
    createProject,
    getProject,
    updateProject,
    deleteProject,

    // Members
    listMembers,
    addMember,
    updateMemberRole,
    removeMember,

    // States
    listStates,
    createState,
    updateState,
    deleteState,

    // Labels
    listLabels,
    createLabel,
    updateLabel,
    deleteLabel,

    // Templates
    listTemplates,
    createTemplate,
    updateTemplate,
    deleteTemplate,

    // Utils
    clearCurrentProject,
  };
}
