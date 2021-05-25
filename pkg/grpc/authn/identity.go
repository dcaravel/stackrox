package authn

import (
	"time"

	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/auth/authproviders"
)

//go:generate mockgen-wrapper

// Identity represents the identity of an entity accessing a service.
type Identity interface {
	UID() string
	FriendlyName() string
	//FullName could be empty
	FullName() string

	User() *storage.UserInfo
	Permissions() *storage.ResourceToAccess
	Roles() []*storage.Role

	Service() *storage.ServiceIdentity
	Attributes() map[string][]string

	Expiry() time.Time
	ExternalAuthProvider() authproviders.Provider
}
