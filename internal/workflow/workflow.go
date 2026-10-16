// Package workflow enforces a project's opt-in state transition rules.
package workflow

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"bereaucat/internal/store"
)

const (
	ActionAddAssignee    = "add_assignee"
	ActionRemoveAssignee = "remove_assignee"
)

// Error is a rule violation; Status is the HTTP status handlers should return.
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string { return e.Message }

// Enabled reports whether the project has a saved workflow that is switched on.
func Enabled(ctx context.Context, q store.Querier, projectID uuid.UUID) (bool, error) {
	wf, err := q.GetProjectWorkflow(ctx, projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return wf.Enabled, nil
}

// CheckTransition verifies actorID may move the task from fromID to toID.
func CheckTransition(ctx context.Context, q store.Querier, projectID, taskID, createdBy, fromID, toID, actorID uuid.UUID, isAdmin bool) error {
	if isAdmin || fromID == toID {
		return nil
	}
	on, err := Enabled(ctx, q, projectID)
	if err != nil || !on {
		return err
	}

	from, _ := q.GetProjectStateByID(ctx, fromID)
	to, _ := q.GetProjectStateByID(ctx, toID)

	tr, err := q.GetWorkflowTransition(ctx, store.GetWorkflowTransitionParams{FromStateID: fromID, ToStateID: toID})
	if errors.Is(err, pgx.ErrNoRows) {
		return &Error{http.StatusUnprocessableEntity, fmt.Sprintf("the project workflow does not allow moving from %q to %q", from.Name, to.Name)}
	}
	if err != nil {
		return err
	}

	movers, err := q.ListTransitionMoverIDs(ctx, tr.ID)
	if err != nil {
		return err
	}
	if len(movers) == 0 && !tr.AllowAssignees && !tr.AllowCreator {
		return nil
	}
	for _, id := range movers {
		if id == actorID {
			return nil
		}
	}
	if tr.AllowCreator && createdBy == actorID {
		return nil
	}
	if tr.AllowAssignees {
		assigned, err := q.IsTaskAssignee(ctx, store.IsTaskAssigneeParams{TaskID: taskID, UserID: actorID})
		if err != nil {
			return err
		}
		if assigned {
			return nil
		}
	}
	return &Error{http.StatusForbidden, fmt.Sprintf("you are not allowed to move tasks from %q to %q", from.Name, to.Name)}
}

// CheckEntry verifies a new task may start in stateID.
func CheckEntry(ctx context.Context, q store.Querier, projectID, stateID uuid.UUID, isAdmin bool) error {
	if isAdmin {
		return nil
	}
	on, err := Enabled(ctx, q, projectID)
	if err != nil || !on {
		return err
	}
	ws, err := q.GetWorkflowState(ctx, stateID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if err == nil && ws.IsEntry {
		return nil
	}
	st, _ := q.GetProjectStateByID(ctx, stateID)
	return &Error{http.StatusUnprocessableEntity, fmt.Sprintf("the project workflow does not allow new tasks to start in %q", st.Name)}
}

// ApplyOnEnter runs stateID's on-enter actions for the task and returns the
// users it assigned and unassigned. Users no longer in the project are skipped.
func ApplyOnEnter(ctx context.Context, q store.Querier, projectID, taskID, stateID, actorID uuid.UUID) (added, removed []uuid.UUID, err error) {
	on, err := Enabled(ctx, q, projectID)
	if err != nil || !on {
		return nil, nil, err
	}
	actions, err := q.ListStateActions(ctx, stateID)
	if err != nil {
		return nil, nil, err
	}
	for _, a := range actions {
		assigned, err := q.IsTaskAssignee(ctx, store.IsTaskAssigneeParams{TaskID: taskID, UserID: a.UserID})
		if err != nil {
			return added, removed, err
		}
		switch a.Action {
		case ActionAddAssignee:
			if assigned {
				continue
			}
			member, err := q.IsProjectMember(ctx, store.IsProjectMemberParams{ProjectID: projectID, UserID: a.UserID})
			if err != nil {
				return added, removed, err
			}
			if !member {
				continue
			}
			if _, err := q.AddTaskAssignee(ctx, store.AddTaskAssigneeParams{TaskID: taskID, UserID: a.UserID, AssignedBy: actorID}); err != nil {
				return added, removed, err
			}
			added = append(added, a.UserID)
		case ActionRemoveAssignee:
			if !assigned {
				continue
			}
			if err := q.RemoveTaskAssignee(ctx, store.RemoveTaskAssigneeParams{TaskID: taskID, UserID: a.UserID}); err != nil {
				return added, removed, err
			}
			removed = append(removed, a.UserID)
		}
	}
	return added, removed, nil
}
