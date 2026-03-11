package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func (c *Client) postMultipartFile(path, fieldName, filePath string, fields map[string]string) (*http.Response, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}
	for k, v := range fields {
		if strings.TrimSpace(v) == "" {
			continue
		}
		if err := writer.WriteField(k, v); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.BaseURL+path, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return c.HTTP.Do(req)
}

func cmdDoctor(args []string) {
	fs := flag.NewFlagSet("doctor", flag.ExitOnError)
	serverFlag := fs.String("server", "", "server URL override")
	fs.Parse(args)

	var cfg *CLIConfig
	cfg, _ = loadConfig()

	okCount := 0
	warnCount := 0
	report := func(status, label, detail string) {
		fmt.Printf("%-5s %-18s %s\n", status, label, detail)
		if status == "OK" {
			okCount++
		} else {
			warnCount++
		}
	}

	for _, command := range []string{"git", "npm", "npx", "codex", "claude", "agent-browser"} {
		if path, err := exec.LookPath(command); err == nil {
			report("OK", "binary:"+command, path)
		} else {
			report("WARN", "binary:"+command, "not found on PATH")
		}
	}

	if cfg != nil {
		report("OK", "config", configPath()+" ("+cfg.AgentID+")")
	} else {
		report("WARN", "config", "missing; run 'ah join' first")
	}

	if root, err := gitOutput("rev-parse", "--show-toplevel"); err == nil {
		branch, _ := gitOutput("branch", "--show-current")
		report("OK", "git", strings.TrimSpace(root)+" ["+strings.TrimSpace(branch)+"]")
	} else {
		report("WARN", "git", "current directory is not a git repo")
	}

	serverURL := strings.TrimRight(*serverFlag, "/")
	if serverURL == "" && cfg != nil {
		serverURL = strings.TrimRight(cfg.ServerURL, "/")
	}
	if serverURL == "" {
		report("WARN", "server", "no server configured; pass --server or run 'ah join'")
	} else {
		resp, err := http.Get(serverURL + "/api/health")
		if err != nil {
			report("WARN", "health", err.Error())
		} else {
			var body map[string]string
			if err := readJSON(resp, &body); err != nil {
				report("WARN", "health", err.Error())
			} else {
				report("OK", "health", serverURL+" -> "+body["status"])
			}
		}
		if cfg != nil {
			client := newClient(cfg)
			resp, err := client.get("/api/channels")
			if err != nil {
				report("WARN", "auth", err.Error())
			} else {
				var channels []map[string]any
				if err := readJSON(resp, &channels); err != nil {
					report("WARN", "auth", err.Error())
				} else {
					report("OK", "auth", fmt.Sprintf("agent %s can read %d channels", cfg.AgentID, len(channels)))
				}
			}
		}
	}

	for _, path := range []string{"AGENTS.md", "CLAUDE.md", ".cursor/skills/agenthub-pentest-browser-validation/SKILL.md"} {
		if _, err := os.Stat(path); err == nil {
			report("OK", "file", path)
		} else {
			report("WARN", "file", path+" missing")
		}
	}

	fmt.Printf("\nsummary: %d ok, %d warnings\n", okCount, warnCount)
}

func cmdInstallTools(args []string) {
	fs := flag.NewFlagSet("install-tools", flag.ExitOnError)
	all := fs.Bool("all", false, "install all supported local tools")
	browser := fs.Bool("browser", false, "install agent-browser globally via npm")
	browserDeps := fs.Bool("browser-deps", false, "install browser runtime dependencies")
	browserSkill := fs.Bool("browser-skill", false, "install the Vercel agent-browser skill")
	fs.Parse(args)

	if !*all && !*browser && !*browserDeps && !*browserSkill {
		*all = true
	}
	if *all {
		*browser = true
		*browserDeps = true
		*browserSkill = true
	}

	run := func(name string, args ...string) {
		cmd := exec.Command(name, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			fatal("%s %s failed: %v", name, strings.Join(args, " "), err)
		}
	}

	if *browser {
		if _, err := exec.LookPath("npm"); err != nil {
			fatal("npm is required to install agent-browser")
		}
		run("npm", "install", "-g", "agent-browser")
	}
	if *browserDeps {
		if _, err := exec.LookPath("agent-browser"); err != nil {
			fatal("agent-browser must be on PATH before running --browser-deps")
		}
		run("agent-browser", "install")
	}
	if *browserSkill {
		if _, err := exec.LookPath("npx"); err != nil {
			fatal("npx is required to install the agent-browser skill")
		}
		run("npx", "skills", "add", "vercel-labs/agent-browser", "--skill", "agent-browser")
	}

	fmt.Println("done. use 'ah doctor' to verify local tool availability.")
	fmt.Println("codex and claude are reported by doctor but not auto-installed here because their install paths vary by environment.")
}

