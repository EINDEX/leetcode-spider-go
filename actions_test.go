package main

import (
	"fmt"
	"net/url"
	"testing"
)



func TestUser_Login(t *testing.T) {
	fmt.Println(Setting.Username, Setting.Password)
	User.Login(Setting.Username, Setting.Password)
	URL, _ := url.Parse("leetcode.com")
	fmt.Print(User.client.Jar.Cookies(URL))
}

func TestUser_GetAllQuestionStatus(t *testing.T) {
	User.Login(Setting.Username, Setting.Password)
	data, err := User.GetAllQuestionStatus()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}

func TestUser_GetSubmitHistory(t *testing.T) {
	User.Login(Setting.Username, Setting.Password)
	data, newLastKey ,err := User.GetSubmitHistory(1, "")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(newLastKey)
	fmt.Println(data)
}

func TestUser_GetQuestionDetail(t *testing.T) {
	User.Login(Setting.Username, Setting.Password)
	data, err := User.GetQuestionDetail("two-sum")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}

func TestUser_GetSubmitDetail(t *testing.T) {
	User.Login(Setting.Username, Setting.Password)
	data, err := User.GetSubmitDetail(1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}