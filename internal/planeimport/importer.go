package planeimport

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"bereaucat/internal/auth"
)

// ImportResult holds the counts of imported entities.
type ImportResult struct {
	UsersCreated    int `json:"users_created"`
	UsersSkipped    int `json:"users_skipped"`
	ProjectsCreated int `json:"projects_created"`
	StatesCreated   int `json:"states_created"`
	LabelsCreated   int `json:"labels_created"`
	TasksCreated    int `json:"tasks_created"`
	AssigneesLinked int `json:"assignees_linked"`
	LabelsAssigned  int `json:"labels_assigned"`
	CommentsCreated int `json:"comments_created"`
}

// Import processes a parsed Plane dump and inserts all data into Bureaucat
// within a single database transaction using bulk inserts.
func Import(ctx context.Context, pool *pgxpool.Pool, adminUserID uuid.UUID, dump *ParsedDump) (*ImportResult, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	result := &ImportResult{}

	// Load existing users into a map for dedup.
	existingUsers, err := loadExistingUsers(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("loading existing users: %w", err)
	}

	// Phase 1: Bulk import users.
	userMap, err := bulkImportUsers(ctx, tx, dump.Users, existingUsers, result)
	if err != nil {
		return nil, fmt.Errorf("importing users: %w", err)
	}
	log.Printf("Users phase complete: %d created, %d skipped", result.UsersCreated, result.UsersSkipped)

	// Phase 2: Bulk import projects.
	existingKeys, err := loadExistingProjectKeys(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("loading project keys: %w", err)
	}
	projectMap, err := bulkImportProjects(ctx, tx, dump.Projects, adminUserID, existingKeys, result)
	if err != nil {
		return nil, fmt.Errorf("importing projects: %w", err)
	}
	log.Printf("Projects phase complete: %d created", result.ProjectsCreated)

	// Phase 3: Bulk import project members.
	membersCreated, err := bulkImportProjectMembers(ctx, tx, dump.ProjectMembers, userMap, projectMap, adminUserID)
	if err != nil {
		return nil, fmt.Errorf("importing project members: %w", err)
	}
	log.Printf("Project members phase complete: %d added", membersCreated)

	// Phase 4: Bulk import states.
	stateMap, err := bulkImportStates(ctx, tx, dump.States, projectMap, result)
	if err != nil {
		return nil, fmt.Errorf("importing states: %w", err)
	}
	log.Printf("States phase complete: %d created", result.StatesCreated)

	// Phase 5: Bulk import labels.
	labelMap, err := bulkImportLabels(ctx, tx, dump.Labels, projectMap, result)
	if err != nil {
		return nil, fmt.Errorf("importing labels: %w", err)
	}
	log.Printf("Labels phase complete: %d created", result.LabelsCreated)

	// Load default states per project for fallback.
	defaultStates, err := loadDefaultStates(ctx, tx, projectMap)
	if err != nil {
		return nil, fmt.Errorf("loading default states: %w", err)
	}

	// Phase 6: Bulk import tasks.
	taskMap, err := bulkImportTasks(ctx, tx, dump.Issues, projectMap, stateMap, userMap, defaultStates, adminUserID, result)
	if err != nil {
		return nil, fmt.Errorf("importing tasks: %w", err)
	}
	log.Printf("Tasks phase complete: %d created", result.TasksCreated)

	// Phase 7: Bulk import task assignees.
	if err := bulkImportTaskAssignees(ctx, tx, dump.IssueAssignees, taskMap, userMap, projectMap, adminUserID, result); err != nil {
		return nil, fmt.Errorf("importing task assignees: %w", err)
	}
	log.Printf("Assignees phase complete: %d linked", result.AssigneesLinked)

	// Phase 8: Bulk import task labels.
	if err := bulkImportTaskLabels(ctx, tx, dump.IssueLabels, taskMap, labelMap, adminUserID, result); err != nil {
		return nil, fmt.Errorf("importing task labels: %w", err)
	}
	log.Printf("Task labels phase complete: %d assigned", result.LabelsAssigned)

	// Phase 9: Bulk import comments.
	if err := bulkImportComments(ctx, tx, dump.IssueComments, taskMap, userMap, adminUserID, result); err != nil {
		return nil, fmt.Errorf("importing comments: %w", err)
	}
	log.Printf("Comments phase complete: %d created", result.CommentsCreated)

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return result, nil
}

