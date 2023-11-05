package client

import (
	"fmt"
	"github.com/dghubble/sling"
)

type ProjectService struct {
	sling *sling.Sling
}

type ProjectRequest struct {
	Project any `json:"project"`
}

func (c *Client) Projects() *ProjectService {
	return &ProjectService{sling: c.sling}
}

type ProjectUpdatePayload struct {
	CaCertificate string `json:"ca_certificate,omitempty"`
	Status        string `json:"status,omitempty"`
}
type ProjectUpdateRequest struct {
	Project *ProjectUpdatePayload `json:"project"`
}

type ProjectUpdateResponse struct {
}

func (p *ProjectService) Update(projectId int, data *ProjectUpdateRequest) (*ProjectUpdateResponse, error) {
	fmt.Println(data.Project)
	path := fmt.Sprintf("/api/v1/projects/%d", projectId)
	res := new(ProjectUpdateResponse)
	_, err := p.sling.Patch(path).BodyJSON(data).ReceiveSuccess(res)
	return res, err
}
