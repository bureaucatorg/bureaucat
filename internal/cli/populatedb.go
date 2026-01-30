package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"bereaucat/internal/auth"
	"bereaucat/internal/server"
	"bereaucat/internal/store"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/urfave/cli/v3"
)

// apiClient is a thin wrapper around net/http.Client that carries a base URL
// and an optional Bearer token for authenticated requests.
type apiClient struct {
	baseURL string
	token   string
	http    *http.Client
}

func (c *apiClient) do(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	return c.http.Do(req)
}

func (c *apiClient) post(path string, body interface{}) (*http.Response, error) {
	return c.do("POST", path, body)
}

func (c *apiClient) patch(path string, body interface{}) (*http.Response, error) {
	return c.do("PATCH", path, body)
}

func (c *apiClient) get(path string) (*http.Response, error) {
	return c.do("GET", path, nil)
}

// decodeJSON reads the response body and decodes it into dst.
func decodeJSON(resp *http.Response, dst interface{}) error {
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

// discardBody reads and closes the response body, returning an error if the
// status code indicates failure.
func discardBody(resp *http.Response) error {
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

// --- Response types (only the fields we need) ---

type signinResponse struct {
	User        userResponse `json:"user"`
	AccessToken string       `json:"access_token"`
}

type userResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

type projectResponse struct {
	ID         uuid.UUID `json:"id"`
	ProjectKey string    `json:"project_key"`
}

type stateResponse struct {
	ID        uuid.UUID `json:"id"`
	StateType string    `json:"state_type"`
	Name      string    `json:"name"`
}

type labelResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type taskResponse struct {
	ID         uuid.UUID `json:"id"`
	TaskNumber int       `json:"task_number"`
}

func PopulateDBCommand() *cli.Command {
	return &cli.Command{
		Name:  "populatedb",
		Usage: "Populate the database with sample demo data",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "database-url",
				Usage:   "Database connection URL",
				Sources: cli.EnvVars("DATABASE_URL"),
			},
			&cli.StringFlag{
				Name:    "demo-user-email",
				Usage:   "Email for an additional demo admin user",
				Sources: cli.EnvVars("DEMO_USER_EMAIL"),
			},
			&cli.StringFlag{
				Name:    "demo-user-password",
				Usage:   "Password for the additional demo admin user",
				Sources: cli.EnvVars("DEMO_USER_PASSWORD"),
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			dbURL := cmd.String("database-url")
			if dbURL == "" {
				return fmt.Errorf("database-url is required (use --database-url or DATABASE_URL env var)")
			}

			// Boot the real server with all handlers and middleware.
			srv, err := server.New(true, dbURL, server.AuthConfig{
				JWTSecret:              "populatedb-dummy-jwt-secret-key-32chars!",
				AccessTokenExpiryMins:  60, // generous expiry for seeding
				RefreshTokenExpiryDays: 1,
			}, nil)
			if err != nil {
				return fmt.Errorf("failed to create server: %w", err)
			}
			defer srv.Close()

			// Wrap with httptest so we get a real base URL without picking a port.
			ts := httptest.NewServer(srv.Echo())
			defer ts.Close()

			client := &apiClient{
				baseURL: ts.URL,
				http:    ts.Client(),
			}

			fmt.Println("Populating database with demo data...")

			// -----------------------------------------------------------------
			// 1. Seed the admin user directly via store (can't sign in without
			//    a user, and signup doesn't let us set user_type=admin explicitly
			//    when there could be race conditions with the "first user" check).
			// -----------------------------------------------------------------
			fmt.Println("Creating admin user...")

			pool, err := pgxpool.New(ctx, dbURL)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer pool.Close()

			q := store.New(pool)

			adminHash, err := auth.HashPassword("admin123")
			if err != nil {
				return fmt.Errorf("failed to hash password: %w", err)
			}

			_, err = q.CreateUser(ctx, store.CreateUserParams{
				Username:     "admin",
				Email:        "admin@example.com",
				PasswordHash: adminHash,
				FirstName:    "Ada",
				LastName:     "Lovelace",
				UserType:     "admin",
			})
			if err != nil {
				return fmt.Errorf("failed to create admin user: %w", err)
			}
			fmt.Println("  Created admin user: admin (admin@example.com)")

			// -----------------------------------------------------------------
			// 2. Sign in as admin to get an access token.
			// -----------------------------------------------------------------
			fmt.Println("Signing in as admin...")

			resp, err := client.post("/api/v1/signin", map[string]string{
				"identifier": "admin@example.com",
				"password":   "admin123",
			})
			if err != nil {
				return fmt.Errorf("failed to sign in: %w", err)
			}
			var adminAuth signinResponse
			if err := decodeJSON(resp, &adminAuth); err != nil {
				return fmt.Errorf("failed to decode signin response: %w", err)
			}
			client.token = adminAuth.AccessToken

			// Track user IDs: username -> uuid
			userIDs := map[string]uuid.UUID{
				"admin": adminAuth.User.ID,
			}

			// -----------------------------------------------------------------
			// 3. Create remaining users via admin endpoint.
			// -----------------------------------------------------------------
			fmt.Println("Creating users...")

			type userSpec struct {
				username, email, first, last, password, userType string
			}
			demoUsers := []userSpec{
				{"alice", "alice@example.com", "Alice", "Zhang", "user123", "user"},
				{"bob", "bob@example.com", "Bob", "Patel", "user123", "user"},
				{"charlie", "charlie@example.com", "Charlie", "Okonkwo", "user123", "user"},
				{"diana", "diana@example.com", "Diana", "Müller", "user123", "user"},
			}

			for _, u := range demoUsers {
				resp, err := client.post("/api/v1/admin/users", map[string]string{
					"username":   u.username,
					"email":      u.email,
					"password":   u.password,
					"first_name": u.first,
					"last_name":  u.last,
					"user_type":  u.userType,
				})
				if err != nil {
					return fmt.Errorf("failed to create user %s: %w", u.username, err)
				}
				var created userResponse
				if err := decodeJSON(resp, &created); err != nil {
					return fmt.Errorf("failed to decode create user response for %s: %w", u.username, err)
				}
				userIDs[u.username] = created.ID
				fmt.Printf("  Created user: %s (%s)\n", u.username, u.email)
			}

			// Create additional demo admin user if env vars are set.
			demoEmail := cmd.String("demo-user-email")
			demoPassword := cmd.String("demo-user-password")
			if demoEmail != "" && demoPassword != "" {
				resp, err := client.post("/api/v1/admin/users", map[string]string{
					"username":   "demo",
					"email":      demoEmail,
					"password":   demoPassword,
					"first_name": "Demo",
					"last_name":  "Admin",
					"user_type":  "admin",
				})
				if err != nil {
					return fmt.Errorf("failed to create demo admin user: %w", err)
				}
				var created userResponse
				if err := decodeJSON(resp, &created); err != nil {
					return fmt.Errorf("failed to decode demo user response: %w", err)
				}
				userIDs["demo"] = created.ID
				fmt.Printf("  Created demo admin user: demo (%s)\n", demoEmail)
			}

			// -----------------------------------------------------------------
			// 4. Create projects via handler (auto-creates default states + adds
			//    creator as admin member).
			// -----------------------------------------------------------------
			fmt.Println("Creating projects...")

			type demoProject struct {
				key, name, desc string
			}
			projects := []demoProject{
				{"INFRA", "Infrastructure Access", "Requests for access to cloud infrastructure, servers, and deployment environments."},
				{"DATA", "Data & Analytics Access", "Requests for access to databases, data warehouses, and analytics platforms."},
				{"SEC", "Security & Compliance", "Security tool access, compliance audit requests, and certificate management."},
			}

			for _, p := range projects {
				desc := p.desc
				resp, err := client.post("/api/v1/projects", map[string]interface{}{
					"project_key": p.key,
					"name":        p.name,
					"description": desc,
				})
				if err != nil {
					return fmt.Errorf("failed to create project %s: %w", p.key, err)
				}
				var created projectResponse
				if err := decodeJSON(resp, &created); err != nil {
					return fmt.Errorf("failed to decode project response for %s: %w", p.key, err)
				}
				fmt.Printf("  Created project: %s - %s\n", p.key, p.name)
			}

			// -----------------------------------------------------------------
			// 5. Add project members.
			//    Admin is already a member (added by CreateProject handler).
			// -----------------------------------------------------------------
			fmt.Println("Adding project members...")

			type memberSpec struct {
				projKey  string
				username string
				role     string
			}
			members := []memberSpec{
				// INFRA: all users
				{"INFRA", "alice", "member"},
				{"INFRA", "bob", "member"},
				{"INFRA", "charlie", "member"},
				{"INFRA", "diana", "member"},
				// DATA: alice, bob
				{"DATA", "alice", "member"},
				{"DATA", "bob", "member"},
				// SEC: charlie, diana
				{"SEC", "charlie", "member"},
				{"SEC", "diana", "member"},
			}

			for _, m := range members {
				resp, err := client.post(
					fmt.Sprintf("/api/v1/projects/%s/members", m.projKey),
					map[string]string{
						"user_id": userIDs[m.username].String(),
						"role":    m.role,
					},
				)
				if err != nil {
					return fmt.Errorf("failed to add %s to %s: %w", m.username, m.projKey, err)
				}
				if err := discardBody(resp); err != nil {
					return fmt.Errorf("failed to add %s to %s: %w", m.username, m.projKey, err)
				}
			}

			// -----------------------------------------------------------------
			// 6. Fetch auto-created states for each project.
			// -----------------------------------------------------------------
			fmt.Println("Fetching project states...")

			// stateByType: project_key -> state_type -> first stateResponse of that type.
			// Some state types have multiple states (e.g. "unstarted" has "Todo"
			// and "Approval Pending"). We pick the first for transitions.
			stateByType := make(map[string]map[string]stateResponse)

			for _, projKey := range []string{"INFRA", "DATA", "SEC"} {
				resp, err := client.get(fmt.Sprintf("/api/v1/projects/%s/states", projKey))
				if err != nil {
					return fmt.Errorf("failed to fetch states for %s: %w", projKey, err)
				}
				var states []stateResponse
				if err := decodeJSON(resp, &states); err != nil {
					return fmt.Errorf("failed to decode states for %s: %w", projKey, err)
				}
				stateByType[projKey] = make(map[string]stateResponse)
				for _, s := range states {
					if _, exists := stateByType[projKey][s.StateType]; !exists {
						stateByType[projKey][s.StateType] = s
					}
				}
			}

			// -----------------------------------------------------------------
			// 7. Create labels for each project.
			// -----------------------------------------------------------------
			fmt.Println("Creating project labels...")

			type labelSpec struct {
				name  string
				color string
			}
			defaultLabels := []labelSpec{
				{"New Access", "#3B82F6"},
				{"Escalation", "#EF4444"},
				{"Renewal", "#8B5CF6"},
				{"Revocation", "#6B7280"},
				{"Urgent", "#F97316"},
			}

			// labelMap: project_key -> label_name -> labelResponse
			labelMap := make(map[string]map[string]labelResponse)

			for _, projKey := range []string{"INFRA", "DATA", "SEC"} {
				labelMap[projKey] = make(map[string]labelResponse)
				for _, l := range defaultLabels {
					resp, err := client.post(
						fmt.Sprintf("/api/v1/projects/%s/labels", projKey),
						map[string]string{
							"name":  l.name,
							"color": l.color,
						},
					)
					if err != nil {
						return fmt.Errorf("failed to create label %s for %s: %w", l.name, projKey, err)
					}
					var created labelResponse
					if err := decodeJSON(resp, &created); err != nil {
						return fmt.Errorf("failed to decode label response: %w", err)
					}
					labelMap[projKey][l.name] = created
				}
			}

			// -----------------------------------------------------------------
			// 8. Create task templates for each project.
			// -----------------------------------------------------------------
			fmt.Println("Creating task templates...")

			type templateSpec struct {
				name string
				title string
				desc string
			}

			projectTemplates := map[string][]templateSpec{
				"INFRA": {
					{"Cloud Access Request", "Requesting access to [service/resource]", "## Details\n- **Cloud Provider:** AWS / Azure / GCP\n- **Resource:** \n- **Role/Permission:** \n- **Justification:** \n\n## Duration\n- [ ] Permanent\n- [ ] Temporary (specify end date): "},
					{"SSH Key Rotation", "Rotate SSH keys for [host/service]", "## Details\n- **Target hosts:** \n- **Reason for rotation:** \n- **Maintenance window:** \n\n## Checklist\n- [ ] Identify all affected keys\n- [ ] Schedule maintenance window\n- [ ] Rotate keys\n- [ ] Verify access\n- [ ] Remove old keys"},
				},
				"DATA": {
					{"Database Access Request", "Requesting access to [database/schema]", "## Details\n- **Database:** \n- **Schema/Table:** \n- **Permission level:** READ / WRITE / ADMIN\n- **Justification:** \n\n## Compliance\n- [ ] Security training completed\n- [ ] DPA signed (if PII involved)"},
					{"Analytics Tool License", "Requesting [tool] license for [team/purpose]", "## Details\n- **Tool:** \n- **License type:** \n- **Number of seats:** \n- **Business justification:** \n\n## Cost\n- Estimated cost per seat: \n- Budget code: "},
				},
				"SEC": {
					{"Security Tool Access", "Requesting access to [security tool]", "## Details\n- **Tool:** \n- **Access level:** \n- **Justification:** \n\n## Context\n- Related incident (if any): \n- Time sensitivity: Normal / Urgent"},
					{"Penetration Test Request", "Q[N] penetration test for [target]", "## Scope\n- **Target systems:** \n- **Testing type:** Black box / Grey box / White box\n- **Environment:** Staging / Production (read-only)\n\n## Prerequisites\n- [ ] Rules of engagement signed\n- [ ] Scope document approved\n- [ ] Notification sent to affected teams"},
				},
			}

			for projKey, templates := range projectTemplates {
				for _, t := range templates {
					resp, err := client.post(
						fmt.Sprintf("/api/v1/projects/%s/templates", projKey),
						map[string]string{
							"name":        t.name,
							"title":       t.title,
							"description": t.desc,
						},
					)
					if err != nil {
						return fmt.Errorf("failed to create template %s for %s: %w", t.name, projKey, err)
					}
					if err := discardBody(resp); err != nil {
						return fmt.Errorf("failed to create template %s for %s: %w", t.name, projKey, err)
					}
					fmt.Printf("  [%s] %s\n", projKey, t.name)
				}
			}

			// -----------------------------------------------------------------
			// 9. Helper: sign in as a specific user and return the token.
			// -----------------------------------------------------------------
			tokenCache := map[string]string{
				"admin": client.token,
			}

			signinAs := func(username string) (string, error) {
				if tok, ok := tokenCache[username]; ok {
					return tok, nil
				}
				resp, err := client.post("/api/v1/signin", map[string]string{
					"identifier": username,
					"password":   "user123",
				})
				if err != nil {
					return "", fmt.Errorf("failed to sign in as %s: %w", username, err)
				}
				var auth signinResponse
				if err := decodeJSON(resp, &auth); err != nil {
					return "", fmt.Errorf("failed to decode signin for %s: %w", username, err)
				}
				tokenCache[username] = auth.AccessToken
				return auth.AccessToken, nil
			}

			// -----------------------------------------------------------------
			// 10. Create tasks via handler (handler logs activity for creation,
			//     assignees, and labels automatically).
			// -----------------------------------------------------------------
			fmt.Println("Creating tasks...")

			type taskSpec struct {
				projKey    string
				title      string
				desc       string
				finalState string // the state_type the task should end up in
				priority   int
				creator    string
				assignees  []string
				labels     []string
				daysAgo    int // how many days in the past the task was "created"
			}

			tasks := []taskSpec{
				// INFRA project
				{"INFRA", "Seeking Azure access for Disaster Recovery setup", "Need Contributor role on the Azure DR resource group to configure geo-replicated failover for the payments service.", "started", 0, "alice", []string{"alice"}, []string{"New Access", "Urgent"}, 3},
				{"INFRA", "Request AWS IAM role for production deployments", "Requesting AssumeRole permissions on prod-deploy IAM role to run CI/CD pipelines against the production EKS cluster.", "started", 1, "bob", []string{"bob"}, []string{"New Access"}, 5},
				{"INFRA", "Requesting GCP project owner access for billing migration", "Need temporary Owner access to migrate billing accounts between GCP projects. Access to be revoked after migration.", "unstarted", 2, "charlie", []string{"charlie"}, []string{"Escalation"}, 1},
				{"INFRA", "Renew Kubernetes cluster admin certificate", "Current cluster admin certificate expires in 14 days. Requesting renewal for another 12 months.", "backlog", 3, "diana", []string{}, []string{"Renewal"}, 12},
				{"INFRA", "Seeking VPN access to staging environment", "Need site-to-site VPN credentials to access staging servers for integration testing of the new checkout flow.", "unstarted", 2, "alice", []string{"alice"}, []string{"New Access"}, 2},
				{"INFRA", "Revoke former contractor SSH keys from jump hosts", "Contractor engagement ended last Friday. All SSH keys for user j.smith need to be removed from bastion and jump hosts.", "unstarted", 0, "admin", []string{"bob"}, []string{"Revocation", "Urgent"}, 1},
				{"INFRA", "Request Terraform Cloud workspace access", "Need write access to the infra-prod workspace in Terraform Cloud to apply changes for the database scaling project.", "completed", 1, "bob", []string{"bob"}, []string{"New Access"}, 14},
				{"INFRA", "Requesting root access to legacy on-prem servers", "Need temporary root access to legacy on-prem servers for the data center decommission. Strictly time-boxed to 2 weeks.", "cancelled", 0, "charlie", []string{"charlie"}, []string{"Escalation"}, 10},

				// DATA project
				{"DATA", "Seeking read access to production Postgres replica", "Need SELECT permissions on the analytics replica to run quarterly revenue reports. Read-only access sufficient.", "completed", 1, "admin", []string{"admin", "alice"}, []string{"New Access"}, 21},
				{"DATA", "Request Snowflake warehouse access for ML training", "Requesting access to the COMPUTE_XL warehouse to run feature engineering queries for the churn prediction model.", "started", 0, "alice", []string{"alice"}, []string{"New Access"}, 4},
				{"DATA", "Seeking Redshift access for customer segmentation", "Need read access to the customer_events and transactions schemas in Redshift for building segmentation dashboards.", "unstarted", 2, "bob", []string{"bob"}, []string{"New Access"}, 2},
				{"DATA", "Renew Looker admin license before Q4", "Current Looker admin license expires end of September. Requesting renewal and addition of 5 viewer seats.", "backlog", 3, "admin", []string{}, []string{"Renewal"}, 30},
				{"DATA", "Request access to PII dataset for fraud analysis", "Need access to unmasked PII fields in the fraud_detection schema. DPA and security training certificates attached.", "unstarted", 0, "bob", []string{"bob"}, []string{"New Access", "Urgent"}, 1},

				// SEC project
				{"SEC", "Seeking Vault admin access for secrets rotation", "Need admin access to HashiCorp Vault to implement automated secrets rotation for production database credentials.", "completed", 1, "admin", []string{"admin", "charlie"}, []string{"New Access"}, 18},
				{"SEC", "Request SentinelOne console access for incident response", "IR team member needs access to the SentinelOne EDR console to investigate the alert triggered on web-prod-03.", "started", 0, "charlie", []string{"charlie"}, []string{"New Access", "Urgent"}, 1},
				{"SEC", "Seeking Cloudflare WAF access for rule updates", "Need editor access to Cloudflare WAF to deploy updated OWASP rules and add custom rate limiting for the /api/v1/auth endpoints.", "unstarted", 2, "diana", []string{"diana"}, []string{"New Access"}, 3},
				{"SEC", "Request access to penetration testing tools", "Requesting access to Burp Suite Pro license and internal pentest VPN for the scheduled Q4 security assessment.", "started", 1, "charlie", []string{"charlie", "diana"}, []string{"New Access"}, 7},
				{"SEC", "Revoke SOC analyst access after role transfer", "SOC analyst transferred to engineering. Remove access to Splunk, CrowdStrike, and PagerDuty SOC dashboards.", "backlog", 2, "admin", []string{}, []string{"Revocation"}, 9},
			}

			// State transition paths: to reach a target state_type from the
			// default state (Backlog, state_type=backlog), tasks move through
			// these intermediate state types. The handler logs state_changed
			// activity automatically on each PATCH.
			//
			// The default state created by CreateProject is "Backlog" (backlog).
			// Tasks are created in whatever state we pass, but to generate
			// realistic activity we create them in default state and then
			// transition.
			stateTransitions := map[string][]string{
				"backlog":   {},                          // already there (default)
				"unstarted": {"unstarted"},               // backlog -> unstarted
				"started":   {"unstarted", "started"},    // backlog -> unstarted -> started
				"completed": {"unstarted", "started", "completed"},
				"cancelled": {"unstarted", "started", "cancelled"},
			}

			// taskRefs tracks task info for comment references and backdating.
			type taskRef struct {
				projKey    string
				taskNumber int
				taskID     uuid.UUID
				daysAgo    int
			}
			taskRefs := make(map[string]taskRef)

			for i, t := range tasks {
				// Sign in as the creator so the task is created under their identity.
				tok, err := signinAs(t.creator)
				if err != nil {
					return err
				}
				client.token = tok

				// Resolve assignee UUIDs.
				assigneeIDs := make([]string, len(t.assignees))
				for j, a := range t.assignees {
					assigneeIDs[j] = userIDs[a].String()
				}

				// Resolve label UUIDs.
				labelIDs := make([]string, len(t.labels))
				for j, l := range t.labels {
					labelIDs[j] = labelMap[t.projKey][l].ID.String()
				}

				// Find the default state ID (Backlog) — tasks start here.
				defaultState := stateByType[t.projKey]["backlog"]

				desc := t.desc
				priority := t.priority
				stateID := defaultState.ID.String()
				resp, err := client.post(
					fmt.Sprintf("/api/v1/projects/%s/tasks", t.projKey),
					map[string]interface{}{
						"title":       t.title,
						"description": &desc,
						"state_id":    &stateID,
						"priority":    &priority,
						"assignees":   assigneeIDs,
						"labels":      labelIDs,
					},
				)
				if err != nil {
					return fmt.Errorf("failed to create task %q: %w", t.title, err)
				}
				var created taskResponse
				if err := decodeJSON(resp, &created); err != nil {
					return fmt.Errorf("failed to decode task response for %q: %w", t.title, err)
				}

				key := fmt.Sprintf("%s-%d", t.projKey, i)
				taskRefs[key] = taskRef{
					projKey:    t.projKey,
					taskNumber: created.TaskNumber,
					taskID:     created.ID,
					daysAgo:    t.daysAgo,
				}

				// Small delay so activity timestamps are distinct.
				time.Sleep(time.Millisecond)

				// Simulate state transitions (handler logs state_changed activity).
				transitions := stateTransitions[t.finalState]
				for _, nextStateType := range transitions {
					nextState := stateByType[t.projKey][nextStateType]
					stID := nextState.ID.String()
					resp, err := client.patch(
						fmt.Sprintf("/api/v1/projects/%s/tasks/%d", t.projKey, created.TaskNumber),
						map[string]interface{}{
							"state_id": &stID,
						},
					)
					if err != nil {
						return fmt.Errorf("failed to transition task %q to %s: %w", t.title, nextStateType, err)
					}
					if err := discardBody(resp); err != nil {
						return fmt.Errorf("failed to transition task %q to %s: %w", t.title, nextStateType, err)
					}
					time.Sleep(time.Millisecond)
				}

				fmt.Printf("  [%s-%d] %s\n", t.projKey, created.TaskNumber, t.title)
			}

			// -----------------------------------------------------------------
			// 11. Create comments (handler logs comment_created activity).
			// -----------------------------------------------------------------
			fmt.Println("Creating comments...")

			type commentSpec struct {
				taskKey string
				author  string
				content string
				daysAgo int // how many days ago this comment was posted
			}

			comments := []commentSpec{
				// INFRA-0: Azure access for DR setup (3 days ago)
				{"INFRA-0", "admin", "Which resource group specifically? We have three DR groups across regions.", 2},
				{"INFRA-0", "alice", "The rg-payments-dr-westus2 group. I've attached the architecture diagram for context.", 2},
				{"INFRA-0", "bob", "I can vouch for this — the failover testing is blocking our SLA compliance deadline.", 1},

				// INFRA-2: GCP project owner access (1 day ago)
				{"INFRA-2", "admin", "Owner access is a high privilege. Can we scope this down to Billing Account Administrator instead?", 1},
				{"INFRA-2", "charlie", "Billing Account Administrator won't let me reassign the linked project. I need Owner for the migration step only.", 1},
				{"INFRA-2", "diana", "Suggest we grant time-boxed access with a calendar reminder for revocation.", 0},

				// INFRA-5: Revoke contractor SSH keys (1 day ago)
				{"INFRA-5", "bob", "I've identified 4 SSH keys across 3 jump hosts. Will need a maintenance window to rotate host keys too.", 1},
				{"INFRA-5", "admin", "Good call. Schedule the rotation for this weekend's maintenance window.", 0},

				// INFRA-6: Terraform Cloud workspace access (14 days ago)
				{"INFRA-6", "bob", "Access granted. Confirmed I can plan and apply in the infra-prod workspace.", 12},
				{"INFRA-6", "admin", "Approved. Please follow the change management process for any production applies.", 11},

				// INFRA-7: Root access to legacy on-prem (10 days ago)
				{"INFRA-7", "charlie", "The decommission requires modifying boot configs that only root can touch.", 9},
				{"INFRA-7", "admin", "Rejected — root access on production on-prem violates our compliance policy. Let's use the managed automation runbooks instead.", 8},

				// DATA-8: Read access to production Postgres replica (21 days ago)
				{"DATA-8", "alice", "Access configured. Connected successfully and ran a test query against the replica.", 19},
				{"DATA-8", "admin", "Approved. Remember to use the read replica endpoint, not the primary.", 18},

				// DATA-9: Snowflake warehouse access (4 days ago)
				{"DATA-9", "alice", "The ML training queries need COMPUTE_XL to finish within our nightly batch window.", 3},
				{"DATA-9", "bob", "How many credits per run are we expecting? Finance flagged our Snowflake spend last month.", 3},
				{"DATA-9", "alice", "Estimated 12-15 credits per run, 3 runs per week. I'll add cost monitoring alerts.", 2},

				// DATA-12: PII dataset access for fraud analysis (1 day ago)
				{"DATA-12", "bob", "I've completed the required security awareness training. Certificate ID: SAT-2024-4821.", 1},
				{"DATA-12", "admin", "Need to verify the DPA is countersigned by the data protection officer before we can approve this.", 0},

				// SEC-13: Vault admin access (18 days ago)
				{"SEC-13", "charlie", "Rotation policy is configured. All prod DB credentials now rotate every 30 days automatically.", 15},
				{"SEC-13", "admin", "Approved and verified. The rotation ran successfully in the last cycle.", 14},

				// SEC-14: SentinelOne console access (1 day ago)
				{"SEC-14", "charlie", "The alert on web-prod-03 shows suspicious outbound connections to an unknown IP. Need console access to isolate.", 1},
				{"SEC-14", "diana", "I've reviewed the network logs. Looks like a false positive from the new health check service, but we should verify.", 1},
				{"SEC-14", "admin", "Granting temporary access. Please document findings in the incident report template.", 0},

				// SEC-16: Pentest tools access (7 days ago)
				{"SEC-16", "diana", "I've prepared the scope document. Testing will be limited to the staging environment and pre-approved IP ranges.", 6},
				{"SEC-16", "charlie", "Burp Suite Pro license key received. Setting up the pentest VPN profile now.", 5},
				{"SEC-16", "admin", "Make sure the rules of engagement doc is signed by both teams before any active scanning begins.", 4},
				{"SEC-16", "charlie", "Signed ROE attached. We're ready to start the assessment on Monday.", 3},
			}

			type commentRef struct {
				commentID uuid.UUID
				taskID    uuid.UUID
				daysAgo   int
			}
			var commentRefs []commentRef

			for _, c := range comments {
				ref, ok := taskRefs[c.taskKey]
				if !ok {
					fmt.Printf("  Warning: task key %s not found, skipping comment\n", c.taskKey)
					continue
				}

				tok, err := signinAs(c.author)
				if err != nil {
					return err
				}
				client.token = tok

				resp, err := client.post(
					fmt.Sprintf("/api/v1/projects/%s/tasks/%d/comments", ref.projKey, ref.taskNumber),
					map[string]string{
						"content": c.content,
					},
				)
				if err != nil {
					return fmt.Errorf("failed to create comment on %s: %w", c.taskKey, err)
				}
				var createdComment struct {
					ID uuid.UUID `json:"id"`
				}
				if err := decodeJSON(resp, &createdComment); err != nil {
					return fmt.Errorf("failed to decode comment response on %s: %w", c.taskKey, err)
				}
				commentRefs = append(commentRefs, commentRef{
					commentID: createdComment.ID,
					taskID:    ref.taskID,
					daysAgo:   c.daysAgo,
				})
				time.Sleep(time.Millisecond)
			}

			// -----------------------------------------------------------------
			// 12. Backdate timestamps to simulate realistic timing.
			//
			// Tasks, comments, and activity logs all got created_at = NOW().
			// We shift them backward so the demo data looks like it accumulated
			// over days/weeks. Activity log checksums include created_at, so we
			// must recompute the entire chain for each task after shifting.
			// -----------------------------------------------------------------
			fmt.Println("Backdating timestamps...")

			now := time.Now().UTC()

			// Backdate tasks.
			for _, ref := range taskRefs {
				if ref.daysAgo == 0 {
					continue
				}
				offset := time.Duration(ref.daysAgo) * 24 * time.Hour
				_, err := pool.Exec(ctx,
					`UPDATE tasks SET created_at = created_at - $1::interval, updated_at = updated_at - $1::interval WHERE id = $2`,
					fmt.Sprintf("%d hours", int(offset.Hours())), ref.taskID,
				)
				if err != nil {
					return fmt.Errorf("failed to backdate task %s: %w", ref.taskID, err)
				}
			}

			// Backdate comments.
			for _, ref := range commentRefs {
				if ref.daysAgo == 0 {
					continue
				}
				offset := time.Duration(ref.daysAgo) * 24 * time.Hour
				_, err := pool.Exec(ctx,
					`UPDATE comments SET created_at = created_at - $1::interval, updated_at = updated_at - $1::interval WHERE id = $2`,
					fmt.Sprintf("%d hours", int(offset.Hours())), ref.commentID,
				)
				if err != nil {
					return fmt.Errorf("failed to backdate comment %s: %w", ref.commentID, err)
				}
			}

			// Backdate activity logs and recompute checksums.
			// Each task has its own independent checksum chain, so we process
			// tasks one at a time.
			taskIDSet := make(map[uuid.UUID]int) // taskID -> daysAgo
			for _, ref := range taskRefs {
				taskIDSet[ref.taskID] = ref.daysAgo
			}

			for taskID, daysAgo := range taskIDSet {
				if daysAgo == 0 {
					continue
				}

				offset := time.Duration(daysAgo) * 24 * time.Hour

				// Shift all activity timestamps for this task.
				_, err := pool.Exec(ctx,
					`UPDATE activity_log SET created_at = created_at - $1::interval WHERE task_id = $2`,
					fmt.Sprintf("%d hours", int(offset.Hours())), taskID,
				)
				if err != nil {
					return fmt.Errorf("failed to backdate activity for task %s: %w", taskID, err)
				}

				// Fetch shifted activity entries in order and recompute checksums.
				rows, err := pool.Query(ctx,
					`SELECT id, task_id, activity_type, actor_id, field_name, old_value, new_value, created_at
					 FROM activity_log WHERE task_id = $1 ORDER BY created_at ASC, id ASC`, taskID,
				)
				if err != nil {
					return fmt.Errorf("failed to fetch activity for recompute: %w", err)
				}

				type activityRow struct {
					id           uuid.UUID
					taskID       uuid.UUID
					activityType string
					actorID      uuid.UUID
					fieldName    pgtype.Text
					oldValue     []byte
					newValue     []byte
					createdAt    pgtype.Timestamptz
				}
				var activities []activityRow
				for rows.Next() {
					var a activityRow
					if err := rows.Scan(&a.id, &a.taskID, &a.activityType, &a.actorID,
						&a.fieldName, &a.oldValue, &a.newValue, &a.createdAt); err != nil {
						rows.Close()
						return fmt.Errorf("failed to scan activity row: %w", err)
					}
					activities = append(activities, a)
				}
				rows.Close()

				previousChecksum := "0000000000000000000000000000000000000000000000000000000000000000"
				for _, a := range activities {
					var fieldName *string
					if a.fieldName.Valid {
						fieldName = &a.fieldName.String
					}

					ts := a.createdAt.Time
					checksum := computeChecksum(previousChecksum, a.taskID, a.activityType, a.actorID, fieldName, a.oldValue, a.newValue, ts)

					_, err := pool.Exec(ctx,
						`UPDATE activity_log SET checksum = $1 WHERE id = $2`,
						checksum, a.id,
					)
					if err != nil {
						return fmt.Errorf("failed to update checksum: %w", err)
					}
					previousChecksum = checksum
				}
			}

			// Also backdate projects so they look older than their tasks.
			_, err = pool.Exec(ctx,
				`UPDATE projects SET created_at = $1, updated_at = $1`,
				now.Add(-45*24*time.Hour),
			)
			if err != nil {
				return fmt.Errorf("failed to backdate projects: %w", err)
			}

			fmt.Println("\nDemo data populated successfully!")
			fmt.Println("\nCredentials:")
			fmt.Println("  Admin:  admin@example.com / admin123")
			if demoEmail != "" && demoPassword != "" {
				fmt.Printf("  Demo:   %s / %s\n", demoEmail, demoPassword)
			}
			fmt.Println("  Users:  alice@example.com / user123")
			fmt.Println("          bob@example.com / user123")
			fmt.Println("          charlie@example.com / user123")
			fmt.Println("          diana@example.com / user123")

			return nil
		},
	}
}

// computeChecksum replicates the checksum calculation from
// internal/activity/activity.go so we can recompute after backdating.
func computeChecksum(
	previousChecksum string,
	taskID uuid.UUID,
	activityType string,
	actorID uuid.UUID,
	fieldName *string,
	oldValue []byte,
	newValue []byte,
	timestamp time.Time,
) string {
	fieldNameStr := ""
	if fieldName != nil {
		fieldNameStr = *fieldName
	}

	// Normalize JSON values the same way activity.go does during verification.
	oldValue = normalizeJSON(oldValue)
	newValue = normalizeJSON(newValue)

	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s",
		previousChecksum,
		taskID.String(),
		activityType,
		actorID.String(),
		fieldNameStr,
		string(oldValue),
		string(newValue),
		timestamp.Format(time.RFC3339Nano),
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// normalizeJSON unmarshals and re-marshals JSON to ensure consistent key ordering.
func normalizeJSON(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return data
	}
	normalized, err := json.Marshal(v)
	if err != nil {
		return data
	}
	return normalized
}
