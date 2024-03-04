package reviewhub

type Config struct {
	Providers []ProviderConfig
	Notifiers []NotifierConfig
	Users     []User
}

type NotifierConfig struct {
	Name     string
	Type     string
	MetaData map[string]string
}

type ProviderConfig struct {
	Name     string
	Type     string
	MetaData map[string]string
}
