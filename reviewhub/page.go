package reviewhub

type Page struct {
	Title string
	Url   string
	Owner User
}

type ReviewPage struct {
	Page

	ApprovedReviewers []User
	Reviewers         []User
}

func NewReviewPage(title, url string, owner User, approved []User, reviewers []User) ReviewPage {
	return ReviewPage{
		Page: Page{
			Title: title,
			Url:   url,
			Owner: owner,
		},
		ApprovedReviewers: approved,
		Reviewers:         reviewers,
	}
}

type ReviewList struct {
	Name  string
	Pages []ReviewPage
}

func (r ReviewList) FilterByReviewer(u User, includeApproved bool) ReviewList {
	// Filter Review Pages
	l := []ReviewPage{}
	for _, page := range r.Pages {
		if !Contains(page.Reviewers, u) {
			continue
		}

		if includeApproved || Contains(page.ApprovedReviewers, u) {
			continue
		}
		l = append(l, page)
	}

	return ReviewList{
		Name:  r.Name,
		Pages: l,
	}
}
