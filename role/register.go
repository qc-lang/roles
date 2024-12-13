package role

import (
	"errors"
	"fmt"
	"github.com/restartfu/gophig"
	"os"
	"sort"
	"strings"
	"sync"
)

var (
	// roleMu is a mutex that protects the roles slice.
	roleMu sync.Mutex
	// roles is a slice of all roles.
	roles []Role
	// rolesName is a map of all roles.
	rolesName = map[string]Role{}
)

// register registers a role.
func register(rls ...Role) {
	roleMu.Lock()
	for _, r := range rls {
		rolesName[strings.ToLower(r.Name())] = r
		roles = append(roles, r)
	}
	roleMu.Unlock()
}

// Load loads all role from a folder.
func Load(folder string, marshaler gophig.Marshaler) error {
	if marshaler == nil {
		marshaler = gophig.JSONMarshaler{}
	}

	folder = strings.TrimSuffix(folder, "/")
	files, err := os.ReadDir(folder)
	if err != nil {
		return errors.New(fmt.Sprintf("error loading role: %v", err))
	}

	var newRoles []Role
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		r, err := loadRole(folder+"/"+file.Name(), marshaler)
		if err != nil {
			return errors.New(fmt.Sprintf("error loading role %s: %v", file.Name(), err))
		}
		newRoles = append(newRoles, r)
	}

	roleMu.Lock()
	roles = make([]Role, 0)
	rolesName = map[string]Role{}
	roleMu.Unlock()

	sortRoles(newRoles)
	register(newRoles...)

	for _, r := range newRoles {
		if r.inherits == "" {
			continue
		}
		if r.inherits == r.Name() {
			return errors.New(fmt.Sprintf("role %s and role %s have circular inheritance", r.Name(), r.Name()))
		}
		_, ok := rolesName[strings.ToLower(r.inherits)]
		if !ok {
			return errors.New(fmt.Sprintf("role %s inherits from the role %s which does not exist", r.Name(), r.inherits))
		}

		// check for circular inheritance
		for _, r2 := range newRoles {
			if r2.Name() == r.inherits {
				if r2.inherits == r.Name() {
					return errors.New(fmt.Sprintf("role %s and role %s have circular inheritance", r.Name(), r2.Name()))
				}
			}
		}
	}
	return nil
}

// roleData is a struct that is used to decode roles from JSON.
type roleData struct {
	Name     string `json:"name"`
	Inherits string `json:"inherits,omitempty"`
	Colour   string `json:"colour,omitempty"`
	Tier     int    `json:"tier"`
}

// loadRole loads a role from a file.
func loadRole(filePath string, marshaler gophig.Marshaler) (Role, error) {
	var data roleData
	err := gophig.GetConfComplex(filePath, marshaler, &data)
	if err != nil {
		return Role{}, err
	}

	for _, r := range roles {
		if r.Name() == data.Name {
			return Role{}, errors.New("role with name " + data.Name + " already exists")
		}
		if r.tier == data.Tier {
			return Role{}, errors.New(fmt.Sprintf("role with tier %d already exists", data.Tier))
		}
	}

	return Role{
		name:     data.Name,
		inherits: data.Inherits,
		colour:   data.Colour,
		tier:     data.Tier,
	}, nil
}

// sortRoles sorts the roles by their tier.
func sortRoles(rls []Role) {
	roleMu.Lock()
	sort.SliceStable(rls, func(i, j int) bool {
		return rls[i].tier < rls[j].tier
	})
	roleMu.Unlock()
}

// All returns all role that are currently registered.
func All() []Role {
	roleMu.Lock()
	r := make([]Role, len(roles))
	copy(r, roles)
	roleMu.Unlock()
	return r
}

// ByName returns a role by its name.
func ByName(name string) (Role, bool) {
	roleMu.Lock()
	r, ok := rolesName[strings.ToLower(name)]
	roleMu.Unlock()
	return r, ok
}

// ByNameMust returns a role by its name. If the role does not exist, it panics.
func ByNameMust(name string) Role {
	r, ok := ByName(name)
	if !ok {
		panic(fmt.Sprintf("role %s does not exist", name))
	}
	return r
}
