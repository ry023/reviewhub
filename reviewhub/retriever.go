package reviewhub

type Retriever interface {
	Retrieve([]User) (*ReviewList, error)
}
