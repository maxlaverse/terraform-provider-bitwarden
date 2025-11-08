package embedded

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

// OrgCache manages cached organization data (groups and members)
type OrgCache struct {
	mu     sync.RWMutex
	cache  map[string]*orgCacheEntry
	client webapi.Client
}

type orgCacheEntry struct {
	groups  []models.OrgGroup
	members []models.OrgMember
    groupsLoaded  bool
    membersLoaded bool	
}

func NewOrgCache(client webapi.Client) OrgCache {
	return OrgCache{
		cache:  make(map[string]*orgCacheEntry),
		client: client,
	}
}

// GetGroups returns cached groups for an organization, loading them if needed
func (c *OrgCache) GetGroups(ctx context.Context, orgId string) ([]models.OrgGroup, error) {
    c.mu.RLock()
    entry, exists := c.cache[orgId]
    if exists && entry.groupsLoaded {
        groups := entry.groups
        c.mu.RUnlock()
        return groups, nil
    }
    c.mu.RUnlock()
    return c.loadGroups(ctx, orgId)
}

// GetMembers returns cached members for an organization, loading them if needed
func (c *OrgCache) GetMembers(ctx context.Context, orgId string) ([]models.OrgMember, error) {
    c.mu.RLock()
    entry, exists := c.cache[orgId]
    if exists && entry.membersLoaded {
        members := entry.members
        c.mu.RUnlock()
        return members, nil
    }
    c.mu.RUnlock()
    return c.loadMembers(ctx, orgId)
}

// FindGroupByID finds a group by ID in the specified organization
func (c *OrgCache) FindGroupByID(ctx context.Context, orgId, groupId string) (*models.OrgGroup, error) {
	if len(groupId) == 0 {
		return nil, fmt.Errorf("BUG: FindGroupByID() called with empty groupId")
	}

	groups, err := c.GetGroups(ctx, orgId)
	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		if group.ID == groupId {
			return &group, nil
		}
	}

	return nil, fmt.Errorf("no group found with groupId '%s' in organization '%s'", groupId, orgId)
}

// FindGroupByName finds a group by name in the specified organization
func (c *OrgCache) FindGroupByName(ctx context.Context, orgId, groupName string) (*models.OrgGroup, error) {
	if len(groupName) == 0 {
		return nil, fmt.Errorf("BUG: FindGroupByName() called with empty groupName")
	}

	groups, err := c.GetGroups(ctx, orgId)
	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		if group.Name == groupName {
			return &group, nil
		}
	}

	return nil, fmt.Errorf("no group found with name '%s' in organization '%s'", groupName, orgId)
}

// FindMemberByID finds a member by ID in the specified organization
func (c *OrgCache) FindMemberByID(ctx context.Context, orgId, memberId string) (*models.OrgMember, error) {
	if len(memberId) == 0 {
		return nil, fmt.Errorf("BUG: FindMemberByID() called with empty memberId")
	}

	members, err := c.GetMembers(ctx, orgId)
	if err != nil {
		return nil, err
	}

	for _, member := range members {
		if member.ID == memberId {
			return &member, nil
		}
	}

	return nil, fmt.Errorf("no member found with memberId '%s' in organization '%s'", memberId, orgId)
}

// FindMemberByEmail finds a member by email in the specified organization
func (c *OrgCache) FindMemberByEmail(ctx context.Context, orgId, userEmail string) (*models.OrgMember, error) {
	if len(userEmail) == 0 {
		return nil, fmt.Errorf("BUG: FindMemberByEmail() called with empty userEmail")
	}

	members, err := c.GetMembers(ctx, orgId)
	if err != nil {
		return nil, err
	}

	for _, member := range members {
		if member.Email == userEmail {
			return &member, nil
		}
	}

	return nil, fmt.Errorf("no member found with email '%s' in organization '%s'", userEmail, orgId)
}

// InvalidateOrganization marks the cache for an organization as stale
func (c *OrgCache) InvalidateOrganization(ctx context.Context, orgId string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tflog.Trace(ctx, "Invalidating organization cache", map[string]interface{}{"org_id": orgId})
	delete(c.cache, orgId)
}

// InvalidateAll clears all cached data
func (c *OrgCache) InvalidateAll(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tflog.Trace(ctx, "Invalidating all organization cache")
	c.cache = make(map[string]*orgCacheEntry)
}

// loadGroups loads groups from the API and caches them
func (c *OrgCache) loadGroups(ctx context.Context, orgId string) ([]models.OrgGroup, error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    tflog.Trace(ctx, "Loading groups for organization", map[string]interface{}{"org_id": orgId})

    orgGroups, err := c.client.GetOrganizationGroups(ctx, orgId)
    if err != nil {
        return nil, fmt.Errorf("error getting organization groups: %w", err)
    }

    groups := make([]models.OrgGroup, len(orgGroups))
    for i, g := range orgGroups {
        groups[i] = models.OrgGroup{ ID: g.Id, Name: g.Name, OrganizationID: orgId }
    }

    if entry, ok := c.cache[orgId]; ok {
        entry.groups = groups
        entry.groupsLoaded = true
        // keep entry.members / entry.membersLoaded as-is
    } else {
        c.cache[orgId] = &orgCacheEntry{
            groups:       groups,
            groupsLoaded: true,
        }
    }
    return groups, nil
}

// loadMembers loads members from the API and caches them
func (c *OrgCache) loadMembers(ctx context.Context, orgId string) ([]models.OrgMember, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tflog.Trace(ctx, "Loading members for organization", map[string]interface{}{"org_id": orgId})

	orgUsers, err := c.client.GetOrganizationUsers(ctx, orgId)
	if err != nil {
		return nil, fmt.Errorf("error getting organization users: %w", err)
	}

	members := make([]models.OrgMember, len(orgUsers))
	for i, user := range orgUsers {
		members[i] = models.OrgMember{
			ID:             user.Id,
			Email:          user.Email,
			Name:           user.Name,
			OrganizationId: orgId,
			UserId:         user.UserId,
		}
	}

    if entry, ok := c.cache[orgId]; ok {
        entry.members = members
        entry.membersLoaded = true
        // keep entry.groups / entry.groupsLoaded as-is
    } else {
        c.cache[orgId] = &orgCacheEntry{
            members:       members,
            membersLoaded: true,
        }
    }

	return members, nil
}
