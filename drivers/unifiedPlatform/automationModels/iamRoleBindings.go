package automationModels

// PDSCreateIAM Struct for IAM roleBindings
type IAMRequest struct {
	Create CreateIAM `copier:"must,nopanic"`
	Update UpdateIAM `copier:"must,nopanic"`
	Get    GetIAM    `copier:"must,nopanic"`
	Delete DeleteIAM `copier:"must,nopanic"`
	List   ListIAM   `copier:"must,nopanic"`
	Grant  GrantIAM  `copier:"must,nopanic"`
	Revoke RevokeIAM `copier:"must,nopanic"`
}

type IAMResponse struct {
	Create V1IAM              `copier:"must,nopanic"`
	Update V1IAM              `copier:"must,nopanic"`
	Get    V1IAM              `copier:"must,nopanic"`
	List   ListIAM            `copier:"must,nopanic"`
	Grant  V1GrantIAMResponse `copier:"must,nopanic"`
	Revoke V1GrantIAMResponse `copier:"must,nopanic"`
}

type ListResponse struct {
	Iam        []V1IAM                        `copier:"must,nopanic"`
	Pagination *V1PageBasedPaginationResponse `copier:"must,nopanic"`
}

type GrantIAM struct {
	IamConfigActorId       string                  `copier:"must,nopanic"`
	IAMServiceGrantIAMBody *IAMServiceGrantIAMBody `copier:"must,nopanic"`
}

type IAMServiceGrantIAMBody struct {
	AccountId *string `copier:"must,nopanic"`
	TenantId  *string `copier:"must,nopanic"`
	ProjectId *string `copier:"must,nopanic"`
	Iam       *V1IAM  `copier:"must,nopanic"`
}

type CreateIAM struct {
	V1IAM V1IAM `copier:"must,nopanic"`
}

type UpdateIAM struct {
	IamMetaUid     string `copier:"must,nopanic"`
	IAMToBeUpdated *V1IAM `copier:"must,nopanic"`
}

type GetIAM struct {
	ActorId string `copier:"must,nopanic"`
}

type DeleteIAM struct {
	ActorId string `copier:"must,nopanic"`
}

type ListIAM struct {
	ActorId              *string `copier:"must,nopanic"`
	AccountId            *string `copier:"must,nopanic"`
	TenantId             *string `copier:"must,nopanic"`
	ProjectId            *string `copier:"must,nopanic"`
	SortSortBy           *string `copier:"must,nopanic"`
	SortSortOrder        *string `copier:"must,nopanic"`
	PaginationPageNumber *string `copier:"must,nopanic"`
	PaginationPageSize   *string `copier:"must,nopanic"`
}

type V1RoleBinding struct {
	RoleName    string   `copier:"must,nopanic"`
	ResourceIds []string `copier:"must,nopanic"`
}
type V1AccessPolicy struct {
	GlobalScope []string        `copier:"must,nopanic"`
	Account     []string        `copier:"must,nopanic"`
	Tenant      []V1RoleBinding `copier:"must,nopanic"`
	Project     []V1RoleBinding `copier:"must,nopanic"`
	Namespace   []V1RoleBinding `copier:"must,nopanic"`
}

type V1IAM struct {
	Meta   V1Meta    `copier:"must,nopanic"`
	Config V1Config3 `copier:"must,nopanic"`
}

type V1GrantIAMResponse struct {
	Message *string `copier:"must,nopanic"`
}

type RevokeIAM struct {
	IamConfigActorId        string                   `copier:"must,nopanic"`
	IAMServiceRevokeIAMBody *IAMServiceRevokeIAMBody `copier:"must,nopanic"`
}

type IAMServiceRevokeIAMBody struct {
	AccountId *string                 `copier:"must,nopanic"`
	TenantId  *string                 `copier:"must,nopanic"`
	ProjectId *string                 `copier:"must,nopanic"`
	Iam       *IAMServiceGrantIAMBody `copier:"must,nopanic"`
}
