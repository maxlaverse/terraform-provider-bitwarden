package embedded

import "fmt"

type OrganizationMember struct {
	Email  string
	Id     string
	UserId string
}

type MemberMapping map[string][]OrganizationMember

func NewMemberMapping() MemberMapping {
	return make(map[string][]OrganizationMember)
}

func (m MemberMapping) AddMember(orgId string, member OrganizationMember) {
	m[orgId] = append(m[orgId], member)
}

func (m MemberMapping) FindMemberByID(orgId, memberId string) (*OrganizationMember, error) {
	if len(memberId) == 0 {
		return nil, fmt.Errorf("BUG: FindMemberByID() called with empty memberId")
	}

	members, ok := m[orgId]
	if ok {
		for _, user := range members {
			if user.Id == memberId {
				return &user, nil
			}
		}
	}

	return nil, fmt.Errorf("no member found with email '%s' in organization '%s' (org exists: %t)", orgId, memberId, ok)
}

func (m MemberMapping) FindMemberByEmail(orgId, userEmail string) (*OrganizationMember, error) {
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

func (m MemberMapping) ForgetOrganization(orgId string) {
	delete(m, orgId)
}

func (m MemberMapping) OrganizationInitialized(orgId string) bool {
	_, ok := m[orgId]
	return ok
}

func (m MemberMapping) ResetOrganization(orgId string) {
	m[orgId] = []OrganizationMember{}
}
