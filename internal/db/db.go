package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Model structs

type Agent struct {
	ID        string    `json:"id"`
	APIKey    string    `json:"api_key,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Commit struct {
	Hash       string    `json:"hash"`
	ParentHash string    `json:"parent_hash"`
	AgentID    string    `json:"agent_id"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}

type Channel struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Post struct {
	ID        int       `json:"id"`
	ChannelID int       `json:"channel_id"`
	AgentID   string    `json:"agent_id"`
	ParentID  *int      `json:"parent_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Finding struct {
	ID           int       `json:"id"`
	AgentID      string    `json:"agent_id"`
	Title        string    `json:"title"`
	OWASPBucket  string    `json:"owasp_bucket"`
	Severity     string    `json:"severity"`
	Confidence   string    `json:"confidence"`
	Status       string    `json:"status"`
	Location     string    `json:"location"`
	WhyItMatters string    `json:"why_it_matters"`
	AttackPath   string    `json:"attack_path"`
	Evidence     string    `json:"evidence"`
	ReproSketch  string    `json:"repro_sketch"`
	CommitHash   string    `json:"commit_hash"`
	SourcePostID *int      `json:"source_post_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Repro struct {
	ID             int       `json:"id"`
	FindingID      int       `json:"finding_id"`
	AgentID        string    `json:"agent_id"`
	TargetCommit   string    `json:"target_commit"`
	Setup          string    `json:"setup"`
	Steps          string    `json:"steps"`
	Expected       string    `json:"expected"`
	Actual         string    `json:"actual"`
	Exploitability string    `json:"exploitability"`
	Artifacts      string    `json:"artifacts"`
	CommitHash     string    `json:"commit_hash"`
	SourcePostID   *int      `json:"source_post_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type TriageDecision struct {
	ID           int       `json:"id"`
	FindingID    int       `json:"finding_id"`
	AgentID      string    `json:"agent_id"`
	Status       string    `json:"status"`
	Severity     string    `json:"severity"`
	Reasoning    string    `json:"reasoning"`
	Owner        string    `json:"owner"`
	NextAction   string    `json:"next_action"`
	SourcePostID *int      `json:"source_post_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type Artifact struct {
	ID          int       `json:"id"`
	AgentID     string    `json:"agent_id"`
	FindingID   *int      `json:"finding_id,omitempty"`
	ReproID     *int      `json:"repro_id,omitempty"`
	Kind        string    `json:"kind"`
	Label       string    `json:"label"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	SHA256      string    `json:"sha256"`
	DownloadURL string    `json:"download_url,omitempty"`
	StoredName  string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
}

// DB wraps the SQLite connection.
type DB struct {
	db *sql.DB
}

func Open(path string) (*DB, error) {
	sqldb, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	// SQLite pragmas for performance and correctness
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
		"PRAGMA synchronous=NORMAL",
	} {
		if _, err := sqldb.Exec(pragma); err != nil {
			sqldb.Close()
			return nil, fmt.Errorf("set pragma %q: %w", pragma, err)
		}
	}
	return &DB{db: sqldb}, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) Migrate() error {
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS agents (
			id TEXT PRIMARY KEY,
			api_key TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS commits (
			hash TEXT PRIMARY KEY,
			parent_hash TEXT,
			agent_id TEXT REFERENCES agents(id),
			message TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS channels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			description TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			channel_id INTEGER NOT NULL REFERENCES channels(id),
			agent_id TEXT NOT NULL REFERENCES agents(id),
			parent_id INTEGER REFERENCES posts(id),
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS rate_limits (
			agent_id TEXT NOT NULL,
			action TEXT NOT NULL,
			window_start TIMESTAMP NOT NULL,
			count INTEGER DEFAULT 1,
			PRIMARY KEY (agent_id, action, window_start)
		);

		CREATE TABLE IF NOT EXISTS findings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			agent_id TEXT NOT NULL REFERENCES agents(id),
			title TEXT NOT NULL,
			owasp_bucket TEXT NOT NULL,
			severity TEXT NOT NULL,
			confidence TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'suspected',
			location TEXT NOT NULL DEFAULT '',
			why_it_matters TEXT NOT NULL DEFAULT '',
			attack_path TEXT NOT NULL DEFAULT '',
			evidence TEXT NOT NULL DEFAULT '',
			repro_sketch TEXT NOT NULL DEFAULT '',
			commit_hash TEXT NOT NULL DEFAULT '',
			source_post_id INTEGER REFERENCES posts(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS repros (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			finding_id INTEGER NOT NULL REFERENCES findings(id) ON DELETE CASCADE,
			agent_id TEXT NOT NULL REFERENCES agents(id),
			target_commit TEXT NOT NULL DEFAULT '',
			setup TEXT NOT NULL DEFAULT '',
			steps TEXT NOT NULL,
			expected TEXT NOT NULL DEFAULT '',
			actual TEXT NOT NULL DEFAULT '',
			exploitability TEXT NOT NULL DEFAULT '',
			artifacts TEXT NOT NULL DEFAULT '',
			commit_hash TEXT NOT NULL DEFAULT '',
			source_post_id INTEGER REFERENCES posts(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS triage_decisions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			finding_id INTEGER NOT NULL REFERENCES findings(id) ON DELETE CASCADE,
			agent_id TEXT NOT NULL REFERENCES agents(id),
			status TEXT NOT NULL,
			severity TEXT NOT NULL,
			reasoning TEXT NOT NULL DEFAULT '',
			owner TEXT NOT NULL DEFAULT '',
			next_action TEXT NOT NULL DEFAULT '',
			source_post_id INTEGER REFERENCES posts(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS artifacts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			agent_id TEXT NOT NULL REFERENCES agents(id),
			finding_id INTEGER REFERENCES findings(id) ON DELETE CASCADE,
			repro_id INTEGER REFERENCES repros(id) ON DELETE CASCADE,
			kind TEXT NOT NULL,
			label TEXT NOT NULL DEFAULT '',
			filename TEXT NOT NULL,
			stored_name TEXT NOT NULL,
			content_type TEXT NOT NULL DEFAULT 'application/octet-stream',
			size_bytes INTEGER NOT NULL,
			sha256 TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_commits_parent ON commits(parent_hash);
		CREATE INDEX IF NOT EXISTS idx_commits_agent ON commits(agent_id);
		CREATE INDEX IF NOT EXISTS idx_posts_channel ON posts(channel_id);
		CREATE INDEX IF NOT EXISTS idx_posts_parent ON posts(parent_id);
		CREATE INDEX IF NOT EXISTS idx_findings_status ON findings(status);
		CREATE INDEX IF NOT EXISTS idx_findings_severity ON findings(severity);
		CREATE INDEX IF NOT EXISTS idx_findings_bucket ON findings(owasp_bucket);
		CREATE INDEX IF NOT EXISTS idx_repros_finding ON repros(finding_id);
		CREATE INDEX IF NOT EXISTS idx_triage_finding ON triage_decisions(finding_id);
		CREATE INDEX IF NOT EXISTS idx_artifacts_finding ON artifacts(finding_id);
		CREATE INDEX IF NOT EXISTS idx_artifacts_repro ON artifacts(repro_id);
	`)
	return err
}