func cmdFindingCreate(args []string) {
	fs := flag.NewFlagSet("finding-create", flag.ExitOnError)
	title := fs.String("title", "", "finding title")
	owasp := fs.String("owasp", "", "OWASP bucket, e.g. A01")
	severity := fs.String("severity", "medium", "severity")
	confidence := fs.String("confidence", "medium", "confidence")
	status := fs.String("status", "suspected", "status")
	location := fs.String("location", "", "location")
	why := fs.String("why", "", "why it matters")
	attackPath := fs.String("attack-path", "", "attack path")
	evidence := fs.String("evidence", "", "evidence")
	reproSketch := fs.String("repro-sketch", "", "repro sketch")
	commitHash := fs.String("commit", "", "commit hash")
	sourcePostID := fs.Int("source-post", 0, "source post id")
	fs.Parse(args)

	cfg := mustLoadConfig()
	client := newClient(cfg)
	body := map[string]any{
		"title":          *title,
		"owasp_bucket":   *owasp,
		"severity":       *severity,
		"confidence":     *confidence,
		"status":         *status,
		"location":       *location,
		"why_it_matters": *why,
		"attack_path":    *attackPath,
		"evidence":       *evidence,
		"repro_sketch":   *reproSketch,
		"commit_hash":    *commitHash,
	}
	if *sourcePostID > 0 {
		body["source_post_id"] = *sourcePostID
	}
	resp, err := client.postJSON("/api/findings", body)
	if err != nil {
		fatal("finding create failed: %v", err)
	}
	var finding map[string]any
	if err := readJSON(resp, &finding); err != nil {
		fatal("finding create failed: %v", err)
	}
	fmt.Printf("created finding #%v [%s/%s] %s\n", finding["id"], finding["severity"], finding["status"], str(finding["title"]))
}

func cmdFindings(args []string) {
	fs := flag.NewFlagSet("findings", flag.ExitOnError)
	status := fs.String("status", "", "filter by status")
	severity := fs.String("severity", "", "filter by severity")
	owasp := fs.String("owasp", "", "filter by OWASP bucket")
	limit := fs.Int("limit", 20, "max results")
	fs.Parse(args)

	cfg := mustLoadConfig()
	client := newClient(cfg)
	path := fmt.Sprintf("/api/findings?limit=%d", *limit)
	if *status != "" {
		path += "&status=" + *status
	}
	if *severity != "" {
		path += "&severity=" + *severity
	}
	if *owasp != "" {
		path += "&owasp=" + *owasp
	}
	resp, err := client.get(path)
	if err != nil {
		fatal("request failed: %v", err)
	}
	var findings []map[string]any
	if err := readJSON(resp, &findings); err != nil {
		fatal("failed: %v", err)
	}
	if len(findings) == 0 {
		fmt.Println("no findings")
		return
	}
	for _, f := range findings {
		fmt.Printf("#%-4v %-8s %-20s %-16s %s\n", f["id"], str(f["severity"]), str(f["status"]), str(f["owasp_bucket"]), str(f["title"]))
	}
}

func cmdFindingGet(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: ah finding-get <id>")
		os.Exit(1)
	}
	id := args[0]
	cfg := mustLoadConfig()
	client := newClient(cfg)

	getJSON := func(path string, target any) {
		resp, err := client.get(path)
		if err != nil {
			fatal("request failed: %v", err)
		}
		if err := readJSON(resp, target); err != nil {
			fatal("request failed: %v", err)
		}
	}

	var finding map[string]any
	getJSON("/api/findings/"+id, &finding)
	var repros []map[string]any
	getJSON("/api/repros?finding_id="+id, &repros)
	var triage []map[string]any
	getJSON("/api/findings/"+id+"/triage", &triage)
	var artifacts []map[string]any
	getJSON("/api/artifacts?finding_id="+id, &artifacts)

	fmt.Printf("Finding #%v\n", finding["id"])
	fmt.Printf("Title: %s\n", str(finding["title"]))
	fmt.Printf("OWASP: %s\n", str(finding["owasp_bucket"]))
	fmt.Printf("Severity: %s\n", str(finding["severity"]))
	fmt.Printf("Confidence: %s\n", str(finding["confidence"]))
	fmt.Printf("Status: %s\n", str(finding["status"]))
	fmt.Printf("Location: %s\n", str(finding["location"]))
	fmt.Printf("Why it matters: %s\n", str(finding["why_it_matters"]))
	fmt.Printf("Attack path: %s\n", str(finding["attack_path"]))
	fmt.Printf("Evidence: %s\n", str(finding["evidence"]))
	fmt.Printf("Repro sketch: %s\n", str(finding["repro_sketch"]))
	if str(finding["commit_hash"]) != "" {
		fmt.Printf("Commit: %s\n", str(finding["commit_hash"]))
	}
	fmt.Printf("Repros: %d | Triage decisions: %d | Artifacts: %d\n", len(repros), len(triage), len(artifacts))
}

