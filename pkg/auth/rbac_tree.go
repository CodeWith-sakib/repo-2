package auth

import (
	"sync"
)

type RoleNode struct {
	Name        string
	Permissions map[Permission]bool
	Parents     []*RoleNode
}

type RoleHierarchyResolver struct {
	mu    sync.RWMutex
	roles map[string]*RoleNode
}

func NewRoleHierarchyResolver() *RoleHierarchyResolver {
	r := &RoleHierarchyResolver{
		roles: make(map[string]*RoleNode),
	}
	r.initDefaults()
	return r
}

func (r *RoleHierarchyResolver) initDefaults() {
	viewer := &RoleNode{
		Name: string(RoleViewer),
		Permissions: map[Permission]bool{
			PermWorkflowRead: true,
			PermRunRead:      true,
			PermEventRead:    true,
		},
	}
	operator := &RoleNode{
		Name: string(RoleOperator),
		Permissions: map[Permission]bool{
			PermRunCreate: true,
			PermRunCancel: true,
		},
		Parents: []*RoleNode{viewer},
	}
	admin := &RoleNode{
		Name: string(RoleAdmin),
		Permissions: map[Permission]bool{
			PermWorkflowWrite: true,
			PermWorkflowAdmin: true,
			PermSystemAdmin:   true,
		},
		Parents: []*RoleNode{operator},
	}

	r.roles[string(RoleViewer)] = viewer
	r.roles[string(RoleOperator)] = operator
	r.roles[string(RoleAdmin)] = admin
}

// EffectivePermissions recursively resolves all inherited permissions for a given role.
func (r *RoleHierarchyResolver) EffectivePermissions(roleName string) map[Permission]bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	effective := make(map[Permission]bool)
	node, exists := r.roles[roleName]
	if !exists {
		return effective
	}

	var traverse func(curr *RoleNode)
	traverse = func(curr *RoleNode) {
		for perm := range curr.Permissions {
			effective[perm] = true
		}
		for _, parent := range curr.Parents {
			traverse(parent)
		}
	}

	traverse(node)
	return effective
}

// HasPermission checks if the role possesses the target permission directly or via inheritance.
func (r *RoleHierarchyResolver) HasPermission(roleName string, perm Permission) bool {
	perms := r.EffectivePermissions(roleName)
	return perms[perm]
}
