package coolify

import (
	"errors"
	"net/url"
)

// CreatePublicAppRequest is the body of POST /api/v1/applications/public.
type CreatePublicAppRequest struct {
	ProjectUUID      string `json:"project_uuid"`
	ServerUUID       string `json:"server_uuid"`
	EnvironmentName  string `json:"environment_name"`
	GitRepository    string `json:"git_repository"`
	GitBranch        string `json:"git_branch"`
	BuildPack        string `json:"build_pack"`
	PortsExposes     string `json:"ports_exposes"`
	Name             string `json:"name"`
	Domains          string `json:"domains"`
	LimitsMemory     string `json:"limits_memory"`
	LimitsMemorySwap string `json:"limits_memory_swap"`
	LimitsCPUs       string `json:"limits_cpus"`
}

// CreatePublicApp creates an app from a public git repo and returns its UUID.
func (c *Client) CreatePublicApp(req CreatePublicAppRequest) (string, error) {
	var out struct {
		UUID string `json:"uuid"`
	}
	if err := c.do("POST", "/applications/public", req, &out); err != nil {
		return "", err
	}
	return out.UUID, nil
}

// DeployApp queues a deployment and returns its deployment UUID.
func (c *Client) DeployApp(appUUID string) (string, error) {
	var out struct {
		Deployments []struct {
			DeploymentUUID string `json:"deployment_uuid"`
		} `json:"deployments"`
	}
	path := "/deploy?" + url.Values{"uuid": {appUUID}}.Encode()
	if err := c.do("POST", path, nil, &out); err != nil {
		return "", err
	}
	if len(out.Deployments) == 0 {
		return "", errors.New("coolify returned no deployment")
	}
	return out.Deployments[0].DeploymentUUID, nil
}