func cmdReproCreate(args []string) {
	fs := flag.NewFlagSet("repro-create", flag.ExitOnError)
	findingID := fs.Int("finding", 0, "finding id")
	targetCommit := fs.String("target-commit", "", "target commit")
	setup := fs.String("setup", "", "setup text")
	steps := fs.String("steps", "", "steps")
	expected := fs.String("expected", "", "expected result")
	actual := fs.String("actual", "", "actual result")
	exploitability := fs.String("exploitability", "", "exploitability")
	artifacts := fs.String("artifacts", "", "artifact references")
	commitHash := fs.String("commit", "", "commit hash")
	sourcePostID := fs.Int("source-post", 0, "source post id")
	fs.Parse(args)

	cfg := mustLoadConfig()
	client := newClient(cfg)
	body := map[string]any{
		"finding_id":     *findingID,
		"target_commit":  *targetCommit,
		"setup":          *setup,
		"steps":          *steps,
		"expected":       *expected,
		"actual":         *actual,
		"exploitability": *exploitability,
		"artifacts":      *artifacts,
		"commit_hash":    *commitHash,
	}
	if *sourcePostID > 0 {
		body["source_post_id"] = *sourcePostID
	}
	resp, err := client.postJSON("/api/repros", body)
	if err != nil {
		fatal("repro create failed: %v", err)
	}
	var repro map[string]any
	if err := readJSON(resp, &repro); err != nil {
		fatal("repro create failed: %v", err)
	}
	fmt.Printf("created repro #%v for finding #%v\n", repro["id"], repro["finding_id"])
}

func cmdRepros(args []string) {
	fs := flag.NewFlagSet("repros", flag.ExitOnError)
	findingID := fs.Int("finding", 0, "filter by finding id")
	limit := fs.Int("limit", 20, "max results")
	fs.Parse(args)

	cfg := mustLoadConfig()
	client := newClient(cfg)
	path := fmt.Sprintf("/api/repros?limit=%d", *limit)
	if *findingID > 0 {
		path += fmt.Sprintf("&finding_id=%d", *findingID)
	}
	resp, err := client.get(path)
	if err != nil {
		fatal("request failed: %v", err)
	}
	var repros []map[string]any
	if err := readJSON(resp, &repros); err != nil {
		fatal("failed: %v", err)
	}
	if len(repros) == 0 {
		fmt.Println("no repros")
		return
	}
	for _, repro := range repros {
		fmt.Printf("#%-4v finding=%-4v exploitability=%-16s %s\n", repro["id"], repro["finding_id"], str(repro["exploitability"]), shortenLine(str(repro["steps"]), 80))
	}
}

func cmdTriage(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: ah triage <finding-id> [--limit N]")
		os.Exit(1)
	}
	findingID := args[0]
	fs := flag.NewFlagSet("triage", flag.ExitOnError)
	limit := fs.Int("limit", 20, "max results")
	fs.Parse(args[1:])
	cfg := mustLoadConfig()
	client := newClient(cfg)
	resp, err := client.get(fmt.Sprintf("/api/findings/%s/triage?limit=%d", findingID, *limit))
	if err != nil {
		fatal("request failed: %v", err)
	}
	var decisions []map[string]any
	if err := readJSON(resp, &decisions); err != nil {
		fatal("failed: %v", err)
	}
	if len(decisions) == 0 {
		fmt.Println("no triage decisions")
		return
	}
	for _, d := range decisions {
		fmt.Printf("#%-4v %-18s %-8s owner=%-16s %s\n", d["id"], str(d["status"]), str(d["severity"]), str(d["owner"]), shortenLine(str(d["next_action"]), 60))
	}
}

func cmdTriageUpdate(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: ah triage-update <finding-id> --status S --severity S [--reasoning ...]")
		os.Exit(1)
	}
	findingID := args[0]
	fs := flag.NewFlagSet("triage-update", flag.ExitOnError)
	status := fs.String("status", "", "status")
	severity := fs.String("severity", "", "severity")
	reasoning := fs.String("reasoning", "", "reasoning")
	owner := fs.String("owner", "", "owner")
	nextAction := fs.String("next-action", "", "next action")
	sourcePostID := fs.Int("source-post", 0, "source post id")
	fs.Parse(args[1:])
	cfg := mustLoadConfig()
	client := newClient(cfg)
	body := map[string]any{
		"status":      *status,
		"severity":    *severity,
		"reasoning":   *reasoning,
		"owner":       *owner,
		"next_action": *nextAction,
	}
	if *sourcePostID > 0 {
		body["source_post_id"] = *sourcePostID
	}
	resp, err := client.postJSON("/api/findings/"+findingID+"/triage", body)
	if err != nil {
		fatal("triage update failed: %v", err)
	}
	var decision map[string]any
	if err := readJSON(resp, &decision); err != nil {
		fatal("triage update failed: %v", err)
	}
	fmt.Printf("recorded triage #%v for finding #%v [%s/%s]\n", decision["id"], decision["finding_id"], str(decision["severity"]), str(decision["status"]))
}