type existingUser struct {
	id       uuid.UUID
	email    string
	username string
}

func loadExistingUsers(ctx context.Context, tx pgx.Tx) (map[string]existingUser, error) {
	rows, err := tx.Query(ctx, "SELECT id, email, username FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make(map[string]existingUser)
	for rows.Next() {
		var u existingUser
		if err := rows.Scan(&u.id, &u.email, &u.username); err != nil {
			return nil, err
		}
		users[u.email] = u
		users[u.username] = u
	}
	return users, rows.Err()
}

func loadExistingProjectKeys(ctx context.Context, tx pgx.Tx) (map[string]bool, error) {
	rows, err := tx.Query(ctx, "SELECT project_key FROM projects WHERE deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make(map[string]bool)
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		keys[k] = true
	}
	return keys, rows.Err()
}

func loadDefaultStates(ctx context.Context, tx pgx.Tx, projectMap map[string]uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	defaults := make(map[uuid.UUID]uuid.UUID)
	for _, projectID := range projectMap {
		var stateID uuid.UUID
		err := tx.QueryRow(ctx,
			"SELECT id FROM project_states WHERE project_id = $1 AND is_default = true LIMIT 1",
			projectID).Scan(&stateID)
		if err == nil {
			defaults[projectID] = stateID
		}
	}
	return defaults, nil
}

// bulkImportUsers creates all new users with a single shared bcrypt hash.
func bulkImportUsers(ctx context.Context, tx pgx.Tx, planeUsers []map[string]string, existing map[string]existingUser, result *ImportResult) (map[string]uuid.UUID, error) {
	userMap := make(map[string]uuid.UUID, len(planeUsers))

	// Generate ONE shared password hash for all imported users.
	randomPass, err := generateRandomPassword()
	if err != nil {
		return nil, err
	}
	sharedHash, err := auth.HashPassword(randomPass)
	if err != nil {
		return nil, err
	}
	log.Printf("Generated shared import password for new users")

	// Collect users to create, dedup against existing.
	type newUser struct {
		planeID   string
		username  string
		email     string
		firstName string
		lastName  string
	}

	var toCreate []newUser
	usedUsernames := make(map[string]bool)
	for _, eu := range existing {
		usedUsernames[eu.username] = true
	}

	for _, pu := range planeUsers {
		if pu["deleted_at"] != "" {
			continue
		}

		planeID := pu["id"]
		email := pu["email"]
		username := pu["username"]

		if email == "" {
			continue
		}

		// Check existing by email or username.
		if eu, ok := existing[email]; ok {
			userMap[planeID] = eu.id
			result.UsersSkipped++
			continue
		}
		if eu, ok := existing[username]; ok {
			userMap[planeID] = eu.id
			result.UsersSkipped++
			continue
		}

		firstName := pu["first_name"]
		lastName := pu["last_name"]
		if firstName == "" && pu["display_name"] != "" {
			parts := strings.SplitN(pu["display_name"], " ", 2)
			firstName = parts[0]
			if len(parts) > 1 {
				lastName = parts[1]
			}
		}
		if firstName == "" {
			firstName = username
		}

		// Ensure unique username.
		uniqueUsername := username
		for i := 2; usedUsernames[uniqueUsername]; i++ {
			uniqueUsername = fmt.Sprintf("%s%d", username, i)
		}
		usedUsernames[uniqueUsername] = true

		toCreate = append(toCreate, newUser{
			planeID:   planeID,
			username:  uniqueUsername,
			email:     email,
			firstName: firstName,
			lastName:  lastName,
		})
	}

	if len(toCreate) == 0 {
		return userMap, nil
	}

	// Bulk insert using COPY protocol for maximum speed.
	inputRows := make([][]interface{}, len(toCreate))
	ids := make([]uuid.UUID, len(toCreate))
	for i, u := range toCreate {
		id := uuid.New()
		ids[i] = id
		inputRows[i] = []interface{}{id, u.username, u.email, sharedHash, u.firstName, u.lastName, "user"}
	}

	_, err = tx.CopyFrom(ctx,
		pgx.Identifier{"users"},
		[]string{"id", "username", "email", "password_hash", "first_name", "last_name", "user_type"},
		pgx.CopyFromRows(inputRows),
	)
	if err != nil {
		return nil, fmt.Errorf("bulk inserting users: %w", err)
	}

	for i, u := range toCreate {
		userMap[u.planeID] = ids[i]
	}
	result.UsersCreated = len(toCreate)

	return userMap, nil
}

// bulkImportProjects creates all projects and admin memberships in bulk.
func bulkImportProjects(ctx context.Context, tx pgx.Tx, planeProjects []map[string]string, adminUserID uuid.UUID, existingKeys map[string]bool, result *ImportResult) (map[string]uuid.UUID, error) {
	projectMap := make(map[string]uuid.UUID, len(planeProjects))

	type newProject struct {
		planeID string
		id      uuid.UUID
		key     string
		name    string
		desc    string
	}

	var projects []newProject
	usedKeys := make(map[string]bool)
	for k := range existingKeys {
		usedKeys[k] = true
	}

	for _, pp := range planeProjects {
		if pp["deleted_at"] != "" {
			continue
		}

		identifier := strings.ToUpper(pp["identifier"])
		if len(identifier) > 10 {
			identifier = identifier[:10]
		}

		uniqueKey := identifier
		for i := 2; usedKeys[uniqueKey]; i++ {
			suffix := strconv.Itoa(i)
			maxLen := 10 - len(suffix)
			if maxLen < 1 {
				maxLen = 1
			}
			base := identifier
			if len(base) > maxLen {
				base = base[:maxLen]
			}
			uniqueKey = base + suffix
		}
		usedKeys[uniqueKey] = true

		id := uuid.New()
		projects = append(projects, newProject{
			planeID: pp["id"],
			id:      id,
			key:     uniqueKey,
			name:    pp["name"],
			desc:    pp["description"],
		})
		log.Printf("Created project: %s (%s)", pp["name"], uniqueKey)
	}

	if len(projects) == 0 {
		return projectMap, nil
	}

	// Bulk insert projects.
	projectRows := make([][]interface{}, len(projects))
	for i, p := range projects {
		var desc *string
		if p.desc != "" {
			desc = &p.desc
		}
		projectRows[i] = []interface{}{p.id, p.key, p.name, desc, nil, nil, adminUserID}
	}

	_, err := tx.CopyFrom(ctx,
		pgx.Identifier{"projects"},
		[]string{"id", "project_key", "name", "description", "icon_id", "cover_id", "created_by"},
		pgx.CopyFromRows(projectRows),
	)
	if err != nil {
		return nil, fmt.Errorf("bulk inserting projects: %w", err)
	}

	// Bulk insert admin as member of each project.
	memberRows := make([][]interface{}, len(projects))
	for i, p := range projects {
		memberRows[i] = []interface{}{p.id, adminUserID, "admin"}
	}

	_, err = tx.CopyFrom(ctx,
		pgx.Identifier{"project_members"},
		[]string{"project_id", "user_id", "role"},
		pgx.CopyFromRows(memberRows),
	)
	if err != nil {
		return nil, fmt.Errorf("bulk inserting admin memberships: %w", err)
	}

	for _, p := range projects {
		projectMap[p.planeID] = p.id
	}
	result.ProjectsCreated = len(projects)

	return projectMap, nil
}

// bulkImportProjectMembers adds all project members in bulk.
func bulkImportProjectMembers(ctx context.Context, tx pgx.Tx, planeMembers []map[string]string, userMap, projectMap map[string]uuid.UUID, adminUserID uuid.UUID) (int, error) {
	// Track existing memberships to avoid duplicates.
	memberSet := make(map[[2]uuid.UUID]bool)

	// Admin is already a member of all projects.
	for _, projectID := range projectMap {
		memberSet[[2]uuid.UUID{projectID, adminUserID}] = true
	}

	var rows [][]interface{}
	for _, pm := range planeMembers {
		if pm["deleted_at"] != "" || pm["is_active"] == "f" {
			continue
		}

		projectID, ok := projectMap[pm["project_id"]]
		if !ok {
			continue
		}
		userID, ok := userMap[pm["member_id"]]
		if !ok {
			continue
		}

		key := [2]uuid.UUID{projectID, userID}
		if memberSet[key] {
			continue
		}
		memberSet[key] = true

		rows = append(rows, []interface{}{projectID, userID, "member"})
	}

	if len(rows) == 0 {
		return 0, nil
	}

	_, err := tx.CopyFrom(ctx,
		pgx.Identifier{"project_members"},
		[]string{"project_id", "user_id", "role"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return 0, fmt.Errorf("bulk inserting project members: %w", err)
	}

	return len(rows), nil
}

// bulkImportStates creates all project states in bulk.
func bulkImportStates(ctx context.Context, tx pgx.Tx, planeStates []map[string]string, projectMap map[string]uuid.UUID, result *ImportResult) (map[string]uuid.UUID, error) {
	stateMap := make(map[string]uuid.UUID, len(planeStates))

	// Sort by sequence for ordering and default assignment.
	sort.Slice(planeStates, func(i, j int) bool {
		si, _ := strconv.ParseFloat(planeStates[i]["sequence"], 64)
		sj, _ := strconv.ParseFloat(planeStates[j]["sequence"], 64)
		return si < sj
	})

	projectHasDefault := make(map[uuid.UUID]bool)
	// Track (project_id, name) to handle duplicates.
	stateNameSet := make(map[[2]string]uuid.UUID)

	var rows [][]interface{}

	for _, ps := range planeStates {
		if ps["deleted_at"] != "" {
			continue
		}

		projectID, ok := projectMap[ps["project_id"]]
		if !ok {
			continue
		}

		name := ps["name"]
		color := ps["color"]
		// Truncate color to 50 chars.
		if len(color) > 50 {
			color = color[:50]
		}

		stateType := mapStateGroup(ps["group"])
		seq, _ := strconv.ParseFloat(ps["sequence"], 64)
		position := int32(seq / 1000)

		// Dedup by (project_id, name).
		nameKey := [2]string{projectID.String(), name}
		if existingID, exists := stateNameSet[nameKey]; exists {
			stateMap[ps["id"]] = existingID
			continue
		}

		isDefault := !projectHasDefault[projectID]
		if isDefault {
			projectHasDefault[projectID] = true
		}

		id := uuid.New()
		stateNameSet[nameKey] = id
		stateMap[ps["id"]] = id

		var colorPtr *string
		if color != "" {
			colorPtr = &color
		}
		rows = append(rows, []interface{}{id, projectID, stateType, name, colorPtr, position, isDefault})
	}

	if len(rows) == 0 {
		return stateMap, nil
	}

	_, err := tx.CopyFrom(ctx,
		pgx.Identifier{"project_states"},
		[]string{"id", "project_id", "state_type", "name", "color", "position", "is_default"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return nil, fmt.Errorf("bulk inserting states: %w", err)
	}

	result.StatesCreated = len(rows)
	return stateMap, nil
}

// bulkImportLabels creates all project labels in bulk.
func bulkImportLabels(ctx context.Context, tx pgx.Tx, planeLabels []map[string]string, projectMap map[string]uuid.UUID, result *ImportResult) (map[string]uuid.UUID, error) {
	labelMap := make(map[string]uuid.UUID, len(planeLabels))
	// Track (project_id, name) for dedup.
	labelNameSet := make(map[[2]string]uuid.UUID)

	var rows [][]interface{}

	for _, pl := range planeLabels {
		if pl["deleted_at"] != "" {
			continue
		}

		projectIDStr := pl["project_id"]
		if projectIDStr == "" {
			continue
		}

		projectID, ok := projectMap[projectIDStr]
		if !ok {
			continue
		}

		name := pl["name"]
		if len(name) > 50 {
			name = name[:50]
		}
		color := pl["color"]
		if len(color) > 50 {
			color = color[:50]
		}

		nameKey := [2]string{projectID.String(), name}
		if existingID, exists := labelNameSet[nameKey]; exists {
			labelMap[pl["id"]] = existingID
			continue
		}

		id := uuid.New()
		labelNameSet[nameKey] = id
		labelMap[pl["id"]] = id

		var colorPtr *string
		if color != "" {
			colorPtr = &color
		}
		rows = append(rows, []interface{}{id, projectID, name, colorPtr})
	}

	if len(rows) == 0 {
		return labelMap, nil
	}

	_, err := tx.CopyFrom(ctx,
		pgx.Identifier{"project_labels"},
		[]string{"id", "project_id", "name", "color"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return nil, fmt.Errorf("bulk inserting labels: %w", err)
	}

	result.LabelsCreated = len(rows)
	return labelMap, nil
}

// bulkImportTasks creates all tasks in bulk.
func bulkImportTasks(ctx context.Context, tx pgx.Tx, planeIssues []map[string]string, projectMap, stateMap, userMap map[string]uuid.UUID, defaultStates map[uuid.UUID]uuid.UUID, adminUserID uuid.UUID, result *ImportResult) (map[string]uuid.UUID, error) {
	taskMap := make(map[string]uuid.UUID, len(planeIssues))

	// Sort by sequence_id for consistent ordering.
	sort.Slice(planeIssues, func(i, j int) bool {
		si, _ := strconv.Atoi(planeIssues[i]["sequence_id"])
		sj, _ := strconv.Atoi(planeIssues[j]["sequence_id"])
		return si < sj
	})

	// Get next task numbers for existing projects.
	nextTaskNum := make(map[uuid.UUID]int32)
	for _, projectID := range projectMap {
		var maxNum *int32
		err := tx.QueryRow(ctx,
			"SELECT MAX(task_number) FROM tasks WHERE project_id = $1", projectID).Scan(&maxNum)
		if err != nil {
			return nil, fmt.Errorf("getting max task number: %w", err)
		}
		if maxNum != nil {
			nextTaskNum[projectID] = *maxNum + 1
		} else {
			nextTaskNum[projectID] = 1
		}
	}

	var rows [][]interface{}

	for _, pi := range planeIssues {
		if pi["deleted_at"] != "" || pi["is_draft"] == "t" || pi["archived_at"] != "" {
			continue
		}

		projectID, ok := projectMap[pi["project_id"]]
		if !ok {
			continue
		}

		stateID, ok := stateMap[pi["state_id"]]
		if !ok {
			stateID, ok = defaultStates[projectID]
			if !ok {
				continue
			}
		}

		createdBy, ok := userMap[pi["created_by_id"]]
		if !ok {
			createdBy = adminUserID
		}

		priority := mapPriority(pi["priority"])
		title := pi["name"]
		if len(title) > 500 {
			title = title[:500]
		}
		description := pi["description_stripped"]

		taskNum := nextTaskNum[projectID]
		nextTaskNum[projectID] = taskNum + 1

		id := uuid.New()
		taskMap[pi["id"]] = id

		var descPtr *string
		if description != "" {
			descPtr = &description
		}
		rows = append(rows, []interface{}{id, projectID, taskNum, title, descPtr, stateID, priority, createdBy})
	}

	if len(rows) == 0 {
		return taskMap, nil
	}

	_, err := tx.CopyFrom(ctx,
		pgx.Identifier{"tasks"},
		[]string{"id", "project_id", "task_number", "title", "description", "state_id", "priority", "created_by"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return nil, fmt.Errorf("bulk inserting tasks: %w", err)
	}

	result.TasksCreated = len(rows)
	return taskMap, nil
}

// bulkImportTaskAssignees links assignees to tasks in bulk.
// Also ensures users are project members first.
func bulkImportTaskAssignees(ctx context.Context, tx pgx.Tx, planeAssignees []map[string]string, taskMap, userMap, projectMap map[string]uuid.UUID, adminUserID uuid.UUID, result *ImportResult) error {
	// First, load existing memberships.
	existingMembers := make(map[[2]uuid.UUID]bool)
	memberRows, err := tx.Query(ctx, "SELECT project_id, user_id FROM project_members")
	if err != nil {
		return fmt.Errorf("loading memberships: %w", err)
	}
	for memberRows.Next() {
		var pid, uid uuid.UUID
		if err := memberRows.Scan(&pid, &uid); err != nil {
			memberRows.Close()
			return err
		}
		existingMembers[[2]uuid.UUID{pid, uid}] = true
	}
	memberRows.Close()

	// Collect new memberships needed for assignees.
	var newMemberRows [][]interface{}
	assigneeSet := make(map[[2]uuid.UUID]bool)
	var assigneeRows [][]interface{}

	for _, pa := range planeAssignees {
		if pa["deleted_at"] != "" {
			continue
		}

		taskID, ok := taskMap[pa["issue_id"]]
		if !ok {
			continue
		}
		userID, ok := userMap[pa["assignee_id"]]
		if !ok {
			continue
		}

		// Ensure project membership.
		if projectID, ok := projectMap[pa["project_id"]]; ok {
			mKey := [2]uuid.UUID{projectID, userID}
			if !existingMembers[mKey] {
				existingMembers[mKey] = true
				newMemberRows = append(newMemberRows, []interface{}{projectID, userID, "member"})
			}
		}

		// Dedup assignees.
		aKey := [2]uuid.UUID{taskID, userID}
		if assigneeSet[aKey] {
			continue
		}
		assigneeSet[aKey] = true

		assigneeRows = append(assigneeRows, []interface{}{taskID, userID, adminUserID})
	}

	// Bulk insert new memberships.
	if len(newMemberRows) > 0 {
		_, err := tx.CopyFrom(ctx,
			pgx.Identifier{"project_members"},
			[]string{"project_id", "user_id", "role"},
			pgx.CopyFromRows(newMemberRows),
		)
		if err != nil {
			return fmt.Errorf("bulk inserting assignee memberships: %w", err)
		}
	}

	// Bulk insert assignees.
	if len(assigneeRows) > 0 {
		_, err := tx.CopyFrom(ctx,
			pgx.Identifier{"task_assignees"},
			[]string{"task_id", "user_id", "assigned_by"},
			pgx.CopyFromRows(assigneeRows),
		)
		if err != nil {
			return fmt.Errorf("bulk inserting task assignees: %w", err)
		}
	}

	result.AssigneesLinked = len(assigneeRows)
	return nil
}

// bulkImportTaskLabels links labels to tasks in bulk.
func bulkImportTaskLabels(ctx context.Context, tx pgx.Tx, planeIssueLabels []map[string]string, taskMap, labelMap map[string]uuid.UUID, adminUserID uuid.UUID, result *ImportResult) error {
	labelSet := make(map[[2]uuid.UUID]bool)
	var rows [][]interface{}

	for _, pil := range planeIssueLabels {
		if pil["deleted_at"] != "" {
			continue
		}

		taskID, ok := taskMap[pil["issue_id"]]
		if !ok {
			continue
		}
		labelID, ok := labelMap[pil["label_id"]]
		if !ok {
			continue
		}

		key := [2]uuid.UUID{taskID, labelID}
		if labelSet[key] {
			continue
		}
		labelSet[key] = true

		rows = append(rows, []interface{}{taskID, labelID, adminUserID})
	}

	if len(rows) == 0 {
		return nil
	}

	_, err := tx.CopyFrom(ctx,
		pgx.Identifier{"task_labels"},
		[]string{"task_id", "label_id", "added_by"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("bulk inserting task labels: %w", err)
	}

	result.LabelsAssigned = len(rows)
	return nil
}

// bulkImportComments creates all comments in bulk.
func bulkImportComments(ctx context.Context, tx pgx.Tx, planeComments []map[string]string, taskMap, userMap map[string]uuid.UUID, adminUserID uuid.UUID, result *ImportResult) error {
	var rows [][]interface{}

	for _, pc := range planeComments {
		if pc["deleted_at"] != "" {
			continue
		}

		taskID, ok := taskMap[pc["issue_id"]]
		if !ok {
			continue
		}

		content := pc["comment_stripped"]
		if content == "" {
			continue
		}

		createdBy, ok := userMap[pc["actor_id"]]
		if !ok {
			createdBy = adminUserID
		}

		rows = append(rows, []interface{}{taskID, content, createdBy})
	}

	if len(rows) == 0 {
		return nil
	}

	_, err := tx.CopyFrom(ctx,
		pgx.Identifier{"comments"},
		[]string{"task_id", "content", "created_by"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("bulk inserting comments: %w", err)
	}

	result.CommentsCreated = len(rows)
	return nil
}

// mapPriority converts Plane priority strings to Bureaucat priority ints.
func mapPriority(planePriority string) int32 {
	switch planePriority {
	case "urgent":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// mapStateGroup converts Plane state group to Bureaucat state_type.
func mapStateGroup(group string) string {
	switch group {
	case "backlog", "unstarted", "started", "completed", "cancelled":
		return group
	default:
		return "backlog"
	}
}

// generateRandomPassword creates a random 32-byte hex password.
func generateRandomPassword() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
