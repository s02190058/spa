package entity

import "time"

type Post struct {
	ID               int        `json:"id"`
	Type             string     `json:"type"`
	Category         string     `json:"category"`
	Title            string     `json:"title"`
	Text             string     `json:"text,omitempty"`
	URL              string     `json:"url,omitempty"`
	Author           *User      `json:"author"`
	Votes            []*Vote    `json:"votes"`
	Comments         []*Comment `json:"comments"`
	Views            int        `json:"views"`
	Score            int        `json:"score"`
	UpvotePercentage int        `json:"upvotePercentage"`
	Created          time.Time  `json:"created"`
}

func (p *Post) CalcAndSetScore() {
	upvotes := 0
	for _, vote := range p.Votes {
		if vote.Vote == 1 {
			upvotes++
		}
	}

	total := len(p.Votes)

	score := 2*upvotes - total // equal to upvotes - downvotes
	p.Score = score
}

func (p *Post) CalcAndSetUpvotePercentage() {
	upvotes := 0
	for _, vote := range p.Votes {
		if vote.Vote == 1 {
			upvotes++
		}
	}

	total := len(p.Votes)

	upvotePercentage := 0
	if total > 0 {
		upvotePercentage = (100 * upvotes) / total
	}
	p.UpvotePercentage = upvotePercentage
}
