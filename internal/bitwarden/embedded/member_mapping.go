package embedded

import "fmt"

type Member struct {
	Email  string
	Id     string
	UserId string
}

type MemberMapping map[string][]Member

func NewMemberMapping() MemberMapping {
	return make(map[string][]Member)
}

func (m MemberMapping) AddMember(orgId string, member Member) {
	m[orgId] = append(m[orgId], member)
}

func (m MemberMapping) FindMemberByID(orgId, memberId string) (*Member, error) {
	if members, ok := m[orgId]; ok {
		for _, user := range members {
			if user.Id == memberId {
				return &user, nil
			}
		}
	} else {
		return nil, fmt.Errorf("BUG: unknown organization '%s'", orgId)
	}

	return nil, fmt.Errorf("no member found with email '%s' in organization '%s': %+v", orgId, memberId, m)
}

func (m MemberMapping) FindMemberByEmail(orgId, userEmail string) (*Member, error) {
	if members, ok := m[orgId]; ok {
		for _, user := range members {
			if user.Email == userEmail {
				return &user, nil
			}
		}
	} else {
		return nil, fmt.Errorf("BUG: unknown organization '%s'", orgId)
	}

	return nil, fmt.Errorf("no member found with email '%s' in organization '%s': %+v", userEmail, orgId, m)
}

func (m MemberMapping) ForgetOrganization(orgId string) {
	delete(m, orgId)
}

func (m MemberMapping) OrganizationInitialized(orgId string) bool {
	_, ok := m[orgId]
	return ok
}

func (m MemberMapping) ResetOrganization(orgId string) {
	m[orgId] = []Member{}
}
