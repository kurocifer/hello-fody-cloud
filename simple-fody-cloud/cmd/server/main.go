// A one-page FodyCloud: paste a public repo, get a deployed app.
// No accounts, no database. It exists to prove the Coolify pipeline works.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"fodycloud/internal/coolify"
)

const (
	serverUUID  = "h9sptsdidaojmfgqdgqj5wko"
	projectUUID = "pbadlskzxholftkwsptmmfhc"
)

var nameRe = regexp.MustCompile(`^[a-z0-9-]{1,30}$`)

type page struct {
	Name, Repo, Branch, Port string
	Error, URL               string
}

var tpl = template.Must(template.New("p").Parse(`<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>FodyCloud</title>
<style>
 :root{--bg:#f6f5f1;--card:#fff;--ink:#14161a;--muted:#6a6f77;--line:#dcdad2;
       --accent:#1f6f5c;--accent-ink:#fff;--bad:#b3261e}
 @media (prefers-color-scheme:dark){
  :root{--bg:#101213;--card:#181b1d;--ink:#ecebe6;--muted:#9aa0a6;--line:#2a2e31;
        --accent:#3fb391;--accent-ink:#06130f;--bad:#ff8a80}}
 *{box-sizing:border-box}
 body{margin:0;background:var(--bg);color:var(--ink);
      font:16px/1.5 system-ui,-apple-system,"Segoe UI",sans-serif;
      max-width:34rem;margin:0 auto;padding:3rem 1.25rem 4rem}
 h1{font-size:1.9rem;margin:0;letter-spacing:.01em}
 .tag{color:var(--muted);margin:.2rem 0 1.75rem}
 .card{background:var(--card);border:1px solid var(--line);border-radius:12px;padding:1.4rem}
 label{display:block;margin:1rem 0 .3rem;font-size:.85rem;color:var(--muted)}
 label:first-child{margin-top:0}
 input{width:100%;padding:.65rem .75rem;font:inherit;color:var(--ink);
       background:var(--bg);border:1px solid var(--line);border-radius:8px}
 input:focus{outline:2px solid var(--accent);outline-offset:1px;border-color:transparent}
 .row{display:grid;grid-template-columns:1fr 1fr;gap:.9rem}
 @media (max-width:28rem){.row{grid-template-columns:1fr}}
 button{margin-top:1.4rem;width:100%;padding:.7rem 1rem;font:inherit;font-weight:600;
        color:var(--accent-ink);background:var(--accent);border:0;border-radius:8px;cursor:pointer}
 button:hover{filter:brightness(1.08)}
 button:focus-visible{outline:2px solid var(--accent);outline-offset:3px}
 .err{margin-top:1.25rem;padding:.75rem 1rem;border:1px solid var(--bad);
      border-radius:8px;color:var(--bad);word-break:break-word}
 .ok{margin-top:1.25rem;padding:.9rem 1rem;border:1px solid var(--accent);border-radius:8px}
 .ok a{color:var(--accent);font-weight:600;word-break:break-all}
</style></head><body>
<h1>FodyCloud</h1>
<p class="tag">Your code. Your deployment. Your cloud.</p>
<form class="card" method="post" action="/deploy">
 <label for="name">Project name (lowercase letters, digits, dashes)</label>
 <input id="name" name="name" value="{{.Name}}" placeholder="my-portfolio" required>
 <label for="repo">Public repository URL</label>
 <input id="repo" name="repo" value="{{.Repo}}" placeholder="https://github.com/you/project" required>
 <div class="row">
  <div><label for="branch">Branch</label>
   <input id="branch" name="branch" value="{{.Branch}}" required></div>
  <div><label for="port">App port</label>
   <input id="port" name="port" value="{{.Port}}" inputmode="numeric" required></div>
 </div>
 <button type="submit">Deploy</button>
</form>
{{if .Error}}<p class="err">{{.Error}}</p>{{end}}
{{if .URL}}<p class="ok">Deploying. It can take a few minutes the first time.<br>
Your site: <a href="{{.URL}}" target="_blank" rel="noopener">{{.URL}}</a></p>{{end}}
</body></html>`))

type server struct {
	cf    *coolify.Client
	vmIP  string
	mem   string
	cpus  string
}

func (s *server) render(w http.ResponseWriter, p page) {
	if err := tpl.Execute(w, p); err != nil {
		log.Println("template:", err)
	}
}

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	s.render(w, page{Branch: "main", Port: "3000"})
}

func (s *server) deploy(w http.ResponseWriter, r *http.Request) {
	p := page{
		Name:   strings.TrimSpace(r.FormValue("name")),
		Repo:   strings.TrimSpace(r.FormValue("repo")),
		Branch: strings.TrimSpace(r.FormValue("branch")),
		Port:   strings.TrimSpace(r.FormValue("port")),
	}
	if !nameRe.MatchString(p.Name) {
		p.Error = "Name: 1-30 characters, lowercase letters, digits and dashes only."
		s.render(w, p)
		return
	}
	if !strings.HasPrefix(p.Repo, "https://") {
		p.Error = "The repository URL must start with https://"
		s.render(w, p)
		return
	}

	// Random suffix so deploying the same name twice doesn't clash on the domain.
	b := make([]byte, 2)
	rand.Read(b)
	name := p.Name + "-" + hex.EncodeToString(b)
	domain := fmt.Sprintf("http://%s.%s.sslip.io", name, s.vmIP)

	uuid, err := s.cf.CreatePublicApp(coolify.CreatePublicAppRequest{
		ProjectUUID:      projectUUID,
		ServerUUID:       serverUUID,
		EnvironmentName:  "production",
		GitRepository:    p.Repo,
		GitBranch:        p.Branch,
		BuildPack:        "railpack",
		PortsExposes:     p.Port,
		Name:             name,
		Domains:          domain,
		LimitsMemory:     s.mem,
		LimitsMemorySwap: s.mem, // swap = memory means no extra swap
		LimitsCPUs:       s.cpus,
	})
	if err != nil {
		p.Error = "Create failed: " + err.Error()
		s.render(w, p)
		return
	}
	if _, err := s.cf.DeployApp(uuid); err != nil {
		p.Error = "App created but deploy failed: " + err.Error()
		s.render(w, p)
		return
	}
	p.URL = domain
	s.render(w, p)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	baseURL, token := os.Getenv("COOLIFY_URL"), os.Getenv("COOLIFY_TOKEN")
	if baseURL == "" || token == "" {
		log.Fatal("set COOLIFY_URL and COOLIFY_TOKEN first")
	}
	s := &server{
		cf:   coolify.New(baseURL, token),
		vmIP: env("VM_IP", "192.168.122.236"),
		mem:  env("APP_MEMORY", "128m"),
		cpus: env("APP_CPUS", "0.5"),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.index)
	mux.HandleFunc("POST /deploy", s.deploy)

	srv := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	log.Println("listening on", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}