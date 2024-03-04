package reviewhub

type Notifier interface {
	Notify(User, []ReviewList) error
}