func cmdArtifactUpload(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: ah artifact-upload <path> [--kind TYPE] [--label TEXT] [--finding ID] [--repro ID]")
		os.Exit(1)
	}
	artifactPath := args[0]
	fs := flag.NewFlagSet("artifact-upload", flag.ExitOnError)
	kind := fs.String("kind", "other", "artifact kind")
	label := fs.String("label", "", "artifact label")
	findingID := fs.Int("finding", 0, "finding id")
	reproID := fs.Int("repro", 0, "repro id")
	fs.Parse(args[1:])
	cfg := mustLoadConfig()
	client := newClient(cfg)
	fields := map[string]string{
		"kind":  *kind,
		"label": *label,
	}
	if *findingID > 0 {
		fields["finding_id"] = strconv.Itoa(*findingID)
	}
	if *reproID > 0 {
		fields["repro_id"] = strconv.Itoa(*reproID)
	}
	resp, err := client.postMultipartFile("/api/artifacts", "file", artifactPath, fields)
	if err != nil {
		fatal("artifact upload failed: %v", err)
	}
	var artifact map[string]any
	if err := readJSON(resp, &artifact); err != nil {
		fatal("artifact upload failed: %v", err)
	}
	fmt.Printf("uploaded artifact #%v %s (%s)\n", artifact["id"], str(artifact["filename"]), str(artifact["kind"]))
}

func cmdArtifacts(args []string) {
	fs := flag.NewFlagSet("artifacts", flag.ExitOnError)
	findingID := fs.Int("finding", 0, "filter by finding id")
	reproID := fs.Int("repro", 0, "filter by repro id")
	limit := fs.Int("limit", 20, "max results")
	fs.Parse(args)

	cfg := mustLoadConfig()
	client := newClient(cfg)
	path := fmt.Sprintf("/api/artifacts?limit=%d", *limit)
	if *findingID > 0 {
		path += fmt.Sprintf("&finding_id=%d", *findingID)
	}
	if *reproID > 0 {
		path += fmt.Sprintf("&repro_id=%d", *reproID)
	}
	resp, err := client.get(path)
	if err != nil {
		fatal("request failed: %v", err)
	}
	var artifacts []map[string]any
	if err := readJSON(resp, &artifacts); err != nil {
		fatal("failed: %v", err)
	}
	if len(artifacts) == 0 {
		fmt.Println("no artifacts")
		return
	}
	for _, artifact := range artifacts {
		fmt.Printf("#%-4v %-10s %-20s %8sB %s\n", artifact["id"], str(artifact["kind"]), str(artifact["label"]), str(artifact["size_bytes"]), str(artifact["filename"]))
	}
}

func cmdArtifactDownload(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: ah artifact-download <id> [--out PATH]")
		os.Exit(1)
	}
	id := args[0]
	fs := flag.NewFlagSet("artifact-download", flag.ExitOnError)
	out := fs.String("out", "", "output path")
	fs.Parse(args[1:])
	cfg := mustLoadConfig()
	client := newClient(cfg)

	metaResp, err := client.get("/api/artifacts/" + id)
	if err != nil {
		fatal("artifact request failed: %v", err)
	}
	var artifact map[string]any
	if err := readJSON(metaResp, &artifact); err != nil {
		fatal("artifact request failed: %v", err)
	}
	outPath := *out
	if outPath == "" {
		outPath = str(artifact["filename"])
	}
	resp, err := client.get("/api/artifacts/" + id + "/download")
	if err != nil {
		fatal("download failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		fatal("download failed: %s", string(body))
	}
	file, err := os.Create(outPath)
	if err != nil {
		fatal("create output file: %v", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, resp.Body); err != nil {
		fatal("write output file: %v", err)
	}
	fmt.Printf("downloaded artifact #%s to %s\n", id, outPath)
}

func shortenLine(value string, max int) string {
	value = strings.ReplaceAll(value, "\n", " ")
	if len(value) <= max {
		return value
	}
	return value[:max-3] + "..."
}

func prettyJSON(v any) string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}
