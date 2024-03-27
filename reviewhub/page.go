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

func FilterReviewList(ls []ReviewList, reviewer User, includeApproved bool) []ReviewList {
	var filtered []ReviewList
	for _, l := range ls {
		pages := []ReviewPage{}
		for _, page := range l.Pages {
			if !Contains(page.Reviewers, reviewer) {
				continue
			}

			if includeApproved || Contains(page.ApprovedReviewers, reviewer) {
				continue
			}
			pages = append(pages, page)
		}

		filtered = append(
			filtered,
			ReviewList{
				Name:  l.Name,
				Pages: pages,
			},
		)
	}
	return filtered
}
