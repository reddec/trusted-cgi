package types

type Credential struct {
	User  int
	Group int
}

func (cr *Credential) Equal(other *Credential) bool {
	return cr == other || (cr != nil && other != nil && cr.User == other.User && cr.Group == other.Group)
}

type File struct {
	Name string `json:"name"`
	Dir  bool   `json:"is_dir"`
}
