package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"leetcode-spider-go/actions"
	"leetcode-spider-go/models"
	"leetcode-spider-go/settings"
	"leetcode-spider-go/utils"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

var (
	questionIDMap   = make(map[int]*models.Question)
	questionSlugMap = make(map[string]*models.Question)
	submitIDMap     = make(map[int64]*models.Submit)
)

//go:embed template
var fs embed.FS

func main() {
	// recovery local data to map
	recovery()

	actions.User.Login(settings.Setting.Username, settings.Setting.Password)
	updateViaQuestionFetchAction(actions.User.GetAllQuestionStatus, false)
	updateViaQuestionFetchAction(actions.User.GetRecentSubmission, true)

	geneFiles()

}

func updateViaQuestionFetchAction(questionFetchFunc func() ([]*models.Question, error), force bool) {
	questions, err := questionFetchFunc()
	if err != nil {
		log.Println(err)
		return
	}
	fetchSubmits(questions, force)
}

func fetchSubmits(questions []*models.Question, force bool) {
	for _, question := range questions {
		if !force && !shouldUpdateQuestion(question) {
			continue
		}
		fillQuestionContent(question)
		log.Printf("process question: %d.%s", question.ID, question.Title)
		fetchQuestionSubmitCode(question)
	}
	save()
}

func geneFiles() {
	questionLang := make(map[int][][]string)
	for id, question := range questionIDMap {
		path := fmt.Sprintf("/leetcode/%d-%s/", question.ID, question.TitleSlug)
		if err := os.MkdirAll(settings.Setting.Out+path, os.ModePerm); err != nil {
			panic(err)
		}
		langSubmit := make(map[string]*models.Submit)
		for _, submit := range question.Submits {
			lang := strings.Replace(submit.Lang, "python3", "python", -1)
			s, ok := langSubmit[lang]
			if (ok && s.RawRuntime() > submit.RawRuntime()) || !ok {
				langSubmit[lang] = submit
			}
		}

		questionLangInfo := make([][]string, 0, len(langSubmit))
		listLang := make([]string, 0, len(langSubmit))
		timestamp := int64(0)
		for _, s := range langSubmit {
			listLang = append(listLang, strings.Replace(s.Lang, "python3", "python", -1))
			codePath := path + question.TitleSlug + "." + s.Lang + "." + utils.GetLangSuffix(s.Lang)
			questionLangInfo = append(questionLangInfo, []string{s.Lang, codePath})
			if _, err := os.Stat(settings.Setting.Out + codePath); !os.IsNotExist(err) {
				continue
			}
			if err := ioutil.WriteFile(settings.Setting.Out+codePath, []byte(s.Code), 0644); err != nil {
				log.Printf("write file error: %v", err)
			}
			utils.GitAddAndCommand(settings.Setting.Out+path, fmt.Sprintf("%s %s solution", question.TitleSlug, s.Lang), s.Timestamp)
			if s.Timestamp > timestamp {
				timestamp = s.Timestamp
			}
		}
		questionLang[id] = questionLangInfo
		sort.Strings(listLang)
		utils.QuestionRender(fs, settings.Setting.Out+path+"README.md", question, langSubmit, listLang)
		if settings.Setting.Enter == "cn" {
			utils.QuestionRender(fs, settings.Setting.Out+path+"README-ZH.md", question, langSubmit, listLang)
		}
		utils.GitAddAndCommand(settings.Setting.Out+path, fmt.Sprintf("%s solution", question.TitleSlug), timestamp)
	}
	keys := make([]int, len(questionIDMap))
	for i := range questionIDMap {
		keys = append(keys, i)
	}
	sort.Ints(keys)

	solutions := make([]*map[string]interface{}, 0)
	for _, qid := range keys {
		if q, ok := questionIDMap[qid]; ok {
			m := map[string]interface{}{
				"question": q,
				"langs":    questionLang[q.ID],
			}
			solutions = append(solutions, &m)
		}
	}

	utils.ReadmeRender(fs, solutions, "README.md", "global")
	if settings.Setting.Enter == "cn" {
		utils.ReadmeRender(fs, solutions, "README-ZH.md", "cn")
	}
	if settings.Setting.EnablePush {
		utils.GitAddAndCommand(settings.Setting.Out, "generate readme", 0)
		utils.GitPush()
	}
}

