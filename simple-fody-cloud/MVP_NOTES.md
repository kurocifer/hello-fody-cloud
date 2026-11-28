**FodyCloud MVP experiment**  
A small test of how a "free Vercel for students" works under the hood, built with Coolify. Below is what was done, what went wrong, and what the team should know before choosing an engine.  
Short version: give it a public Git repo URL, and a few minutes later there's a live container with memory and CPU limits. A small Go server does this through Coolify's API, so a student never sees Coolify at all.  
It was a throwaway test on one VM, so treat the results with that in mind.  
**Setup**  
- Fedora host, Ubuntu Server 26 VM in GNOME Boxes (about 3.3 GB RAM)  
- Coolify 4.7.6, Traefik v3.7, Docker 29.10.1, Railpack v0.43.0  
- Backend in Go, one page, no database  
**What happens on a deploy**  
The Go server calls Coolify, and Coolify does the rest:  
1. The job goes on a queue (Redis) and gets recorded (Postgres).  
2. The repo is cloned. Railpack looks at the files and decides how to build it.  
3. BuildKit builds a Docker image, tagged with the git commit.  
4. Coolify starts a container with the memory and CPU limits, plus labels for Traefik.  
5. Traefik sees the labels and routes the hostname to the container.  
Routing is just hostnames. If the name in the browser doesn't match the label on the container, Traefik shows its 503 page, even though the app is running fine.  
**Reproducing it**  
Install Coolify on an Ubuntu VM using the command from its docs, then deploy a public repo from the dashboard to see it work by hand first. For the API:  
1. Turn on API access and create a token (Keys & Tokens) with read, write and deploy.  
2. Look up the IDs needed:  
```  
export COOLIFY_URL="http://<vm-ip>:8000"  
 export COOLIFY_TOKEN="<token>"  
 curl -s -H "Authorization: Bearer $COOLIFY_TOKEN" "$COOLIFY_URL/api/v1/servers"  
 curl -s -H "Authorization: Bearer $COOLIFY_TOKEN" "$COOLIFY_URL/api/v1/projects"  
   
```  
1. Run the Go server and open http://localhost:8080:  
```  
cd fodycloud  
 go run ./cmd/server  
   
```  
It only makes two calls. `POST /api/v1/applications/public` creates the app (project and server UUIDs, repo, branch, `build_pack` set to `railpack`, `ports_exposes`, `domains`, and the limits `limits_memory`, `limits_memory_swap` and `limits_cpus`). Then `POST /api/v1/deploy?uuid=<app-uuid>` starts the build. The swap limit is set equal to the memory limit so a container can't get around its cap by using swap.  
**What went wrong**  
**The site wasn't reachable after deploying.** Coolify generated the sslip.io domain from the public IP instead of the VM's local one. The fix was to set the domain by hand with the VM's IP and redeploy. When using the API, always send `domains` explicitly.  
**Traefik 503.** Either the domain didn't match or the exposed port was wrong. Routes come from labels that are only set when the container starts, so changing the domain does nothing until a redeploy.  
**Everything timed out in the browser, but curl worked.** A VPN extension was sending traffic through its own servers, which can't reach a private IP. If curl works and the browser doesn't, the browser is the place to look.  
**Build failed on `apt-get update`.** The host went to sleep overnight and the VM's clock stayed behind. Debian's package signatures were dated after the VM's "now", so apt rejected them. Turning time sync back on (`timedatectl set-ntp true`) and redeploying fixed it. A real VPS wouldn't sleep, but the platform still depends on a correct clock.  
**Findings**  
The whole pipeline works, from URL to a live site. The limits set through Coolify do reach the container, and a container that goes over its memory limit gets killed and restarted without affecting the others (exit code 137, restart count went up, and the kernel logged a memory cgroup kill). Coolify's API was enough to create and deploy from Go in a few dozen lines. Idle, Coolify itself uses roughly 550 MiB of RAM.  
First builds are slow on a weak connection because the big image layers download once. After that they're cached.  
What Coolify doesn't cover for FodyCloud: student accounts and keeping users separate, plans and quotas, billing, repo validation and abuse detection, and a dashboard students would actually use. It assumes one admin. Students shouldn't ever get a Coolify login or Docker socket access, since that's effectively root on the server.  
**Questions for the team**  
- When a build fails because of the platform (clock, network, registry), how do students get told it isn't their code?  
- Who watches time sync and server health in production?  
- Stay on Coolify behind a custom backend, or build directly on Docker and Traefik later?  
**Other options considered**  
  
|  |  |  |  
| - | - | - |  
| **Option** |  **What it is** |  **Note** |  
| Dokploy |  Self-hosted PaaS with a web UI |  Closest alternative to Coolify, with Docker Swarm clustering |  
| CapRover |  Self-hosted PaaS with a web UI |  Mature and stable, but the UI is dated |  
| Dokku |  Self-hosted PaaS, CLI only |  Lightest option, but works through SSH commands with no built-in REST API, so driving it from a backend is harder |  
| Kamal |  Deploy tool, CLI only |  Not a dashboard |  
| CloudStack, OpenStack, Proxmox |  VM management |  Work below the app layer, so they only matter if FodyCloud ever runs its own hardware |  
  
   
**Not tested**  
Everything ran on one VM on a laptop with a slow connection and one user. Malicious builds, many users, custom domains, real HTTPS with Let's Encrypt, and multiple servers were not tried.  
