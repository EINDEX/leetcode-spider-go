package utils

import (
	"io/fs"
	"leetcode-spider-go/models"
	"leetcode-spider-go/settings"
	"log"
	"os"
	"text/template"
	"time"
)

func QuestionRender(fs fs.FS, readmePath string, question *models.Question, langSubmit map[string]*models.Submit, listLang []string) {
	tmpl, err := template.ParseFS(fs, "template/question_readme.tmpl")
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

func ReadmeRender(fs fs.FS, solutions []*map[string]interface{}, filename, mode string) {
	temp, err := template.ParseFS(fs, "template/readme.tmpl")
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