func fetchQuestionSubmitCode(question *models.Question) {
	times := 5
	lastKey := ""
	hasNext := true
	for {
		// get submit history
		if !hasNext {
			break
		}
		log.Printf("fetch %d.%s submit status ", question.ID, question.Title)
		pageSubmits, newLastKey, err := actions.User.GetSubmitHistory(1, lastKey, question.TitleSlug)
		if err != nil {
			times -= 1
			if times < 0 {
				break
			}
			time.Sleep(1 * time.Second)
			continue
		}
		if len(pageSubmits) < 20 {
			hasNext = false
		}
		lastKey = newLastKey
		log.Printf("nums of answers: %d", len(pageSubmits))
		for _, submit := range pageSubmits {
			if submit.StatusDisplay != "Accepted" {
				continue
			}
			if _, ok := submitIDMap[submit.ID]; ok {
				continue
			}
			log.Printf("download %d.%s submit %d code ", question.ID, question.Title, submit.ID)
			if question.Submits == nil {
				question.Submits = make(map[int64]*models.Submit)
			}
			submitIDMap[submit.ID] = submit
			question.Submits[submit.ID] = submit
			// get code
			code, err := actions.User.GetSubmitDetail(submit.ID)
			if err != nil {

			}
			submit.Code = strings.ReplaceAll(code, "\r\n", "\n")
			time.Sleep(1 * time.Second)
		}
		time.Sleep(2 * time.Second)
	}
}

func shouldUpdateQuestion(question *models.Question) bool {
	_, ok := questionIDMap[question.ID]
	if !ok && question.Status == "ac" {
		questionIDMap[question.ID] = question
		questionSlugMap[question.TitleSlug] = question
		return true
	}
	return false
}

func fillQuestionContent(question *models.Question) {
	q, err := actions.User.GetQuestionDetail(question.TitleSlug)
	if err == nil {
		question.Content = q.Content
		if settings.Setting.Enter == "cn" && question.TranslatedContent == "" {
			question.TranslatedContent = q.TranslatedContent
			question.TranslatedTitle = q.TranslatedTitle
		}
		if question.Tags == nil {
			question.Tags = q.Tags
		}
	}
}

func save() {
	data := make(map[string]interface{})
	data["questions"] = make([]*models.Question, 0)
	for _, q := range questionIDMap {
		data["questions"] = append(data["questions"].([]*models.Question), q)
	}
	// save data to file
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(settings.Setting.SaveFile, jsonData, 0644)
	if err != nil {
		log.Fatalln("save error")
	}
}

func recovery() {
	log.Println("starting to load backup data")
	file, err := os.Open(settings.Setting.SaveFile)
	if err != nil {
		log.Printf("%v: skip loading backup data", err)
		return
	}
	defer file.Close()
	recoveryData, _ := ioutil.ReadAll(file)
	for _, q := range gjson.GetBytes(recoveryData, "questions").Array() {
		var question models.Question
		if err := json.Unmarshal([]byte(q.String()), &question); err != nil {
			continue
		}
		questionIDMap[question.ID] = &question
		questionSlugMap[question.TitleSlug] = &question
		if question.Submits != nil {
			for _, submit := range question.Submits {
				submitIDMap[submit.ID] = submit
				if submit.Code[0] == '\'' {
					submit.Code = "\"" + strings.TrimFunc(submit.Code, func(r rune) bool {
						return r == '\''
					}) + "\""
					var code string
					if err := json.Unmarshal([]byte(submit.Code), &code); err != nil {
						log.Printf("%v", err)
					}
					submit.Code = code
				}
			}
		} else {
			question.Submits = make(map[int64]*models.Submit)
		}
	}
	return
}
