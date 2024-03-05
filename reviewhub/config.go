package reviewhub

type Config struct {
	Retrievers []RetrieverConfig
	Notifiers  []NotifierConfig
	Users      []User
}

type NotifierConfig struct {
	Name     string
	Type     string
	MetaData map[string]string
}

type RetrieverConfig struct {
	Name     string
	Type     string
	MetaData map[string]string
}
