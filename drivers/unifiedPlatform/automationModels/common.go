package automationModels

import (
	"time"
)

type PaginationRequest struct {
	PageNumber int
	PageSize   int
}

// ProtobufAny4 `Any` contains an arbitrary serialized protocol buffer message along with a URL that describes the type of the serialized message.  Protobuf library provides support to pack/unpack Any values in the form of utility functions or additional generated methods of the Any type.  Example 1: Pack and unpack a message in C++.      Foo foo = ...;     Any any;     any.PackFrom(foo);     ...     if (any.UnpackTo(&foo)) {       ...     }  Example 2: Pack and unpack a message in Java.      Foo foo = ...;     Any any = Any.pack(foo);     ...     if (any.is(Foo.class)) {       foo = any.unpack(Foo.class);     }     // or ...     if (any.isSameTypeAs(Foo.getDefaultInstance())) {       foo = any.unpack(Foo.getDefaultInstance());     }   Example 3: Pack and unpack a message in Python.      foo = Foo(...)     any = Any()     any.Pack(foo)     ...     if any.Is(Foo.DESCRIPTOR):       any.Unpack(foo)       ...   Example 4: Pack and unpack a message in Go       foo := &pb.Foo{...}      any, err := anypb.New(foo)      if err != nil {        ...      }      ...      foo := &pb.Foo{}      if err := any.UnmarshalTo(foo); err != nil {        ...      }  The pack methods provided by protobuf library will by default use 'type.googleapis.com/full.type.name' as the type URL and the unpack methods only use the fully qualified type name after the last '/' in the type URL, for example \"foo.bar.com/x/y.z\" will yield type name \"y.z\".  JSON ==== The JSON representation of an `Any` value uses the regular representation of the deserialized, embedded message, with an additional field `@type` which contains the type URL. Example:      package google.profile;     message Person {       string first_name = 1;       string last_name = 2;     }      {       \"@type\": \"type.googleapis.com/google.profile.Person\",       \"firstName\": <string>,       \"lastName\": <string>     }  If the embedded message type is well-known and has a custom JSON representation, that representation will be embedded adding a field `value` which holds the custom JSON in addition to the `@type` field. Example (for message [google.protobuf.Duration][]):      {       \"@type\": \"type.googleapis.com/google.protobuf.Duration\",       \"value\": \"1.212s\"     }
type ProtobufAny4 struct {
	// A URL/resource name that uniquely identifies the type of the serialized protocol buffer message. This string must contain at least one \"/\" character. The last segment of the URL's path must represent the fully qualified name of the type (as in `path/google.protobuf.Duration`). The name should be in a canonical form (e.g., leading \".\" is not accepted).  In practice, teams usually precompile into the binary all types that they expect it to use in the context of Any. However, for URLs which use the scheme `http`, `https`, or no scheme, one can optionally set up a type server that maps type URLs to message definitions as follows:  * If no scheme is provided, `https` is assumed. * An HTTP GET on the URL must yield a [google.protobuf.Type][]   value in binary format, or produce an error. * Applications are allowed to cache lookup results based on the   URL, or have them precompiled into a binary to avoid any   lookup. Therefore, binary compatibility needs to be preserved   on changes to types. (Use versioned type names to manage   breaking changes.)  Note: this functionality is not currently available in the official protobuf release, and it is not used for type URLs beginning with type.googleapis.com. As of May 2023, there are no widely used type server implementations and no plans to implement one.  Schemes other than `http`, `https` (or the empty scheme) might be used with implementation specific semantics.
	Type                 *string `copier:"must,nopanic"`
	AdditionalProperties map[string]interface{}
}

type Meta struct {
	Uid             *string            `copier:"must"`
	Name            *string            `copier:"must"`
	Description     *string            `copier:"must,nopanic"`
	ResourceVersion *string            `copier:"must,nopanic"`
	CreateTime      *time.Time         `copier:"must,nopanic"`
	UpdateTime      *time.Time         `copier:"must,nopanic"`
	Labels          *map[string]string `copier:"must,nopanic"`
	Annotations     *map[string]string `copier:"must,nopanic"`
}

type Config struct {
	UserEmail       *string `copier:"must,nopanic"`
	DnsName         *string `copier:"must,nopanic"`
	DisplayName     *string `copier:"must,nopanic"`
	References      *References
	JobHistoryLimit *int32  `copier:"must,nopanic"`
	Suspend         *bool   `copier:"must,nopanic"`
	BackupType      *string `copier:"must,nopanic"`
	BackupLevel     *string `copier:"must,nopanic"`
	ReclaimPolicy   *string `copier:"must,nopanic"`
}

type V1Info struct {
	References *V1Reference
	// Image registry where the image is stored.
	Registry *string
	// Image registry namespace where the image is stored.
	Namespace *string
	// Tag associated with the image.
	Tag *string
	// Build version of the image.
	Build *string
	// Flag indicating if TLS is supported for a data service using this image.
	TlsSupport *bool
	// Capabilities associated with this image.
	Capabilities *map[string]string
	// Additional images associated with this data service image.
	AdditionalImages *map[string]string
}

