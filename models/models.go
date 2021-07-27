package models

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
)

type Question struct {
	ID         int    `json:"questionId,string"`
	FrontendID string `json:"questionFrontendId"`
	Title      string
	TitleSlug  string
	Content    string
	Difficulty string
	IsPaidOnly bool

	TranslatedTitle   string
	TranslatedContent string

	Status string

	Tags    []*Tag
	Submits map[int64]*Submit
}

type Tag struct {
	Name           string
	Slug           string
	TranslatedName string
}

type Submit struct {
	ID            int64 `json:",string"`
	StatusDisplay string
	Lang          string
	Runtime       string
	Memory        string
	URL           string
	Code          string
	Timestamp     int64 `json:",string"`
}

func (question *Question) String() string {
	return fmt.Sprintf(`Question %s %s status:%s Tags: %v`, question.FrontendID, question.Title, question.Status, question.Tags)
}

func (tag *Tag) String() string {
	return fmt.Sprintf("Tag %s Name: %s", tag.Slug, tag.Name)
}

func (submit *Submit) String() string {
	return fmt.Sprintf("Submit %d lang: %s, status: %s", submit.ID, submit.Lang, submit.StatusDisplay)
}

func (submit *Submit) RawRuntime() int {
	rawRuntimeRe := regexp.MustCompile(`(?m)(\d+) ms`)
	rawRuntime, err := strconv.Atoi(rawRuntimeRe.FindString(submit.Runtime))
	if err != nil {
		return math.MaxInt32
	}
	return rawRuntime
}
