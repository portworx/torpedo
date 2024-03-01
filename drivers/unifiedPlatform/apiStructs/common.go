package apiStructs

import "time"

type PaginationRequest struct {
	PageNumber int
	PageSize   int
}

// ProtobufAny4 `Any` contains an arbitrary serialized protocol buffer message along with a URL that describes the type of the serialized message.  Protobuf library provides support to pack/unpack Any values in the form of utility functions or additional generated methods of the Any type.  Example 1: Pack and unpack a message in C++.      Foo foo = ...;     Any any;     any.PackFrom(foo);     ...     if (any.UnpackTo(&foo)) {       ...     }  Example 2: Pack and unpack a message in Java.      Foo foo = ...;     Any any = Any.pack(foo);     ...     if (any.is(Foo.class)) {       foo = any.unpack(Foo.class);     }     // or ...     if (any.isSameTypeAs(Foo.getDefaultInstance())) {       foo = any.unpack(Foo.getDefaultInstance());     }   Example 3: Pack and unpack a message in Python.      foo = Foo(...)     any = Any()     any.Pack(foo)     ...     if any.Is(Foo.DESCRIPTOR):       any.Unpack(foo)       ...   Example 4: Pack and unpack a message in Go       foo := &pb.Foo{...}      any, err := anypb.New(foo)      if err != nil {        ...      }      ...      foo := &pb.Foo{}      if err := any.UnmarshalTo(foo); err != nil {        ...      }  The pack methods provided by protobuf library will by default use 'type.googleapis.com/full.type.name' as the type URL and the unpack methods only use the fully qualified type name after the last '/' in the type URL, for example \"foo.bar.com/x/y.z\" will yield type name \"y.z\".  JSON ==== The JSON representation of an `Any` value uses the regular representation of the deserialized, embedded message, with an additional field `@type` which contains the type URL. Example:      package google.profile;     message Person {       string first_name = 1;       string last_name = 2;     }      {       \"@type\": \"type.googleapis.com/google.profile.Person\",       \"firstName\": <string>,       \"lastName\": <string>     }  If the embedded message type is well-known and has a custom JSON representation, that representation will be embedded adding a field `value` which holds the custom JSON in addition to the `@type` field. Example (for message [google.protobuf.Duration][]):      {       \"@type\": \"type.googleapis.com/google.protobuf.Duration\",       \"value\": \"1.212s\"     }
type ProtobufAny4 struct {
	// A URL/resource name that uniquely identifies the type of the serialized protocol buffer message. This string must contain at least one \"/\" character. The last segment of the URL's path must represent the fully qualified name of the type (as in `path/google.protobuf.Duration`). The name should be in a canonical form (e.g., leading \".\" is not accepted).  In practice, teams usually precompile into the binary all types that they expect it to use in the context of Any. However, for URLs which use the scheme `http`, `https`, or no scheme, one can optionally set up a type server that maps type URLs to message definitions as follows:  * If no scheme is provided, `https` is assumed. * An HTTP GET on the URL must yield a [google.protobuf.Type][]   value in binary format, or produce an error. * Applications are allowed to cache lookup results based on the   URL, or have them precompiled into a binary to avoid any   lookup. Therefore, binary compatibility needs to be preserved   on changes to types. (Use versioned type names to manage   breaking changes.)  Note: this functionality is not currently available in the official protobuf release, and it is not used for type URLs beginning with type.googleapis.com. As of May 2023, there are no widely used type server implementations and no plans to implement one.  Schemes other than `http`, `https` (or the empty scheme) might be used with implementation specific semantics.
	Type                 *string `json:"@type,omitempty"`
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

type V1Meta struct {
	Uid             *string            `json:"uid,omitempty"`
	Name            *string            `json:"name,omitempty"`
	Description     *string            `json:"description,omitempty"`
	ResourceVersion *string            `json:"resourceVersion,omitempty"`
	CreateTime      *time.Time         `json:"createTime,omitempty"`
	UpdateTime      *time.Time         `json:"updateTime,omitempty"`
	Labels          *map[string]string `json:"labels,omitempty"`
	Annotations     *map[string]string `json:"annotations,omitempty"`
	ParentReference *V1Reference
}

type Config struct {
	UserEmail   *string `copier:"must,nopanic"`
	DnsName     *string `copier:"must,nopanic"`
	DisplayName *string `copier:"must,nopanic"`
}

type V1Config1 struct {
	References *Reference `copier:"must,nopanic"`
	// Flag to enable TLS for the Data Service.
	TlsEnabled *bool `copier:"must,nopanic"`
	// A deployment topology contains a number of nodes that have various attributes as a collective group.
	DeploymentTopologies []DeploymentTopology `copier:"must,nopanic"`
}

type V1Config3 struct {
	ActorId      *string `json:"actorId,omitempty"`
	ActorType    *string `json:"actorType,omitempty"`
	AccessPolicy *V1AccessPolicy
}

type V1Deployment struct {
	Meta   Meta      `copier:"must,nopanic"`
	Config V1Config1 `copier:"must,nopanic"`
}

type SetRbacToken struct {
	SetRbac  bool   `copier:"must,nopanic"`
	JwtToken string `copier:"must,nopanic"`
}

type Status struct {
	Phase string
}

type V1Reference struct {
	Type    *string `json:"type,omitempty"`
	Version *string `json:"version,omitempty"`
	Uid     *string
}