type V1Config2 struct {
	ClientId     *string `json:"clientId,omitempty"`
	ClientSecret *string `json:"clientSecret,omitempty"`
	Disabled     *bool   `json:"disabled,omitempty"`
}

// V1Config3 USED from creating IAM Roles
type V1Config3 struct {
	ActorId      *string `json:"actorId,omitempty"`
	ActorType    *string `json:"actorType,omitempty"`
	AccessPolicy *V1AccessPolicy
}

type V1Config struct {
	Kind            *string                `copier:"must,nopanic"`
	SemanticVersion *string                `copier:"must,nopanic"`
	RevisionUid     *string                `copier:"must,nopanic"`
	TemplateValues  map[string]interface{} `copier:"must,nopanic"`
}

type Status struct {
	Phase              string      `copier:"must,nopanic"`
	StartedAt          time.Time   `copier:"must,nopanic"`
	CompletedAt        time.Time   `copier:"must,nopanic"`
	ErrorCode          V1ErrorCode `copier:"must,nopanic"`
	ErrorMessage       string      `copier:"must,nopanic"`
	CustomResourceName string      `copier:"must,nopanic"`
}

type StatusPhase string

type V1Meta struct {
	Uid             *string            `copier:"must,nopanic"`
	Name            *string            `copier:"must,nopanic"`
	Description     *string            `copier:"must,nopanic"`
	ResourceVersion *string            `copier:"must,nopanic"`
	CreateTime      *time.Time         `copier:"must,nopanic"`
	UpdateTime      *time.Time         `copier:"must,nopanic"`
	Labels          *map[string]string `copier:"must,nopanic"`
	Annotations     *map[string]string `copier:"must,nopanic"`
	ParentReference *V1Reference       `copier:"must,nopanic"`
}

type V1ErrorCode string

type V1Phase string

type V1Reference struct {
	Type    *string `copier:"must,nopanic"`
	Version *string `copier:"must,nopanic"`
	Uid     *string `copier:"must,nopanic"`
}

type V1DeploymentMetaData struct {
	Name                 *string `copier:"must,nopanic"`
	CustomResourceName   *string `copier:"must,nopanic"`
	DeploymentTargetName *string `copier:"must,nopanic"`
	NamespaceName        *string `copier:"must,nopanic"`
	TlsEnabled           *bool   `copier:"must,nopanic"`
}

type PageBasedPaginationRequest struct {
	PageNumber int64 `copier:"must,nopanic"`
	PageSize   int64 `copier:"must,nopanic"`
}

type Sort struct {
	SortBy    SortBy_Field    `copier:"must,nopanic"`
	SortOrder SortOrder_Value `copier:"must,nopanic"`
}

type SortBy_Field int32

type SortOrder_Value int32

type SourceReferences struct {
	// UID of the deployment which was backed up.
	DataServiceDeploymentId string `copier:"must,nopanic"`
	// UID of the backup.
	BackupId string `copier:"must,nopanic"`
	// UID of the backup location.
	BackupLocationId string `copier:"must,nopanic"`
	// UID of the cloud snapshot of the backup volume used for restore.
	CloudsnapId string `copier:"must,nopanic"`
}

type DestinationReferences struct {
	// UID of the target cluster where restore will be created.
	TargetClusterId string `copier:"must,nopanic"`
	// UID of the deployment created by the restore.
	DataServiceDeploymentId string `copier:"must,nopanic"`
	// UID of the project.
	ProjectId string `copier:"must,nopanic"`
}

type V1PhaseType string

type Templatev1Status struct {
	Phase *StatusPhase `copier:"must,nopanic"`
}

type Template struct {
	Meta   *V1Meta           `copier:"must,nopanic"`
	Config *V1Config         `copier:"must,nopanic"`
	Status *Templatev1Status `copier:"must,nopanic"`
}

type ProtobufAny struct {
	Type                 *string `copier:"must,nopanic"`
	AdditionalProperties map[string]interface{}
}

type V1Metadata struct {
	KubeServerVersion *string             `copier:"must,nopanic"`
	KubePlatform      *V1KubePlatformType `copier:"must,nopanic"`
	PxeMetadata       *V1PXEMetadata      `copier:"must,nopanic"`
}

type V1KubePlatformType string

type V1PXEMetadata struct {
	CsiEnabled       *bool   `copier:"must,nopanic"`
	ServiceName      *string `copier:"must,nopanic"`
	ServiceNamespace *string `copier:"must,nopanic"`
	Version          *string `copier:"must,nopanic"`
}

type WorkflowResiliency struct {
	ResiliencyFlag bool `copier:"must,nopanic"`
}

type V1PageBasedPaginationResponse struct {
	TotalRecords *string `copier:"must,nopanic"`
	CurrentPage  *string `copier:"must,nopanic"`
	PageSize     *string `copier:"must,nopanic"`
	TotalPages   *string `copier:"must,nopanic"`
	NextPage     *string `copier:"must,nopanic"`
	PrevPage     *string `copier:"must,nopanic"`
}
