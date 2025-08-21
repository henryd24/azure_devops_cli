package models

type RoleAssignmentPayload struct {
	Value   []RoleAssignment `json:"value"`
	Inherit bool             `json:"inherit"`
	Token   string           `json:"token"`
	Merge   bool             `json:"merge"`
}

type RoleAssignment struct {
	Descriptor   string   `json:"descriptor"`
	Allow        int      `json:"allow"`
	Deny         int      `json:"deny"`
	ExtendedInfo struct{} `json:"extendedInfo"`
}

type SecurityRoleAssignment struct {
	RoleName string `json:"roleName"`
	UserID   string `json:"userId"` // El ID (GUID) del usuario o grupo
}
