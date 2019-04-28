package models

import "fmt"

type Question struct {
	ID         int64 `json:"questionId,string"`
	FrontendID int64 `json:"questionFrontendId,string"`
	Title      string
	TitleSlug  string
	Content    string

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
}

func (question *Question) String() string {
	return fmt.Sprintf(`Question %d %s status:%s
Tags: %v
Submits: %v`, question.FrontendID, question.TitleSlug, question.Status, question.Tags, question.Submits)
}

func (tag *Tag) String() string {
	return fmt.Sprintf("Tag %s Name: %s, TranslatedName: %s", tag.Slug, tag.Name, tag.TranslatedName)
}

func (submit *Submit) String() string {
	return fmt.Sprintf("Submit %d lang: %s, status: %s", submit.ID, submit.Lang, submit.StatusDisplay)
}
