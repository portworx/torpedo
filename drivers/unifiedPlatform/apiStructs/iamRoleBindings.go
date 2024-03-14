package apiStructs

// PDSCreateIAM Struct for IAM roleBindings
type PDSIam struct {
	Create CreateIAM
}

type CreateIAM struct {
	V1IAM V1IAM
}
type V1RoleBinding struct {
	RoleName    *string
	ResourceIds []string
}
type V1AccessPolicy struct {
	GlobalScope []string        `json:"globalScope,omitempty"`
	Account     []string        `json:"account,omitempty"`
	Tenant      []V1RoleBinding `json:"tenant,omitempty"`
	Project     []V1RoleBinding `json:"project,omitempty"`
	Namespace   []V1RoleBinding
}

type V1IAM struct {
	Meta   V1Meta    `json:"meta,omitempty"`
	Config V1Config3 `json:"config,omitempty"`
}
