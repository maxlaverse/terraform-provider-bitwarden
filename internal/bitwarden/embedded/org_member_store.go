package embedded

import (
	"fmt"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

type OrgMemberStore map[string][]models.OrgMember

func NewOrgMemberStore() OrgMemberStore {
	return make(map[string][]models.OrgMember)
}

func (m OrgMemberStore) AddMember(orgId string, member models.OrgMember) {
	m[orgId] = append(m[orgId], member)
}

func (m OrgMemberStore) FindMemberByID(orgId, memberId string) (*models.OrgMember, error) {
	if len(memberId) == 0 {
		return nil, fmt.Errorf("BUG: FindMemberByID() called with empty memberId")
	}

	members, ok := m[orgId]
	if ok {
		for _, user := range members {
			if user.ID == memberId {
				return &user, nil
			}
		}
	}

	return nil, fmt.Errorf("no member found with email '%s' in organization '%s' (org exists: %t)", orgId, memberId, ok)
}

func (m OrgMemberStore) FindMemberByEmail(orgId, userEmail string) (*models.OrgMember, error) {
	if len(userEmail) == 0 {
		return nil, fmt.Errorf("BUG: FindMemberByEmail() called with empty userEmail")
	}

	members, ok := m[orgId]
	if ok {
		for _, user := range members {
			if user.Email == userEmail {
				return &user, nil
			}
		}
	}

	return nil, fmt.Errorf("no member found with email '%s' in organization '%s' (org exists: %t)", userEmail, orgId, ok)
}

func (m OrgMemberStore) FindMemberByName(orgId, userName string) (*models.OrgMember, error) {
	if len(userName) == 0 {
		return nil, fmt.Errorf("BUG: FindMemberByName() called with empty userName")
	}

	members, ok := m[orgId]
	if ok {
		for _, user := range members {
			if user.Name == userName {
				return &user, nil
			}
		}
	}

	return nil, fmt.Errorf("no member found with name '%s' in organization '%s' (org exists: %t)", userName, orgId, ok)
}

func (m OrgMemberStore) ForgetOrganization(orgId string) {
	delete(m, orgId)
}

func (m OrgMemberStore) OrganizationInitialized(orgId string) bool {
	_, ok := m[orgId]
	return ok
}

func (m OrgMemberStore) LoadMembers(orgId string, users []webapi.OrganizationUserDetails) {
	m[orgId] = []models.OrgMember{}
	for _, user := range users {
		m[orgId] = append(m[orgId], models.OrgMember{
			ID:             user.Id,
			Email:          user.Email,
			Name:           user.Name,
			OrganizationId: orgId,
			UserId:         user.UserId,
		})
	}
}