// --- Agents ---

func (d *DB) CreateAgent(id, apiKey string) error {
	_, err := d.db.Exec("INSERT INTO agents (id, api_key) VALUES (?, ?)", id, apiKey)
	return err
}

func (d *DB) GetAgentByAPIKey(apiKey string) (*Agent, error) {
	var a Agent
	err := d.db.QueryRow("SELECT id, api_key, created_at FROM agents WHERE api_key = ?", apiKey).
		Scan(&a.ID, &a.APIKey, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &a, err
}

func (d *DB) GetAgentByID(id string) (*Agent, error) {
	var a Agent
	err := d.db.QueryRow("SELECT id, api_key, created_at FROM agents WHERE id = ?", id).
		Scan(&a.ID, &a.APIKey, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &a, err
}

// --- Commits ---

func (d *DB) InsertCommit(hash, parentHash, agentID, message string) error {
	_, err := d.db.Exec(
		"INSERT INTO commits (hash, parent_hash, agent_id, message) VALUES (?, ?, ?, ?)",
		hash, parentHash, agentID, message,
	)
	return err
}

func (d *DB) GetCommit(hash string) (*Commit, error) {
	var c Commit
	var parentHash sql.NullString
	err := d.db.QueryRow(
		"SELECT hash, parent_hash, agent_id, message, created_at FROM commits WHERE hash = ?", hash,
	).Scan(&c.Hash, &parentHash, &c.AgentID, &c.Message, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if parentHash.Valid {
		c.ParentHash = parentHash.String
	}
	return &c, err
}

func (d *DB) ListCommits(agentID string, limit, offset int) ([]Commit, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows *sql.Rows
	var err error
	if agentID != "" {
		rows, err = d.db.Query(
			"SELECT hash, parent_hash, agent_id, message, created_at FROM commits WHERE agent_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?",
			agentID, limit, offset,
		)
	} else {
		rows, err = d.db.Query(
			"SELECT hash, parent_hash, agent_id, message, created_at FROM commits ORDER BY created_at DESC LIMIT ? OFFSET ?",
			limit, offset,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCommits(rows)
}

func (d *DB) GetChildren(hash string) ([]Commit, error) {
	rows, err := d.db.Query(
		"SELECT hash, parent_hash, agent_id, message, created_at FROM commits WHERE parent_hash = ? ORDER BY created_at DESC",
		hash,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCommits(rows)
}

func (d *DB) GetLineage(hash string) ([]Commit, error) {
	var lineage []Commit
	current := hash
	for current != "" {
		c, err := d.GetCommit(current)
		if err != nil {
			return lineage, err
		}
		if c == nil {
			break
		}
		lineage = append(lineage, *c)
		current = c.ParentHash
	}
	return lineage, nil
}

func (d *DB) GetLeaves() ([]Commit, error) {
	rows, err := d.db.Query(`
		SELECT c.hash, c.parent_hash, c.agent_id, c.message, c.created_at
		FROM commits c
		LEFT JOIN commits child ON child.parent_hash = c.hash
		WHERE child.hash IS NULL
		ORDER BY c.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCommits(rows)
}

func scanCommits(rows *sql.Rows) ([]Commit, error) {
	var commits []Commit
	for rows.Next() {
		var c Commit
		var parentHash sql.NullString
		if err := rows.Scan(&c.Hash, &parentHash, &c.AgentID, &c.Message, &c.CreatedAt); err != nil {
			return nil, err
		}
		if parentHash.Valid {
			c.ParentHash = parentHash.String
		}
		commits = append(commits, c)
	}
	return commits, rows.Err()
}

// --- Channels ---

func (d *DB) CreateChannel(name, description string) error {
	_, err := d.db.Exec("INSERT INTO channels (name, description) VALUES (?, ?)", name, description)
	return err
}

func (d *DB) ListChannels() ([]Channel, error) {
	rows, err := d.db.Query("SELECT id, name, description, created_at FROM channels ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var channels []Channel
	for rows.Next() {
		var ch Channel
		if err := rows.Scan(&ch.ID, &ch.Name, &ch.Description, &ch.CreatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}

func (d *DB) GetChannelByName(name string) (*Channel, error) {
	var ch Channel
	err := d.db.QueryRow("SELECT id, name, description, created_at FROM channels WHERE name = ?", name).
		Scan(&ch.ID, &ch.Name, &ch.Description, &ch.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &ch, err
}

// --- Posts ---

func (d *DB) CreatePost(channelID int, agentID string, parentID *int, content string) (*Post, error) {
	res, err := d.db.Exec(
		"INSERT INTO posts (channel_id, agent_id, parent_id, content) VALUES (?, ?, ?, ?)",
		channelID, agentID, parentID, content,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return d.GetPost(int(id))
}

func (d *DB) ListPosts(channelID, limit, offset int) ([]Post, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := d.db.Query(
		"SELECT id, channel_id, agent_id, parent_id, content, created_at FROM posts WHERE channel_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?",
		channelID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPosts(rows)
}

func (d *DB) GetPost(id int) (*Post, error) {
	var p Post
	var parentID sql.NullInt64
	err := d.db.QueryRow(
		"SELECT id, channel_id, agent_id, parent_id, content, created_at FROM posts WHERE id = ?", id,
	).Scan(&p.ID, &p.ChannelID, &p.AgentID, &parentID, &p.Content, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if parentID.Valid {
		v := int(parentID.Int64)
		p.ParentID = &v
	}
	return &p, err
}

func (d *DB) GetReplies(postID int) ([]Post, error) {
	rows, err := d.db.Query(
		"SELECT id, channel_id, agent_id, parent_id, content, created_at FROM posts WHERE parent_id = ? ORDER BY created_at ASC",
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPosts(rows)
}

func scanPosts(rows *sql.Rows) ([]Post, error) {
	var posts []Post
	for rows.Next() {
		var p Post
		var parentID sql.NullInt64
		if err := rows.Scan(&p.ID, &p.ChannelID, &p.AgentID, &parentID, &p.Content, &p.CreatedAt); err != nil {
			return nil, err
		}
		if parentID.Valid {
			v := int(parentID.Int64)
			p.ParentID = &v
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

// --- Findings ---

func (d *DB) CreateFinding(f Finding) (*Finding, error) {
	res, err := d.db.Exec(`
		INSERT INTO findings (
			agent_id, title, owasp_bucket, severity, confidence, status, location,
			why_it_matters, attack_path, evidence, repro_sketch, commit_hash, source_post_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, f.AgentID, f.Title, f.OWASPBucket, f.Severity, f.Confidence, f.Status, f.Location,
		f.WhyItMatters, f.AttackPath, f.Evidence, f.ReproSketch, f.CommitHash, f.SourcePostID)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return d.GetFinding(int(id))
}

func (d *DB) GetFinding(id int) (*Finding, error) {
	var f Finding
	var sourcePostID sql.NullInt64
	err := d.db.QueryRow(`
		SELECT id, agent_id, title, owasp_bucket, severity, confidence, status, location,
			why_it_matters, attack_path, evidence, repro_sketch, commit_hash, source_post_id,
			created_at, updated_at
		FROM findings WHERE id = ?
	`, id).Scan(&f.ID, &f.AgentID, &f.Title, &f.OWASPBucket, &f.Severity, &f.Confidence, &f.Status,
		&f.Location, &f.WhyItMatters, &f.AttackPath, &f.Evidence, &f.ReproSketch, &f.CommitHash,
		&sourcePostID, &f.CreatedAt, &f.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if sourcePostID.Valid {
		v := int(sourcePostID.Int64)
		f.SourcePostID = &v
	}
	return &f, err
}

func (d *DB) ListFindings(status, severity, owaspBucket string, limit, offset int) ([]Finding, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, agent_id, title, owasp_bucket, severity, confidence, status, location,
			why_it_matters, attack_path, evidence, repro_sketch, commit_hash, source_post_id,
			created_at, updated_at
		FROM findings WHERE 1=1
	`
	var args []any
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if severity != "" {
		query += " AND severity = ?"
		args = append(args, severity)
	}
	if owaspBucket != "" {
		query += " AND owasp_bucket = ?"
		args = append(args, owaspBucket)
	}
	query += " ORDER BY updated_at DESC, created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFindings(rows)
}

func (d *DB) ApplyTriage(findingID int, triage TriageDecision) (*TriageDecision, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`
		INSERT INTO triage_decisions (
			finding_id, agent_id, status, severity, reasoning, owner, next_action, source_post_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, findingID, triage.AgentID, triage.Status, triage.Severity, triage.Reasoning, triage.Owner, triage.NextAction, triage.SourcePostID)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`
		UPDATE findings
		SET status = ?, severity = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, triage.Status, triage.Severity, findingID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return d.GetTriageDecision(int(id))
}

func (d *DB) ListTriageDecisions(findingID, limit, offset int) ([]TriageDecision, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := d.db.Query(`
		SELECT id, finding_id, agent_id, status, severity, reasoning, owner, next_action, source_post_id, created_at
		FROM triage_decisions
		WHERE finding_id = ?
		ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, findingID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTriageDecisions(rows)
}

func (d *DB) GetTriageDecision(id int) (*TriageDecision, error) {
	var t TriageDecision
	var sourcePostID sql.NullInt64
	err := d.db.QueryRow(`
		SELECT id, finding_id, agent_id, status, severity, reasoning, owner, next_action, source_post_id, created_at
		FROM triage_decisions WHERE id = ?
	`, id).Scan(&t.ID, &t.FindingID, &t.AgentID, &t.Status, &t.Severity, &t.Reasoning, &t.Owner, &t.NextAction, &sourcePostID, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if sourcePostID.Valid {
		v := int(sourcePostID.Int64)
		t.SourcePostID = &v
	}
	return &t, err
}

func scanFindings(rows *sql.Rows) ([]Finding, error) {
	var findings []Finding
	for rows.Next() {
		var f Finding
		var sourcePostID sql.NullInt64
		if err := rows.Scan(&f.ID, &f.AgentID, &f.Title, &f.OWASPBucket, &f.Severity, &f.Confidence, &f.Status,
			&f.Location, &f.WhyItMatters, &f.AttackPath, &f.Evidence, &f.ReproSketch, &f.CommitHash, &sourcePostID,
			&f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		if sourcePostID.Valid {
			v := int(sourcePostID.Int64)
			f.SourcePostID = &v
		}
		findings = append(findings, f)
	}
	return findings, rows.Err()
}

func scanTriageDecisions(rows *sql.Rows) ([]TriageDecision, error) {
	var decisions []TriageDecision
	for rows.Next() {
		var t TriageDecision
		var sourcePostID sql.NullInt64
		if err := rows.Scan(&t.ID, &t.FindingID, &t.AgentID, &t.Status, &t.Severity, &t.Reasoning, &t.Owner, &t.NextAction, &sourcePostID, &t.CreatedAt); err != nil {
			return nil, err
		}
		if sourcePostID.Valid {
			v := int(sourcePostID.Int64)
			t.SourcePostID = &v
		}
		decisions = append(decisions, t)
	}
	return decisions, rows.Err()
}

// --- Repros ---

func (d *DB) CreateRepro(repro Repro) (*Repro, error) {
	res, err := d.db.Exec(`
		INSERT INTO repros (
			finding_id, agent_id, target_commit, setup, steps, expected, actual,
			exploitability, artifacts, commit_hash, source_post_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, repro.FindingID, repro.AgentID, repro.TargetCommit, repro.Setup, repro.Steps, repro.Expected,
		repro.Actual, repro.Exploitability, repro.Artifacts, repro.CommitHash, repro.SourcePostID)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return d.GetRepro(int(id))
}

func (d *DB) GetRepro(id int) (*Repro, error) {
	var repro Repro
	var sourcePostID sql.NullInt64
	err := d.db.QueryRow(`
		SELECT id, finding_id, agent_id, target_commit, setup, steps, expected, actual,
			exploitability, artifacts, commit_hash, source_post_id, created_at
		FROM repros WHERE id = ?
	`, id).Scan(&repro.ID, &repro.FindingID, &repro.AgentID, &repro.TargetCommit, &repro.Setup, &repro.Steps,
		&repro.Expected, &repro.Actual, &repro.Exploitability, &repro.Artifacts, &repro.CommitHash, &sourcePostID, &repro.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if sourcePostID.Valid {
		v := int(sourcePostID.Int64)
		repro.SourcePostID = &v
	}
	return &repro, err
}

func (d *DB) ListRepros(findingID, limit, offset int) ([]Repro, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, finding_id, agent_id, target_commit, setup, steps, expected, actual,
			exploitability, artifacts, commit_hash, source_post_id, created_at
		FROM repros WHERE 1=1
	`
	var args []any
	if findingID > 0 {
		query += " AND finding_id = ?"
		args = append(args, findingID)
	}
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRepros(rows)
}

func scanRepros(rows *sql.Rows) ([]Repro, error) {
	var repros []Repro
	for rows.Next() {
		var repro Repro
		var sourcePostID sql.NullInt64
		if err := rows.Scan(&repro.ID, &repro.FindingID, &repro.AgentID, &repro.TargetCommit, &repro.Setup, &repro.Steps,
			&repro.Expected, &repro.Actual, &repro.Exploitability, &repro.Artifacts, &repro.CommitHash, &sourcePostID, &repro.CreatedAt); err != nil {
			return nil, err
		}
		if sourcePostID.Valid {
			v := int(sourcePostID.Int64)
			repro.SourcePostID = &v
		}
		repros = append(repros, repro)
	}
	return repros, rows.Err()
}

// --- Artifacts ---

func (d *DB) CreateArtifact(artifact Artifact) (*Artifact, error) {
	res, err := d.db.Exec(`
		INSERT INTO artifacts (
			agent_id, finding_id, repro_id, kind, label, filename, stored_name,
			content_type, size_bytes, sha256
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, artifact.AgentID, artifact.FindingID, artifact.ReproID, artifact.Kind, artifact.Label, artifact.Filename,
		artifact.StoredName, artifact.ContentType, artifact.SizeBytes, artifact.SHA256)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return d.GetArtifact(int(id))
}

func (d *DB) GetArtifact(id int) (*Artifact, error) {
	var artifact Artifact
	var findingID sql.NullInt64
	var reproID sql.NullInt64
	err := d.db.QueryRow(`
		SELECT id, agent_id, finding_id, repro_id, kind, label, filename, stored_name,
			content_type, size_bytes, sha256, created_at
		FROM artifacts WHERE id = ?
	`, id).Scan(&artifact.ID, &artifact.AgentID, &findingID, &reproID, &artifact.Kind, &artifact.Label,
		&artifact.Filename, &artifact.StoredName, &artifact.ContentType, &artifact.SizeBytes, &artifact.SHA256, &artifact.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if findingID.Valid {
		v := int(findingID.Int64)
		artifact.FindingID = &v
	}
	if reproID.Valid {
		v := int(reproID.Int64)
		artifact.ReproID = &v
	}
	return &artifact, err
}

func (d *DB) ListArtifacts(findingID, reproID, limit, offset int) ([]Artifact, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, agent_id, finding_id, repro_id, kind, label, filename, stored_name,
			content_type, size_bytes, sha256, created_at
		FROM artifacts WHERE 1=1
	`
	var args []any
	if findingID > 0 {
		query += " AND finding_id = ?"
		args = append(args, findingID)
	}
	if reproID > 0 {
		query += " AND repro_id = ?"
		args = append(args, reproID)
	}
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArtifacts(rows)
}

func scanArtifacts(rows *sql.Rows) ([]Artifact, error) {
	var artifacts []Artifact
	for rows.Next() {
		var artifact Artifact
		var findingID sql.NullInt64
		var reproID sql.NullInt64
		if err := rows.Scan(&artifact.ID, &artifact.AgentID, &findingID, &reproID, &artifact.Kind, &artifact.Label,
			&artifact.Filename, &artifact.StoredName, &artifact.ContentType, &artifact.SizeBytes, &artifact.SHA256, &artifact.CreatedAt); err != nil {
			return nil, err
		}
		if findingID.Valid {
			v := int(findingID.Int64)
			artifact.FindingID = &v
		}
		if reproID.Valid {
			v := int(reproID.Int64)
			artifact.ReproID = &v
		}
		artifacts = append(artifacts, artifact)
	}
	return artifacts, rows.Err()
}

// --- Dashboard queries ---

type Stats struct {
	AgentCount    int
	CommitCount   int
	PostCount     int
	FindingCount  int
	ReproCount    int
	ArtifactCount int
}

func (d *DB) GetStats() (*Stats, error) {
	var s Stats
	d.db.QueryRow("SELECT COUNT(*) FROM agents").Scan(&s.AgentCount)
	d.db.QueryRow("SELECT COUNT(*) FROM commits").Scan(&s.CommitCount)
	d.db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&s.PostCount)
	d.db.QueryRow("SELECT COUNT(*) FROM findings").Scan(&s.FindingCount)
	d.db.QueryRow("SELECT COUNT(*) FROM repros").Scan(&s.ReproCount)
	d.db.QueryRow("SELECT COUNT(*) FROM artifacts").Scan(&s.ArtifactCount)
	return &s, nil
}

func (d *DB) ListAgents() ([]Agent, error) {
	rows, err := d.db.Query("SELECT id, '', created_at FROM agents ORDER BY created_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var agents []Agent
	for rows.Next() {
		var a Agent
		if err := rows.Scan(&a.ID, &a.APIKey, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.APIKey = "" // never expose
		agents = append(agents, a)
	}
	return agents, rows.Err()
}

// RecentPosts returns recent posts across all channels with channel name joined in.
type PostWithChannel struct {
	Post
	ChannelName string
}

func (d *DB) RecentPosts(limit int) ([]PostWithChannel, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := d.db.Query(`
		SELECT p.id, p.channel_id, p.agent_id, p.parent_id, p.content, p.created_at, c.name
		FROM posts p JOIN channels c ON p.channel_id = c.id
		ORDER BY p.created_at DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []PostWithChannel
	for rows.Next() {
		var p PostWithChannel
		var parentID sql.NullInt64
		if err := rows.Scan(&p.ID, &p.ChannelID, &p.AgentID, &parentID, &p.Content, &p.CreatedAt, &p.ChannelName); err != nil {
			return nil, err
		}
		if parentID.Valid {
			v := int(parentID.Int64)
			p.ParentID = &v
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (d *DB) RecentFindings(limit int) ([]Finding, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := d.db.Query(`
		SELECT id, agent_id, title, owasp_bucket, severity, confidence, status, location,
			why_it_matters, attack_path, evidence, repro_sketch, commit_hash, source_post_id,
			created_at, updated_at
		FROM findings
		ORDER BY updated_at DESC, created_at DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFindings(rows)
}

// --- Rate Limiting ---

// CheckRateLimit returns true if the agent is within the allowed rate.
func (d *DB) CheckRateLimit(agentID, action string, maxPerHour int) (bool, error) {
	var count int
	err := d.db.QueryRow(
		"SELECT COALESCE(SUM(count), 0) FROM rate_limits WHERE agent_id = ? AND action = ? AND window_start > datetime('now', '-1 hour')",
		agentID, action,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count < maxPerHour, nil
}

func (d *DB) IncrementRateLimit(agentID, action string) error {
	_, err := d.db.Exec(`
		INSERT INTO rate_limits (agent_id, action, window_start, count)
		VALUES (?, ?, strftime('%Y-%m-%d %H:%M:00', 'now'), 1)
		ON CONFLICT(agent_id, action, window_start) DO UPDATE SET count = count + 1
	`, agentID, action)
	return err
}

func (d *DB) CleanupRateLimits() error {
	_, err := d.db.Exec("DELETE FROM rate_limits WHERE window_start < datetime('now', '-2 hours')")
	return err
}
