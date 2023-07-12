package user_by_name_spec

const (
	// DefaultEmailDomain is the default EmailDomain for KeycloakUserSpec
	DefaultEmailDomain = "@cnbu.com"
)

// KeycloakUserSpec represents the specification for an KeycloakUserSpec
type KeycloakUserSpec struct {
	Name      string
	FirstName *string
	LastName  *string
	Email     *string
	Password  *string
}

func (s *KeycloakUserSpec) GetUsername() string {
	return s.Name
}

func (s *KeycloakUserSpec) SetUsername(name string) *KeycloakUserSpec {
	s.Name = name
	return s
}

func (s *KeycloakUserSpec) GetFirstName() string {
	return *s.FirstName
}

func (s *KeycloakUserSpec) SetFirstName(name string) *KeycloakUserSpec {
	s.FirstName = &name
	return s
}

func (s *KeycloakUserSpec) GetLastName() string {
	return *s.LastName
}

func (s *KeycloakUserSpec) SetLastName(name string) *KeycloakUserSpec {
	s.LastName = &name
	return s
}

func (s *KeycloakUserSpec) GetEmail() string {
	return *s.Email
}

func (s *KeycloakUserSpec) SetEmail(email string) *KeycloakUserSpec {
	s.Email = &email
	return s
}

func (s *KeycloakUserSpec) GetPassword() string {
	return *s.Password
}

func (s *KeycloakUserSpec) SetPassword(password string) *KeycloakUserSpec {
	s.Password = &password
	return s
}
