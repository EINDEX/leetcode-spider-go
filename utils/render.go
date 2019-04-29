package utils

import (
	"leetcode-tools/models"
	"leetcode-tools/settings"
	"log"
	"os"
	"text/template"
	"time"
)

func QuestionRender(readmePath string, question *models.Question, langSubmit map[string]*models.Submit, listLang []string) {
	tmpl, err := template.ParseFiles("template/question_readme.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(readmePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	err = tmpl.Execute(f, &struct {
		Question   *models.Question
		LangSubmit map[string]*models.Submit
		ListLang   []string
		Mode       string
	}{
		Question:   question,
		LangSubmit: langSubmit,
		ListLang:   listLang,
	})
}

func Render(solutions []*map[string]interface{}, filename, mode string) {
	temp, err := template.ParseFiles("template/readme.tmpl")
	if err != nil {
		log.Fatalln(err)
	}
	f, err := os.Create(settings.Setting.Out + "/" + filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	err = temp.Execute(f, &struct {
		Time     string
		Solution []*map[string]interface{}
		Mode     string
	}{
		Time:     time.Now().Format("2006-01-02"),
		Solution: solutions,
		Mode:     mode,
	})
	if err != nil {
		log.Fatal(err)
	}
}
