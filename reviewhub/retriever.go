package reviewhub

type Retriever interface {
	Retrieve(RetrieverConfig, []User) (*ReviewList, error)
}
