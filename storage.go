package main

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"time"
)

type Storage struct {
	timePattern                                                                                            string
	invite_path, emails_path, policy_path, configuration_path, displayprefs_path, ombi_path, profiles_path string
	invites                                                                                                Invites
	profiles                                                                                               map[string]Profile
	defaultProfile                                                                                         string
	emails, policy, configuration, displayprefs, ombi_template                                             map[string]interface{}
}

// timePattern: %Y-%m-%dT%H:%M:%S.%f

type Profile struct {
	Admin         bool                   `json:"admin,omitempty"`
	LibraryAccess string                 `json:"libraries,omitempty"`
	FromUser      string                 `json:"fromUser,omitempty"`
	Policy        map[string]interface{} `json:"policy,omitempty"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
	Displayprefs  map[string]interface{} `json:"displayprefs,omitempty"`
	Default       bool                   `json:"default,omitempty"`
}

type Invite struct {
	Created       time.Time                  `json:"created"`
	NoLimit       bool                       `json:"no-limit"`
	RemainingUses int                        `json:"remaining-uses"`
	ValidTill     time.Time                  `json:"valid_till"`
	Email         string                     `json:"email"`
	UsedBy        [][]string                 `json:"used-by"`
	Notify        map[string]map[string]bool `json:"notify"`
	Profile       string                     `json:"profile"`
}

type Invites map[string]Invite

func (st *Storage) loadInvites() error {
	return loadJSON(st.invite_path, &st.invites)
}

func (st *Storage) storeInvites() error {
	return storeJSON(st.invite_path, st.invites)
}

func (st *Storage) loadEmails() error {
	return loadJSON(st.emails_path, &st.emails)
}

func (st *Storage) storeEmails() error {
	return storeJSON(st.emails_path, st.emails)
}

func (st *Storage) loadPolicy() error {
	return loadJSON(st.policy_path, &st.policy)
}

func (st *Storage) storePolicy() error {
	return storeJSON(st.policy_path, st.policy)
}

func (st *Storage) loadConfiguration() error {
	return loadJSON(st.configuration_path, &st.configuration)
}

func (st *Storage) storeConfiguration() error {
	return storeJSON(st.configuration_path, st.configuration)
}

func (st *Storage) loadDisplayprefs() error {
	return loadJSON(st.displayprefs_path, &st.displayprefs)
}

func (st *Storage) storeDisplayprefs() error {
	return storeJSON(st.displayprefs_path, st.displayprefs)
}

func (st *Storage) loadOmbiTemplate() error {
	return loadJSON(st.ombi_path, &st.ombi_template)
}

func (st *Storage) storeOmbiTemplate() error {
	return storeJSON(st.ombi_path, st.ombi_template)
}

func (st *Storage) loadProfiles() error {
	err := loadJSON(st.profiles_path, &st.profiles)
	for name, profile := range st.profiles {
		if profile.Default {
			st.defaultProfile = name
		}
		change := false
		if profile.Policy["IsAdministrator"] != nil {
			profile.Admin = profile.Policy["IsAdministrator"].(bool)
			change = true
		}
		if profile.Policy["EnabledFolders"] != nil {
			length := len(profile.Policy["EnabledFolders"].([]interface{}))
			if length == 0 {
				profile.LibraryAccess = "All"
			} else {
				profile.LibraryAccess = strconv.Itoa(length)
			}
			change = true
		}
		if profile.FromUser == "" {
			profile.FromUser = "Unknown"
			change = true
		}
		if change {
			st.profiles[name] = profile
		}
	}
	if st.defaultProfile == "" {
		for n := range st.profiles {
			st.defaultProfile = n
		}
	}
	return err
}

func (st *Storage) storeProfiles() error {
	return storeJSON(st.profiles_path, st.profiles)
}

func (st *Storage) migrateToProfile() error {
	st.loadPolicy()
	st.loadConfiguration()
	st.loadDisplayprefs()
	st.loadProfiles()
	st.profiles["Default"] = Profile{
		Policy:        st.policy,
		Configuration: st.configuration,
		Displayprefs:  st.displayprefs,
	}
	return st.storeProfiles()
}

func loadJSON(path string, obj interface{}) error {
	var file []byte
	var err error
	file, err = ioutil.ReadFile(path)
	if err != nil {
		file = []byte("{}")
	}
	err = json.Unmarshal(file, &obj)
	return err
}

func storeJSON(path string, obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, data, 0644)
	return err
}
