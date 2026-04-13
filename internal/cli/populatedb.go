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
	"strconv"
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
				PasswordHash: pgtype.Text{String: adminHash, Valid: true},
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
				// DATA: all users
				{"DATA", "alice", "member"},
				{"DATA", "bob", "member"},
				{"DATA", "charlie", "member"},
				{"DATA", "diana", "member"},
				// SEC: all users
				{"SEC", "charlie", "member"},
				{"SEC", "diana", "member"},
				{"SEC", "alice", "member"},
				{"SEC", "bob", "member"},
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
				name  string
				title string
				desc  string
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
				{"INFRA", "Seeking Azure access for Disaster Recovery setup", "## Request\nNeed **Contributor** role on the Azure DR resource group to configure geo-replicated failover for the payments service.\n\n## Scope\n- **Resource Group:** `rg-payments-dr-westus2`\n- **Role:** Contributor\n- **Duration:** Permanent\n\n## Justification\nThe payments service SLA requires a documented DR failover procedure. This access is needed to:\n1. Configure Azure Site Recovery vaults\n2. Set up geo-replicated storage accounts\n3. Test failover runbooks\n\n## Compliance\n- [x] Manager approval obtained\n- [x] Security training up to date\n- [ ] DR runbook reviewed", "started", 0, "alice", []string{"alice"}, []string{"New Access", "Urgent"}, 3},

				{"INFRA", "Request AWS IAM role for production deployments", "## Request\nRequesting `AssumeRole` permissions on the **prod-deploy** IAM role to run CI/CD pipelines against the production EKS cluster.\n\n## Details\n- **AWS Account:** `123456789012` (production)\n- **IAM Role:** `prod-deploy`\n- **Permission:** `sts:AssumeRole`\n- **Source:** GitHub Actions OIDC provider\n\n## Pipeline Configuration\n```yaml\nrole-to-assume: arn:aws:iam::123456789012:role/prod-deploy\nrole-session-name: github-actions-deploy\n```\n\n## Checklist\n- [x] IAM role trust policy reviewed\n- [x] Least-privilege permissions verified\n- [ ] CloudTrail monitoring configured", "started", 1, "bob", []string{"bob"}, []string{"New Access"}, 5},

				{"INFRA", "Requesting GCP project owner access for billing migration", "## Request\nNeed temporary **Owner** access to migrate billing accounts between GCP projects.\n\n## Migration Plan\n| Step | Action | Risk |\n|------|--------|------|\n| 1 | Export current billing config | None |\n| 2 | Reassign billing account | **Medium** - brief service interruption possible |\n| 3 | Verify resource attribution | None |\n| 4 | Remove Owner access | None |\n\n## Access Window\n- **Start:** Upon approval\n- **End:** 48 hours after approval (auto-revoke)\n\n## Justification\n`Billing Account Administrator` role does **not** permit project-level billing reassignment. Owner is the minimum viable role for this operation.", "unstarted", 2, "charlie", []string{"charlie"}, []string{"Escalation"}, 1},

				{"INFRA", "Renew Kubernetes cluster admin certificate", "## Request\nCurrent cluster admin certificate expires in **14 days**. Requesting renewal for another 12 months.\n\n## Affected Clusters\n| Cluster | Region | Expiry |\n|---------|--------|--------|\n| `prod-eks-01` | us-east-1 | 14 days |\n| `prod-eks-02` | eu-west-1 | 14 days |\n\n## Renewal Process\n1. Generate new CSR with existing CN\n2. Sign with cluster CA\n3. Distribute to CI/CD secrets\n4. Rotate kubeconfig in Vault\n5. Verify cluster access\n\n## Impact\n> **Warning:** If not renewed, all automated deployments will fail after expiry.", "backlog", 3, "diana", []string{}, []string{"Renewal"}, 12},

				{"INFRA", "Seeking VPN access to staging environment", "## Request\nNeed site-to-site VPN credentials to access staging servers for integration testing of the new checkout flow.\n\n## Details\n- **VPN Gateway:** `vpn-staging.internal.example.com`\n- **Protocol:** WireGuard\n- **Subnet:** `10.20.0.0/16` (staging)\n\n## Testing Scope\n- Checkout API integration tests\n- Payment gateway sandbox connectivity\n- Order processing pipeline verification\n\n## Checklist\n- [x] VPN client installed\n- [x] 2FA device registered\n- [ ] VPN profile configured", "unstarted", 2, "alice", []string{"alice"}, []string{"New Access"}, 2},

				{"INFRA", "Revoke former contractor SSH keys from jump hosts", "## Request\nContractor engagement ended last Friday. All SSH keys for user `j.smith` need to be removed from bastion and jump hosts.\n\n## Affected Hosts\n- `bastion-prod-01` — 2 keys\n- `jump-staging-01` — 1 key\n- `jump-prod-01` — 1 key\n\n## Key Fingerprints\n```\nSHA256:xK3j9... (j.smith@contractor-laptop)\nSHA256:mP8w2... (j.smith@contractor-desktop)\n```\n\n## Actions Required\n1. Remove SSH public keys from `~/.ssh/authorized_keys`\n2. Rotate host keys on affected servers\n3. Update Ansible inventory\n4. Verify no residual access", "unstarted", 0, "admin", []string{"bob"}, []string{"Revocation", "Urgent"}, 1},

				{"INFRA", "Request Terraform Cloud workspace access", "## Request\nNeed **write** access to the `infra-prod` workspace in Terraform Cloud to apply changes for the database scaling project.\n\n## Workspace Details\n- **Organization:** `zerodha-infra`\n- **Workspace:** `infra-prod`\n- **VCS Repo:** `github.com/zerodha/infra-prod`\n- **Permission Level:** Write (plan + apply)\n\n## Changes Planned\n- Scale RDS instances from `db.r6g.xlarge` to `db.r6g.2xlarge`\n- Add read replicas in `ap-south-1`\n- Update security group rules for new replica endpoints\n\n## Compliance\n- [x] Change request filed (CR-2024-892)\n- [x] Peer review completed\n- [x] Rollback plan documented", "completed", 1, "bob", []string{"bob"}, []string{"New Access"}, 14},

				{"INFRA", "Requesting root access to legacy on-prem servers", "## Request\nNeed temporary **root** access to legacy on-prem servers for the data center decommission.\n\n## Servers\n| Hostname | OS | Purpose |\n|----------|----|---------|\n| `legacy-db-01` | RHEL 7 | MySQL 5.7 |\n| `legacy-app-01` | CentOS 7 | Java 8 app server |\n| `legacy-app-02` | CentOS 7 | Java 8 app server |\n\n## Justification\nBoot configuration modifications require root. Specifically:\n- Modify `/etc/fstab` for storage detach\n- Update GRUB config for decommission mode\n- Extract final data exports\n\n## Time Box\n- **Strictly limited to 2 weeks** from approval date\n- Auto-revocation configured via PAM", "cancelled", 0, "charlie", []string{"charlie"}, []string{"Escalation"}, 10},

				// DATA project
				{"DATA", "Seeking read access to production Postgres replica", "## Request\nNeed `SELECT` permissions on the analytics replica to run quarterly revenue reports.\n\n## Database Details\n- **Host:** `analytics-replica.db.internal`\n- **Port:** `5432`\n- **Database:** `production`\n- **Schemas:** `public`, `billing`, `analytics`\n\n## Access Level\n- **Read-only** — `SELECT` on specified schemas\n- No write, no DDL, no function execution\n\n## Queries Planned\n```sql\nSELECT date_trunc('month', created_at) AS month,\n       SUM(amount) AS revenue\nFROM billing.transactions\nWHERE status = 'completed'\nGROUP BY 1 ORDER BY 1;\n```\n\n## Compliance\n- [x] Data access training completed\n- [x] Read replica endpoint confirmed (not primary)", "completed", 1, "admin", []string{"admin", "alice"}, []string{"New Access"}, 21},

				{"DATA", "Request Snowflake warehouse access for ML training", "## Request\nRequesting access to the `COMPUTE_XL` warehouse to run feature engineering queries for the churn prediction model.\n\n## Warehouse Details\n- **Warehouse:** `COMPUTE_XL`\n- **Size:** X-Large (16 credits/hr)\n- **Auto-suspend:** 5 minutes\n- **Database:** `ML_FEATURES`\n\n## Estimated Usage\n| Metric | Value |\n|--------|-------|\n| Credits per run | 12-15 |\n| Runs per week | 3 |\n| Monthly cost | ~$1,800 |\n\n## ML Pipeline\n1. Extract user behavioral features\n2. Join with transaction history\n3. Compute rolling aggregates\n4. Export to S3 for model training\n\n## Cost Controls\n- [x] Resource monitor configured at 50 credits/day\n- [x] Alert at 80% threshold\n- [ ] Monthly review with finance", "started", 0, "alice", []string{"alice"}, []string{"New Access"}, 4},

				{"DATA", "Seeking Redshift access for customer segmentation", "## Request\nNeed read access to the `customer_events` and `transactions` schemas in Redshift for building segmentation dashboards.\n\n## Access Details\n- **Cluster:** `analytics-prod.redshift.amazonaws.com`\n- **Database:** `warehouse`\n- **Schemas:** `customer_events`, `transactions`\n- **Permission:** `SELECT` only\n\n## Dashboard Requirements\n- Customer cohort analysis by signup month\n- Transaction frequency distributions\n- Churn risk segmentation\n- Geographic revenue breakdown\n\n## Tools\n- Visualization: **Metabase** (existing license)\n- ETL: dbt models in `analytics-dbt` repo", "unstarted", 2, "bob", []string{"bob"}, []string{"New Access"}, 2},

				{"DATA", "Renew Looker admin license before Q4", "## Request\nCurrent Looker admin license expires **end of September**. Requesting renewal and addition of 5 viewer seats.\n\n## License Details\n| Type | Current | Requested |\n|------|---------|----------|\n| Admin | 2 | 2 (renew) |\n| Developer | 5 | 5 (renew) |\n| Viewer | 10 | **15** (+5) |\n\n## Business Case\nThe analytics team is onboarding 5 new business analysts in Q4 who need dashboard access.\n\n## Budget\n- Renewal cost: ~$24,000/year\n- Additional seats: ~$6,000/year\n- **Total:** ~$30,000\n- **Budget code:** ANALYTICS-2024-Q4\n\n## Timeline\n> License must be renewed by **September 30** to avoid service interruption.", "backlog", 3, "admin", []string{}, []string{"Renewal"}, 30},

				{"DATA", "Request access to PII dataset for fraud analysis", "## Request\nNeed access to unmasked PII fields in the `fraud_detection` schema for active fraud investigation.\n\n## Data Fields Required\n| Field | Table | Sensitivity |\n|-------|-------|-------------|\n| `email` | `users` | PII |\n| `phone` | `users` | PII |\n| `ip_address` | `sessions` | PII |\n| `billing_address` | `transactions` | PII |\n\n## Justification\nActive fraud ring investigation — masked data insufficient for cross-referencing identities across accounts.\n\n## Compliance\n- [x] Security awareness training completed (cert: `SAT-2024-4821`)\n- [x] DPA signed and filed\n- [ ] DPO countersignature pending\n- [ ] Access auto-expires after 30 days", "unstarted", 0, "bob", []string{"bob"}, []string{"New Access", "Urgent"}, 1},

				// SEC project
				{"SEC", "Seeking Vault admin access for secrets rotation", "## Request\nNeed **admin** access to HashiCorp Vault to implement automated secrets rotation for production database credentials.\n\n## Vault Details\n- **Cluster:** `vault.internal.example.com`\n- **Namespace:** `production/databases`\n- **Auth Method:** OIDC (Okta)\n\n## Rotation Policy\n```hcl\npath \"database/rotate-root/*\" {\n  capabilities = [\"update\"]\n}\npath \"database/creds/*\" {\n  capabilities = [\"read\"]\n}\n```\n\n## Implementation Plan\n1. Configure dynamic database secrets engine\n2. Set rotation period to 30 days\n3. Update application configs to use dynamic credentials\n4. Test failover with rotated credentials\n\n## Rollback\n- Static credentials preserved in break-glass envelope\n- Manual rotation procedure documented in runbook", "completed", 1, "admin", []string{"admin", "charlie"}, []string{"New Access"}, 18},

				{"SEC", "Request SentinelOne console access for incident response", "## Request\nIR team member needs access to the SentinelOne EDR console to investigate alert on `web-prod-03`.\n\n## Alert Details\n- **Alert ID:** `S1-2024-78432`\n- **Severity:** High\n- **Host:** `web-prod-03`\n- **Detection:** Suspicious outbound connections to `198.51.100.42`\n- **Timestamp:** 2024-01-15T03:42:00Z\n\n## Investigation Scope\n1. Review process tree for anomalous activity\n2. Check network connections timeline\n3. Analyze file modifications in last 24h\n4. Correlate with other host alerts\n\n## Urgency\n> **High Priority** — Potential active compromise requires immediate investigation.\n\n## Access Duration\n- Temporary (72-hour window)\n- Auto-revokes after investigation closes", "started", 0, "charlie", []string{"charlie"}, []string{"New Access", "Urgent"}, 1},

				{"SEC", "Seeking Cloudflare WAF access for rule updates", "## Request\nNeed **editor** access to Cloudflare WAF to deploy updated OWASP rules and add custom rate limiting.\n\n## Changes Planned\n### OWASP Rule Updates\n- Update to OWASP CRS v4.0\n- Enable paranoia level 2 for `/api/v1/auth/*` endpoints\n\n### Custom Rate Limiting\n| Endpoint | Limit | Window |\n|----------|-------|--------|\n| `/api/v1/auth/login` | 10 req | 1 min |\n| `/api/v1/auth/signup` | 5 req | 5 min |\n| `/api/v1/auth/reset` | 3 req | 15 min |\n\n## Testing\n- [ ] Rules tested in staging zone first\n- [ ] False positive analysis completed\n- [ ] Rollback procedure documented\n\n## Impact\nExpected to block ~15% of current bot traffic based on staging analysis.", "unstarted", 2, "diana", []string{"diana"}, []string{"New Access"}, 3},

				{"SEC", "Request access to penetration testing tools", "## Request\nRequesting access to **Burp Suite Pro** license and internal pentest VPN for the scheduled Q4 security assessment.\n\n## Assessment Scope\n- **Target:** Production web application (`app.example.com`)\n- **Type:** Grey box\n- **Environment:** Staging (with production-like data)\n- **Duration:** 2 weeks\n\n## Tools Required\n| Tool | License | Status |\n|------|---------|--------|\n| Burp Suite Pro | Annual | Pending |\n| Pentest VPN | Internal | Pending |\n| Nessus Pro | Existing | Active |\n\n## Prerequisites\n- [x] Scope document drafted\n- [x] Rules of engagement prepared\n- [ ] ROE signed by both teams\n- [ ] Notification sent to affected teams\n\n## Deliverables\n- Vulnerability assessment report\n- Risk rating matrix\n- Remediation recommendations", "started", 1, "charlie", []string{"charlie", "diana"}, []string{"New Access"}, 7},

				{"SEC", "Revoke SOC analyst access after role transfer", "## Request\nSOC analyst transferred to engineering. Remove access to all security monitoring tools.\n\n## Access to Revoke\n| Tool | Access Level | Account |\n|------|--------------|---------|\n| Splunk | SOC Dashboard | `analyst.jones` |\n| CrowdStrike | Responder | `ajones@cs` |\n| PagerDuty | SOC On-Call | Team: `soc-tier1` |\n| Jira | Security Board | Project: `SEC` |\n| Slack | `#soc-alerts` | Member |\n\n## Timeline\n- **Transfer effective:** Last Monday\n- **Access removal deadline:** End of this week\n\n## Process\n1. Disable accounts in each tool\n2. Remove from SOC distribution lists\n3. Transfer ownership of active investigations\n4. Update on-call rotation\n5. Confirm with HR and new manager", "backlog", 2, "admin", []string{}, []string{"Revocation"}, 9},
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
				"backlog":   {},                       // already there (default)
				"unstarted": {"unstarted"},            // backlog -> unstarted
				"started":   {"unstarted", "started"}, // backlog -> unstarted -> started
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
			// 11. Mid-lifecycle mutations: add/remove assignees, add/remove
			//     labels, and update task fields after creation. Each of these
			//     generates additional activity log entries.
			// -----------------------------------------------------------------
			fmt.Println("Simulating mid-lifecycle changes...")

			type mutationSpec struct {
				taskKey  string
				actor    string
				mutation string // "add_assignee", "remove_assignee", "add_label", "remove_label", "update_priority", "update_title", "update_description"
				target   string // username for assignees, label name for labels, or new value for updates
			}

			mutations := []mutationSpec{
				// INFRA-0: Azure DR — admin reviews, adds reviewers, relabels
				{"INFRA-0", "admin", "add_assignee", "bob"},
				{"INFRA-0", "admin", "add_assignee", "diana"},
				{"INFRA-0", "admin", "add_label", "Escalation"},
				{"INFRA-0", "admin", "update_title", "Azure Contributor access for Disaster Recovery failover setup"},
				{"INFRA-0", "admin", "remove_label", "Urgent"},
				{"INFRA-0", "admin", "add_label", "Renewal"},

				// INFRA-1: AWS IAM — alice joins, priority and title updated
				{"INFRA-1", "admin", "add_assignee", "alice"},
				{"INFRA-1", "admin", "add_assignee", "charlie"},
				{"INFRA-1", "admin", "add_label", "Urgent"},
				{"INFRA-1", "bob", "update_title", "Request AWS IAM role for production EKS deployments"},
				{"INFRA-1", "admin", "add_label", "Escalation"},
				{"INFRA-1", "admin", "update_priority", "0"},

				// INFRA-2: GCP owner — multiple reviewers, label changes
				{"INFRA-2", "admin", "add_assignee", "diana"},
				{"INFRA-2", "admin", "add_assignee", "admin"},
				{"INFRA-2", "admin", "add_assignee", "bob"},
				{"INFRA-2", "admin", "add_label", "Urgent"},
				{"INFRA-2", "admin", "add_label", "Renewal"},
				{"INFRA-2", "admin", "remove_assignee", "bob"},

				// INFRA-3: K8s cert renewal — admin picks it up, team assembled
				{"INFRA-3", "admin", "add_assignee", "admin"},
				{"INFRA-3", "admin", "add_assignee", "alice"},
				{"INFRA-3", "admin", "add_assignee", "bob"},
				{"INFRA-3", "admin", "add_label", "Urgent"},
				{"INFRA-3", "admin", "add_label", "Escalation"},
				{"INFRA-3", "admin", "update_priority", "0"},

				// INFRA-4: VPN access — bob added, labels updated
				{"INFRA-4", "admin", "add_assignee", "bob"},
				{"INFRA-4", "admin", "add_assignee", "charlie"},
				{"INFRA-4", "admin", "add_label", "Renewal"},
				{"INFRA-4", "admin", "add_label", "Escalation"},
				{"INFRA-4", "admin", "update_priority", "1"},

				// INFRA-5: Revoke SSH keys — team assembled, labels managed
				{"INFRA-5", "admin", "add_assignee", "alice"},
				{"INFRA-5", "admin", "add_assignee", "charlie"},
				{"INFRA-5", "admin", "add_assignee", "diana"},
				{"INFRA-5", "admin", "add_label", "Escalation"},
				{"INFRA-5", "admin", "add_label", "New Access"},
				{"INFRA-5", "admin", "remove_label", "New Access"},

				// INFRA-6: Terraform workspace — completed, reviewers added, labels cycled
				{"INFRA-6", "admin", "add_assignee", "admin"},
				{"INFRA-6", "admin", "add_assignee", "alice"},
				{"INFRA-6", "admin", "add_assignee", "charlie"},
				{"INFRA-6", "admin", "remove_label", "New Access"},
				{"INFRA-6", "admin", "add_label", "Renewal"},
				{"INFRA-6", "admin", "update_priority", "0"},

				// INFRA-7: Root access — escalated then cancelled
				{"INFRA-7", "admin", "add_assignee", "admin"},
				{"INFRA-7", "admin", "add_assignee", "diana"},
				{"INFRA-7", "admin", "add_label", "Urgent"},
				{"INFRA-7", "admin", "remove_label", "Urgent"},
				{"INFRA-7", "admin", "add_label", "Revocation"},
				{"INFRA-7", "admin", "update_priority", "1"},

				// DATA-8: Postgres replica — team assembled, priority updated
				{"DATA-8", "admin", "add_assignee", "bob"},
				{"DATA-8", "admin", "add_assignee", "charlie"},
				{"DATA-8", "admin", "add_label", "Renewal"},
				{"DATA-8", "admin", "update_priority", "0"},
				{"DATA-8", "admin", "add_label", "Escalation"},

				// DATA-9: Snowflake — cost review team, labels updated
				{"DATA-9", "admin", "add_assignee", "bob"},
				{"DATA-9", "admin", "add_assignee", "admin"},
				{"DATA-9", "alice", "add_label", "Urgent"},
				{"DATA-9", "alice", "update_title", "Request Snowflake COMPUTE_XL warehouse for ML feature engineering"},
				{"DATA-9", "admin", "add_label", "Escalation"},
				{"DATA-9", "admin", "add_assignee", "charlie"},

				// DATA-10: Redshift — admin and alice added, labels updated
				{"DATA-10", "admin", "add_assignee", "admin"},
				{"DATA-10", "admin", "add_assignee", "alice"},
				{"DATA-10", "admin", "add_label", "Escalation"},
				{"DATA-10", "admin", "add_label", "Urgent"},
				{"DATA-10", "admin", "update_priority", "1"},

				// DATA-11: Looker renewal — team assembled, urgency added
				{"DATA-11", "admin", "add_assignee", "alice"},
				{"DATA-11", "admin", "add_assignee", "bob"},
				{"DATA-11", "admin", "add_assignee", "diana"},
				{"DATA-11", "admin", "add_label", "Escalation"},
				{"DATA-11", "admin", "add_label", "Urgent"},
				{"DATA-11", "admin", "update_priority", "1"},

				// DATA-12: PII dataset — compliance reviewers, labels added
				{"DATA-12", "admin", "add_assignee", "admin"},
				{"DATA-12", "admin", "add_assignee", "alice"},
				{"DATA-12", "admin", "add_assignee", "charlie"},
				{"DATA-12", "admin", "add_label", "Escalation"},
				{"DATA-12", "admin", "add_label", "Renewal"},

				// SEC-13: Vault admin — verification team, priority updated
				{"SEC-13", "admin", "add_assignee", "diana"},
				{"SEC-13", "admin", "add_assignee", "bob"},
				{"SEC-13", "admin", "add_label", "Escalation"},
				{"SEC-13", "admin", "update_priority", "0"},
				{"SEC-13", "admin", "add_label", "Renewal"},

				// SEC-14: SentinelOne — IR team assembled, labels managed
				{"SEC-14", "admin", "add_assignee", "admin"},
				{"SEC-14", "admin", "add_assignee", "diana"},
				{"SEC-14", "admin", "add_assignee", "bob"},
				{"SEC-14", "admin", "add_label", "Escalation"},
				{"SEC-14", "admin", "add_label", "Revocation"},
				{"SEC-14", "admin", "remove_label", "Revocation"},

				// SEC-15: Cloudflare WAF — team expanded, labels updated
				{"SEC-15", "admin", "add_assignee", "charlie"},
				{"SEC-15", "admin", "add_assignee", "admin"},
				{"SEC-15", "admin", "add_assignee", "bob"},
				{"SEC-15", "admin", "add_label", "Urgent"},
				{"SEC-15", "diana", "add_label", "Escalation"},
				{"SEC-15", "admin", "update_priority", "1"},

				// SEC-16: Pentest tools — admin added, labels cycled
				{"SEC-16", "admin", "add_assignee", "admin"},
				{"SEC-16", "admin", "add_assignee", "bob"},
				{"SEC-16", "admin", "add_label", "Escalation"},
				{"SEC-16", "admin", "add_label", "Urgent"},
				{"SEC-16", "admin", "remove_label", "Urgent"},
				{"SEC-16", "admin", "update_priority", "0"},

				// SEC-17: SOC analyst revocation — team assigned, labels added
				{"SEC-17", "admin", "add_assignee", "charlie"},
				{"SEC-17", "admin", "add_assignee", "diana"},
				{"SEC-17", "admin", "add_assignee", "bob"},
				{"SEC-17", "admin", "add_label", "Urgent"},
				{"SEC-17", "admin", "add_label", "Escalation"},
				{"SEC-17", "admin", "add_label", "New Access"},
				{"SEC-17", "admin", "remove_label", "New Access"},
			}

			for _, m := range mutations {
				ref, ok := taskRefs[m.taskKey]
				if !ok {
					fmt.Printf("  Warning: task key %s not found, skipping mutation\n", m.taskKey)
					continue
				}

				tok, err := signinAs(m.actor)
				if err != nil {
					return err
				}
				client.token = tok

				switch m.mutation {
				case "add_assignee":
					resp, err := client.post(
						fmt.Sprintf("/api/v1/projects/%s/tasks/%d/assignees", ref.projKey, ref.taskNumber),
						map[string]string{"user_id": userIDs[m.target].String()},
					)
					if err != nil {
						return fmt.Errorf("failed to add assignee %s to %s: %w", m.target, m.taskKey, err)
					}
					if err := discardBody(resp); err != nil {
						return fmt.Errorf("failed to add assignee %s to %s: %w", m.target, m.taskKey, err)
					}

				case "remove_assignee":
					resp, err := client.do("DELETE",
						fmt.Sprintf("/api/v1/projects/%s/tasks/%d/assignees/%s", ref.projKey, ref.taskNumber, userIDs[m.target].String()),
						nil,
					)
					if err != nil {
						return fmt.Errorf("failed to remove assignee %s from %s: %w", m.target, m.taskKey, err)
					}
					if err := discardBody(resp); err != nil {
						return fmt.Errorf("failed to remove assignee %s from %s: %w", m.target, m.taskKey, err)
					}

				case "add_label":
					resp, err := client.post(
						fmt.Sprintf("/api/v1/projects/%s/tasks/%d/labels", ref.projKey, ref.taskNumber),
						map[string]string{"label_id": labelMap[ref.projKey][m.target].ID.String()},
					)
					if err != nil {
						return fmt.Errorf("failed to add label %s to %s: %w", m.target, m.taskKey, err)
					}
					if err := discardBody(resp); err != nil {
						return fmt.Errorf("failed to add label %s to %s: %w", m.target, m.taskKey, err)
					}

				case "remove_label":
					resp, err := client.do("DELETE",
						fmt.Sprintf("/api/v1/projects/%s/tasks/%d/labels/%s", ref.projKey, ref.taskNumber, labelMap[ref.projKey][m.target].ID.String()),
						nil,
					)
					if err != nil {
						return fmt.Errorf("failed to remove label %s from %s: %w", m.target, m.taskKey, err)
					}
					if err := discardBody(resp); err != nil {
						return fmt.Errorf("failed to remove label %s from %s: %w", m.target, m.taskKey, err)
					}

				case "update_priority":
					prio, _ := strconv.Atoi(m.target)
					resp, err := client.patch(
						fmt.Sprintf("/api/v1/projects/%s/tasks/%d", ref.projKey, ref.taskNumber),
						map[string]interface{}{"priority": prio},
					)
					if err != nil {
						return fmt.Errorf("failed to update priority on %s: %w", m.taskKey, err)
					}
					if err := discardBody(resp); err != nil {
						return fmt.Errorf("failed to update priority on %s: %w", m.taskKey, err)
					}

				case "update_title":
					resp, err := client.patch(
						fmt.Sprintf("/api/v1/projects/%s/tasks/%d", ref.projKey, ref.taskNumber),
						map[string]interface{}{"title": m.target},
					)
					if err != nil {
						return fmt.Errorf("failed to update title on %s: %w", m.taskKey, err)
					}
					if err := discardBody(resp); err != nil {
						return fmt.Errorf("failed to update title on %s: %w", m.taskKey, err)
					}
				}

				time.Sleep(time.Millisecond)
			}

			// -----------------------------------------------------------------
			// 12. Create comments (handler logs comment_created activity).
			// -----------------------------------------------------------------
			fmt.Println("Creating comments...")

			type commentSpec struct {
				taskKey string
				author  string
				content string
				daysAgo int // how many days ago this comment was posted
			}

			comments := []commentSpec{
				// INFRA-0: Azure access for DR setup (3 days ago) — 9 comments
				{"INFRA-0", "admin", "Which resource group specifically? We have three DR groups across regions.", 2},
				{"INFRA-0", "alice", "The `rg-payments-dr-westus2` group. I've attached the architecture diagram for context.", 2},
				{"INFRA-0", "bob", "I can vouch for this — the failover testing is blocking our SLA compliance deadline.", 2},
				{"INFRA-0", "admin", "Understood. I'll fast-track this given the SLA implications. Bob, can you verify the resource group exists?", 2},
				{"INFRA-0", "bob", "Confirmed — `rg-payments-dr-westus2` exists and currently has no Contributor assignments outside infra-admins.", 1},
				{"INFRA-0", "alice", "I've also added the DR runbook link to the ticket description for reference.", 1},
				{"INFRA-0", "diana", "Heads up — I ran into RBAC propagation delays last time. Allow 15-20 min after assignment before testing.", 1},
				{"INFRA-0", "admin", "Good point Diana. Alice, once access is granted, please confirm you can list resources in the group.", 1},
				{"INFRA-0", "alice", "Will do. I'll run `az resource list --resource-group rg-payments-dr-westus2` to verify.", 0},

				// INFRA-1: AWS IAM role (5 days ago) — 8 comments
				{"INFRA-1", "admin", "Which AWS account is this for? We have separate accounts for staging and production.", 4},
				{"INFRA-1", "bob", "Production account `123456789012`. The IAM role `prod-deploy` already exists — I just need `sts:AssumeRole` on it.", 4},
				{"INFRA-1", "alice", "I reviewed the trust policy on that role. It currently only allows our Jenkins instance. We'll need to add the GitHub OIDC provider.", 3},
				{"INFRA-1", "bob", "Correct. Here's the trust policy addition needed:\n```json\n{\n  \"Effect\": \"Allow\",\n  \"Principal\": {\"Federated\": \"arn:aws:iam::123456789012:oidc-provider/token.actions.githubusercontent.com\"},\n  \"Action\": \"sts:AssumeRole\"\n}\n```", 3},
				{"INFRA-1", "admin", "Looks good. CloudTrail will log all AssumeRole events. Make sure the session name includes the workflow run ID for traceability.", 3},
				{"INFRA-1", "alice", "I've also set up an SNS alert for any unusual AssumeRole patterns on this role.", 2},
				{"INFRA-1", "bob", "Pipeline tested successfully in a dry-run. The EKS deployment completes in ~4 minutes with the new role.", 1},
				{"INFRA-1", "admin", "Approved. Keep the session duration to 15 minutes max — that should be sufficient for deployments.", 1},

				// INFRA-2: GCP project owner access (1 day ago) — 8 comments
				{"INFRA-2", "admin", "Owner access is a high privilege. Can we scope this down to Billing Account Administrator instead?", 1},
				{"INFRA-2", "charlie", "Billing Account Administrator won't let me reassign the linked project. I need Owner for the migration step only.", 1},
				{"INFRA-2", "diana", "Suggest we grant time-boxed access with a calendar reminder for revocation.", 1},
				{"INFRA-2", "admin", "Agreed on time-boxing. Charlie, how long do you need? Can we do this within a single business day?", 1},
				{"INFRA-2", "charlie", "Yes — the actual migration takes about 2 hours. I'll start first thing Monday morning and relinquish by noon.", 1},
				{"INFRA-2", "diana", "I'll set up a Cloud Function to auto-revoke the Owner role after 8 hours as a safety net.", 0},
				{"INFRA-2", "admin", "Excellent. I'll grant access Monday at 8 AM with the auto-revoke in place. Charlie, document each step as you go.", 0},
				{"INFRA-2", "charlie", "Will do. I'll post updates here as I complete each migration step.", 0},

				// INFRA-3: K8s cert renewal (12 days ago) — 10 comments
				{"INFRA-3", "admin", "Diana, what's the current expiry date? We should plan this well before the deadline.", 11},
				{"INFRA-3", "diana", "Both clusters expire on the 28th. I recommend we renew this week to give us buffer for any issues.", 11},
				{"INFRA-3", "alice", "I can help verify cluster access after the renewal. Do we need to update any CI/CD kubeconfigs too?", 10},
				{"INFRA-3", "diana", "Yes, the kubeconfig in Vault needs updating. The CI/CD pipeline pulls it dynamically, so it should pick up the new cert automatically.", 10},
				{"INFRA-3", "bob", "I checked the monitoring — we have Prometheus alerts set for cert expiry at 7 days and 3 days. Should trigger soon.", 10},
				{"INFRA-3", "admin", "Let's coordinate this for Wednesday. Alice, can you be available to verify access on both clusters?", 9},
				{"INFRA-3", "alice", "Wednesday works. I'll run the full connectivity test suite after the renewal.", 9},
				{"INFRA-3", "diana", "Also flagging: the `prod-eks-02` cluster in `eu-west-1` has a different CA. We'll need two separate CSRs.", 9},
				{"INFRA-3", "bob", "I can prepare the CSR for `eu-west-1` while Diana handles `us-east-1`. Parallel work should speed this up.", 8},
				{"INFRA-3", "admin", "Good plan. Bob takes EU, Diana takes US. Alice verifies both. Let's aim for completion by Thursday EOD.", 8},

				// INFRA-4: VPN access to staging (2 days ago) — 10 comments
				{"INFRA-4", "admin", "Which staging environment? We have `staging-v1` and `staging-v2`.", 2},
				{"INFRA-4", "alice", "The `staging-v2` environment. I need to test the new checkout flow against the payment gateway sandbox.", 2},
				{"INFRA-4", "bob", "I can confirm the WireGuard config for staging-v2. The gateway is `vpn-staging-v2.internal.example.com`.", 1},
				{"INFRA-4", "alice", "Thanks Bob. I've updated the ticket description with the correct gateway address.", 1},
				{"INFRA-4", "charlie", "Quick note — staging-v2 was recently moved to a new subnet `10.20.1.0/24`. Make sure the VPN config reflects this.", 1},
				{"INFRA-4", "admin", "VPN profile generated with the updated subnet. Alice, you should have received the config via encrypted email.", 1},
				{"INFRA-4", "alice", "Received and installed. Testing connectivity now.", 1},
				{"INFRA-4", "bob", "Verified — I can see Alice's connection in the VPN dashboard. All looks good.", 0},
				{"INFRA-4", "alice", "Connectivity confirmed. I can reach the payment gateway sandbox on `10.20.1.50:443`.", 0},
				{"INFRA-4", "admin", "Great. The VPN access is set to expire in 30 days. Submit a renewal if needed beyond that.", 0},

				// INFRA-5: Revoke contractor SSH keys (1 day ago) — 8 comments
				{"INFRA-5", "bob", "I've identified 4 SSH keys across 3 jump hosts. Will need a maintenance window to rotate host keys too.", 1},
				{"INFRA-5", "admin", "Good call. Schedule the rotation for this weekend's maintenance window.", 1},
				{"INFRA-5", "alice", "Should we also audit the `authorized_keys` files for any other inactive users while we're at it?", 1},
				{"INFRA-5", "admin", "Yes, good idea. Let's do a full audit. Alice, can you pull the current list of all SSH key fingerprints across jump hosts?", 1},
				{"INFRA-5", "alice", "Running the audit now. I'll cross-reference fingerprints against our active contractor list.", 0},
				{"INFRA-5", "charlie", "FYI — I found one of j.smith's keys was also on the monitoring server `mon-prod-01`. Adding it to the removal list.", 0},
				{"INFRA-5", "bob", "Updated the removal list. Total is now 5 keys across 4 hosts. Ansible playbook is ready for the maintenance window.", 0},
				{"INFRA-5", "admin", "Perfect. Bob, please run the playbook during Saturday's maintenance window and confirm removal in this thread.", 0},

				// INFRA-6: Terraform Cloud workspace access (14 days ago) — 8 comments
				{"INFRA-6", "bob", "Access granted. Confirmed I can plan and apply in the `infra-prod` workspace.", 12},
				{"INFRA-6", "admin", "Approved. Please follow the change management process for any production applies.", 12},
				{"INFRA-6", "alice", "Bob, remember to tag your Terraform plans with the CR number for audit tracking.", 11},
				{"INFRA-6", "bob", "Good reminder. All plans will include the tag `cr:CR-2024-892`.", 11},
				{"INFRA-6", "admin", "First apply completed successfully — RDS instances scaled to `db.r6g.2xlarge` without downtime.", 8},
				{"INFRA-6", "bob", "Read replicas in `ap-south-1` are now provisioned and syncing. Replication lag is under 50ms.", 7},
				{"INFRA-6", "alice", "Verified the security group rules for the new replica endpoints look correct.", 6},
				{"INFRA-6", "admin", "All changes verified. Marking this as complete. Great work team.", 5},

				// INFRA-7: Root access to legacy on-prem (10 days ago) — 8 comments
				{"INFRA-7", "charlie", "The decommission requires modifying boot configs that only root can touch.", 9},
				{"INFRA-7", "admin", "Rejected — root access on production on-prem violates our compliance policy. Let's use the managed automation runbooks instead.", 9},
				{"INFRA-7", "charlie", "The automation runbooks don't cover `/etc/fstab` modifications. Can we get an exception?", 8},
				{"INFRA-7", "admin", "I've escalated to the compliance team. They'll review the exception request within 48 hours.", 8},
				{"INFRA-7", "diana", "As an alternative, we could use a privileged Ansible playbook that runs as root but is audited and version-controlled.", 7},
				{"INFRA-7", "charlie", "That could work. Diana, can you help set up the playbook?", 7},
				{"INFRA-7", "admin", "Compliance team denied the exception. The Ansible playbook approach is the approved path forward. Cancelling this request.", 6},
				{"INFRA-7", "charlie", "Understood. Diana and I will proceed with the Ansible approach. Opening a new ticket for that.", 6},

				// DATA-8: Read access to production Postgres replica (21 days ago) — 8 comments
				{"DATA-8", "alice", "Access configured. Connected successfully and ran a test query against the replica.", 19},
				{"DATA-8", "admin", "Approved. Remember to use the read replica endpoint, not the primary.", 19},
				{"DATA-8", "bob", "I've set up a monitoring alert for any queries exceeding 30 seconds on the replica.", 18},
				{"DATA-8", "alice", "Good thinking. The quarterly revenue query takes about 12 seconds, well within limits.", 18},
				{"DATA-8", "admin", "Alice, please share the query execution plan so we can optimize if needed.", 17},
				{"DATA-8", "alice", "Execution plan shared in the analytics Slack channel. The query uses index scans on `billing.transactions`.", 16},
				{"DATA-8", "bob", "Looks efficient. The `idx_transactions_status_created_at` index is being used as expected.", 15},
				{"DATA-8", "admin", "All good. Closing this out. Q3 revenue reports are on track.", 14},

				// DATA-9: Snowflake warehouse access (4 days ago) — 8 comments
				{"DATA-9", "alice", "The ML training queries need `COMPUTE_XL` to finish within our nightly batch window.", 3},
				{"DATA-9", "bob", "How many credits per run are we expecting? Finance flagged our Snowflake spend last month.", 3},
				{"DATA-9", "alice", "Estimated 12-15 credits per run, 3 runs per week. I'll add cost monitoring alerts.", 3},
				{"DATA-9", "admin", "That's ~$1,800/month. We need finance approval for any warehouse usage over $1,000/month. Have we submitted the request?", 2},
				{"DATA-9", "alice", "Finance approval submitted yesterday. Reference: `FIN-2024-3421`. Awaiting sign-off from the VP of Engineering.", 2},
				{"DATA-9", "bob", "I've configured the resource monitor to cap daily usage at 50 credits. That should prevent any runaway queries.", 2},
				{"DATA-9", "admin", "Good safeguard. Alice, once finance approves, we can enable the warehouse access immediately.", 1},
				{"DATA-9", "alice", "Finance approved this morning. Ready for warehouse access to be granted.", 1},

				// DATA-10: Redshift access for customer segmentation (2 days ago) — 10 comments
				{"DATA-10", "admin", "Bob, which IAM role will you be connecting through? We need to whitelist it in the Redshift cluster security config.", 2},
				{"DATA-10", "bob", "I'll be using the `analytics-readonly` IAM role. ARN: `arn:aws:iam::123456789012:role/analytics-readonly`.", 2},
				{"DATA-10", "alice", "I've used that role before for similar work. It has the right permissions for SELECT on those schemas.", 2},
				{"DATA-10", "admin", "Verified the role has no write permissions. Adding it to the Redshift cluster security group.", 1},
				{"DATA-10", "bob", "Also, I'll need the Metabase JDBC connection string. Is that documented somewhere?", 1},
				{"DATA-10", "alice", "It's in the analytics team wiki under 'Redshift Connections'. I'll DM you the link.", 1},
				{"DATA-10", "bob", "Got it, thanks. One more question — are there any query timeout limits on the analytics-readonly role?", 1},
				{"DATA-10", "admin", "Yes, there's a 300-second statement timeout and a 5GB memory limit per query. Should be sufficient for dashboard queries.", 1},
				{"DATA-10", "alice", "For the segmentation dashboards, I'd recommend using the `customer_events_daily_agg` materialized view. Much faster than raw events.", 0},
				{"DATA-10", "bob", "Good tip. I'll start with the aggregated view and only go to raw events if I need sub-daily granularity.", 0},

				// DATA-11: Looker license renewal (30 days ago) — 10 comments
				{"DATA-11", "admin", "Finance needs a business justification for the additional 5 viewer seats. Can someone provide the details?", 28},
				{"DATA-11", "alice", "We're onboarding 5 new BAs in Q4. They need dashboard access for the customer analytics project.", 28},
				{"DATA-11", "bob", "I can confirm — the BA team has been using shared credentials which is a compliance issue. Individual seats are required.", 27},
				{"DATA-11", "admin", "Good point on compliance. I'll flag that in the budget request to expedite approval.", 27},
				{"DATA-11", "alice", "The vendor offered a 10% discount if we renew before August 15th. That saves ~$3,000.", 26},
				{"DATA-11", "admin", "Noted. I'll push finance to approve before the discount deadline.", 25},
				{"DATA-11", "diana", "Can we also look into Looker's new embedded analytics tier? It might be cheaper for view-only users.", 25},
				{"DATA-11", "alice", "Good idea Diana. The embedded tier is $800/seat vs $1,200 for standard viewer. I'll check if it meets our needs.", 24},
				{"DATA-11", "bob", "Embedded tier doesn't support scheduled email reports. Two of the BAs need that feature. We'd need a mix.", 23},
				{"DATA-11", "admin", "Let's go with 3 standard viewers and 2 embedded. That optimizes cost while meeting all requirements.", 22},

				// DATA-12: PII dataset access for fraud analysis (1 day ago) — 8 comments
				{"DATA-12", "bob", "I've completed the required security awareness training. Certificate ID: `SAT-2024-4821`.", 1},
				{"DATA-12", "admin", "Need to verify the DPA is countersigned by the data protection officer before we can approve this.", 1},
				{"DATA-12", "alice", "I've forwarded the DPA to the DPO for countersignature. She's reviewing it today.", 1},
				{"DATA-12", "bob", "The fraud investigation is time-sensitive. We've identified a pattern affecting 47 accounts in the last week.", 1},
				{"DATA-12", "admin", "Understood the urgency. I've asked the DPO to prioritize the review. We should have the countersignature by EOD.", 0},
				{"DATA-12", "alice", "DPO signed off. The DPA is now fully executed. Document ID: `DPA-2024-0892`.", 0},
				{"DATA-12", "bob", "Great. Once access is granted, I'll start with IP correlation across the flagged accounts.", 0},
				{"DATA-12", "admin", "Approved with conditions: 30-day access window, audit logging enabled, and weekly progress reports required.", 0},

				// SEC-13: Vault admin access (18 days ago) — 8 comments
				{"SEC-13", "charlie", "Rotation policy is configured. All prod DB credentials now rotate every 30 days automatically.", 16},
				{"SEC-13", "admin", "Approved and verified. The rotation ran successfully in the last cycle.", 16},
				{"SEC-13", "diana", "I've reviewed the Vault audit logs. The dynamic secrets engine is generating credentials correctly.", 15},
				{"SEC-13", "charlie", "Application health checks are all green after the first rotation. No service disruptions.", 14},
				{"SEC-13", "admin", "Excellent. Let's schedule a review after the second rotation cycle to confirm stability.", 13},
				{"SEC-13", "diana", "Second rotation completed successfully. All applications reconnected within 5 seconds.", 11},
				{"SEC-13", "charlie", "I've documented the rotation procedure and added monitoring dashboards for credential lifecycle events.", 10},
				{"SEC-13", "admin", "Perfect. This is now fully operational. Closing this ticket. Great work Charlie and Diana.", 9},

				// SEC-14: SentinelOne console access (1 day ago) — 9 comments
				{"SEC-14", "charlie", "The alert on web-prod-03 shows suspicious outbound connections to an unknown IP. Need console access to isolate.", 1},
				{"SEC-14", "diana", "I've reviewed the network logs. Looks like a false positive from the new health check service, but we should verify.", 1},
				{"SEC-14", "admin", "Granting temporary access. Please document findings in the incident report template.", 1},
				{"SEC-14", "charlie", "Console access confirmed. Starting deep process analysis on web-prod-03.", 1},
				{"SEC-14", "diana", "I've correlated the outbound IP `198.51.100.42` with our CDN provider's health check infrastructure. Likely benign.", 0},
				{"SEC-14", "charlie", "Confirmed — the process tree shows it's our health check agent v2.1 that was deployed last Tuesday. It uses a new IP range.", 0},
				{"SEC-14", "admin", "Good to hear it's a false positive. Charlie, can you update the allowlist to prevent future alerts on this IP range?", 0},
				{"SEC-14", "charlie", "Allowlist updated. Added the entire CDN health check range `198.51.100.0/24` to the SentinelOne exclusion policy.", 0},
				{"SEC-14", "diana", "I'll also update our threat intel feed to include CDN provider IP ranges as known-good. That should reduce false positives.", 0},

				// SEC-15: Cloudflare WAF access (3 days ago) — 10 comments
				{"SEC-15", "admin", "Diana, do we have a staging zone in Cloudflare to test the rules before deploying to production?", 3},
				{"SEC-15", "diana", "Yes, `staging.example.com` mirrors the production zone config. I'll deploy there first.", 3},
				{"SEC-15", "charlie", "I can help with testing. I'll run our automated security scan suite against staging after the rules are deployed.", 2},
				{"SEC-15", "diana", "Rules deployed to staging. The OWASP CRS v4.0 update is in simulation mode for now.", 2},
				{"SEC-15", "bob", "FYI — our API gateway processes ~2M requests/day. Any WAF latency increase over 5ms will impact SLA.", 2},
				{"SEC-15", "charlie", "Scan complete. Two false positives detected on the `/api/v1/upload` endpoint. The multipart parser triggers rule 920230.", 1},
				{"SEC-15", "diana", "Good catch. I've added an exception for rule 920230 on the upload endpoint. Re-running scan.", 1},
				{"SEC-15", "admin", "Let's run the scan one more time after the exception is applied before we approve production deployment.", 1},
				{"SEC-15", "charlie", "Second scan clean — zero false positives. Latency impact measured at 2.3ms average, well within SLA.", 0},
				{"SEC-15", "diana", "Ready for production deployment. I'll schedule it for Tuesday's low-traffic window (2-4 AM UTC).", 0},

				// SEC-16: Pentest tools access (7 days ago) — 9 comments
				{"SEC-16", "diana", "I've prepared the scope document. Testing will be limited to the staging environment and pre-approved IP ranges.", 6},
				{"SEC-16", "charlie", "Burp Suite Pro license key received. Setting up the pentest VPN profile now.", 6},
				{"SEC-16", "admin", "Make sure the rules of engagement doc is signed by both teams before any active scanning begins.", 5},
				{"SEC-16", "charlie", "Signed ROE attached. We're ready to start the assessment on Monday.", 5},
				{"SEC-16", "diana", "VPN profile is configured. Test connection to staging successful — I can reach all in-scope targets.", 4},
				{"SEC-16", "admin", "Verified the ROE signatures. One question: does the scope include the WebSocket endpoints on `/ws/*`?", 4},
				{"SEC-16", "charlie", "Yes, WebSocket endpoints are in scope. We'll test for injection, auth bypass, and message tampering.", 3},
				{"SEC-16", "diana", "I've set up a dedicated Burp project with the target scope configured. Shared the project file in the team drive.", 3},
				{"SEC-16", "admin", "All clear. You're approved to begin scanning Monday 9 AM. Report any critical findings immediately via PagerDuty.", 2},

				// SEC-17: SOC analyst access revocation (9 days ago) — 10 comments
				{"SEC-17", "admin", "HR confirmed the role transfer was effective last Monday. We need to revoke all SOC access by Friday.", 8},
				{"SEC-17", "charlie", "I'll handle the Splunk and CrowdStrike deprovisioning. Diana, can you take PagerDuty and Jira?", 8},
				{"SEC-17", "diana", "On it. I'll also remove them from the `#soc-alerts` Slack channel and the SOC mailing list.", 7},
				{"SEC-17", "charlie", "Splunk access disabled. CrowdStrike responder role removed. Confirmed with both consoles.", 7},
				{"SEC-17", "diana", "PagerDuty schedule updated — removed from `soc-tier1` on-call rotation. Jira board access revoked.", 7},
				{"SEC-17", "admin", "What about any active investigations they were leading? We need to transfer ownership.", 6},
				{"SEC-17", "charlie", "Two active investigations reassigned: `INC-2024-312` to me, `INC-2024-318` to Diana. All caught up on context.", 6},
				{"SEC-17", "bob", "I checked the VPN logs — their last SOC VPN connection was 3 days ago. VPN access also needs revoking.", 5},
				{"SEC-17", "diana", "Good catch Bob. I've disabled their SOC VPN profile. Also removed from the security@example.com distribution list.", 5},
				{"SEC-17", "admin", "All access revoked and verified. I'll send the confirmation to HR and the new manager. Good job team.", 4},
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
			// 13. Backdate timestamps to simulate realistic timing.
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
