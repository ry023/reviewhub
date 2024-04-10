package reviewhub

type Notifier interface {
	Notify(NotifierConfig, User, []ReviewList) error
}
