package models

import "time"

type Secret struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organizationId"`
	ProjectID      string     `json:"projectId"`
	Key            string     `json:"key"`
	Value          string     `json:"value"`
	Note           string     `json:"note"`
	CreationDate   time.Time  `json:"creationDate"`
	RevisionDate   time.Time  `json:"revisionDate"`
	Object         ObjectType `json:"object"`
}

type Project struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	Name           string    `json:"name"`
	CreationDate   time.Time `json:"creationDate"`
	RevisionDate   time.Time `json:"revisionDate"`
}
