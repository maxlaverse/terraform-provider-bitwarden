package embedded

import (
	"fmt"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

type OrgGroupStore map[string][]models.OrgGroup

func NewOrgGroupStore() OrgGroupStore {
	return make(map[string][]models.OrgGroup)
}

func (m OrgGroupStore) AddGroup(orgId string, group models.OrgGroup) {
	m[orgId] = append(m[orgId], group)
}

func (m OrgGroupStore) FindGroupByID(orgId, groupId string) (*models.OrgGroup, error) {
	if len(groupId) == 0 {
		return nil, fmt.Errorf("BUG: FindGroupByID() called with empty groupId")
	}

	groups, ok := m[orgId]
	if ok {
		for _, group := range groups {
			if group.ID == groupId {
				return &group, nil
			}
		}
	}

	return nil, fmt.Errorf("no group found with groupId '%s' in organization '%s' (org exists: %t)", groupId, orgId, ok)
}

func (m OrgGroupStore) FindGroupByName(orgId, groupName string) (*models.OrgGroup, error) {
	if len(groupName) == 0 {
		return nil, fmt.Errorf("BUG: FindGroupByName() called with empty groupName")
	}

	groups, ok := m[orgId]
	if ok {
		for _, group := range groups {
			if group.Name == groupName {
				return &group, nil
			}
		}
	}

	return nil, fmt.Errorf("no group found with name '%s' in organization '%s' (org exists: %t)", groupName, orgId, ok)
}

func (m OrgGroupStore) ForgetOrganization(orgId string) {
	delete(m, orgId)
}

func (m OrgGroupStore) OrganizationInitialized(orgId string) bool {
	_, ok := m[orgId]
	return ok
}

func (m OrgGroupStore) LoadGroups(orgId string, groups []webapi.OrganizationGroupDetails) {
	m[orgId] = []models.OrgGroup{}
	for _, group := range groups {
		m[orgId] = append(m[orgId], models.OrgGroup{
			ID:             group.Id,
			Name:           group.Name,
			OrganizationID: orgId,
		})
	}
}
