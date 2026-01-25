package auth

import (
	"sync"
)

// RolePermissionMatrix evaluates permission grants with hierarchical inheritance.
type RolePermissionMatrix struct {
	mu           sync.RWMutex
	roleGrants   map[string]map[string]bool // role -> set of permissions
	roleParents  map[string][]string        // role -> list of parent roles
}

// NewRolePermissionMatrix creates a permission evaluation matrix.
func NewRolePermissionMatrix() *RolePermissionMatrix {
	return &RolePermissionMatrix{
		roleGrants:  make(map[string]map[string]bool),
		roleParents: make(map[string][]string),
	}
}

// GrantPermission attaches an action permission to a role.
func (m *RolePermissionMatrix) GrantPermission(role, permission string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.roleGrants[role]; !exists {
		m.roleGrants[role] = make(map[string]bool)
	}
	m.roleGrants[role][permission] = true
}

// SetRoleInheritance sets parent roles from which a child role inherits privileges.
func (m *RolePermissionMatrix) SetRoleInheritance(childRole string, parentRoles ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.roleParents[childRole] = append(m.roleParents[childRole], parentRoles...)
}

// HasPermission recursively checks if a role has the target permission.
func (m *RolePermissionMatrix) HasPermission(role, permission string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	visited := make(map[string]bool)
	return m.checkPermissionRecursive(role, permission, visited)
}

func (m *RolePermissionMatrix) checkPermissionRecursive(role, permission string, visited map[string]bool) bool {
	if visited[role] {
		return false // cycle protection
	}
	visited[role] = true

	// Check direct grants
	if grants, ok := m.roleGrants[role]; ok {
		if grants[permission] || grants["*"] {
			return true
		}
	}

	// Check inherited parent grants
	if parents, ok := m.roleParents[role]; ok {
		for _, parent := range parents {
			if m.checkPermissionRecursive(parent, permission, visited) {
				return true
			}
		}
	}

	return false
}
