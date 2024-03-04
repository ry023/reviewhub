package reviewhub

type User struct {
	Name     string
	MetaData map[string]string
	Unknown  bool
}

func NewUnknownUser(name string) *User {
	return &User{
		Name:    name,
		Unknown: true,
	}
}

func Contains(users []User, target User) bool {
  for _, u := range users {
    if u.Name == target.Name {
      return true
    }
  }
  return false
}
