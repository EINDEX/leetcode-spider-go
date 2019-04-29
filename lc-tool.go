package main

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"leetcode-tools/actions"
	"leetcode-tools/models"
	"leetcode-tools/settings"
	"leetcode-tools/utils"
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

func main() {
	// recovery local data to map
	recovery()

	actions.User.Login(settings.Setting.Username, settings.Setting.Password)
	// get question status
	log.Println("开始同步问题状态")
	questions, err := actions.User.GetAllQuestionStatus()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("同步问题状态完成")

	for _, question := range questions {
		// find questions ac
		question = checkQuestion(question)
		if question == nil {
			continue
		}
		log.Printf("处理题目: %d, %s", question.ID, question.Title)

		fetchQuestionSubmitCode(question)
	}
	save()
	// gene code to file
	geneFiles()
}

func geneFiles() {
	questionLang := make(map[int][][]string)
	for id, question := range questionIDMap {
		path := fmt.Sprintf("/%d-%s/", question.FrontendID, question.TitleSlug)
		if err := os.MkdirAll(settings.Setting.Out+path, os.ModePerm); err != nil {
			panic(err)
		}
		langSubmit := make(map[string]*models.Submit)
		for _, submit := range question.Submits {
			s, ok := langSubmit[submit.Lang]
			if (ok && s.Runtime > submit.Runtime) || !ok {
				langSubmit[submit.Lang] = submit
			}
		}

		questionLangInfo := make([][]string, 0, len(langSubmit))
		listLang := make([]string, 0, len(langSubmit))
		for _, s := range langSubmit {
			listLang = append(listLang, s.Lang)
			codePath := path + question.TitleSlug + "." + s.Lang + "." + utils.GetLangSuffix(s.Lang)
			questionLangInfo = append(questionLangInfo, []string{s.Lang, codePath})
			if _, err := os.Stat(settings.Setting.Out + codePath); !os.IsNotExist(err) {
				continue
			}
			if err := ioutil.WriteFile(settings.Setting.Out+codePath, []byte(s.Code), 0644); err != nil {
				log.Printf("write file error: %v", err)
			}
			if err := utils.ExecCommend("git", "add", settings.Setting.Out+codePath); err != nil {
				log.Fatalln(err)
			}
			if err := utils.ExecCommend("git", "commit", "--date", time.Unix(s.Timestamp, 0).Format("2006-01-02 15:04:05"), "-m", fmt.Sprintf("%s %s solution", question.TitleSlug, s.Lang)); err != nil {
				log.Fatalln(err)
			}
		}
		questionLang[id] = questionLangInfo
		sort.Strings(listLang)

		utils.QuestionRender(settings.Setting.Out+path+"README.md", question, langSubmit, listLang)
		if settings.Setting.Enter == "cn" {
			utils.QuestionRender(settings.Setting.Out+path+"README-ZH.md", question, langSubmit, listLang)
		}
	}
	keys := make([]int, len(questionIDMap))
	for i, _ := range questionIDMap {
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

	utils.Render(solutions, "README.md", "global")
	if settings.Setting.Enter == "cn" {
		utils.Render(solutions, "README-ZN.md", "cn")
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
		log.Printf("获取 %d, %s 提交状态 ", question.ID, question.Title)
		pageSubmits, newLastKey, err := actions.User.GetSubmitHistory(1, lastKey, question.TitleSlug)
		if err != nil {
			times -= 1
			if times < 0 {
				break
			}
			continue
		}
		if len(pageSubmits) < 20 {
			hasNext = false
		}
		lastKey = newLastKey
		log.Printf("答案数量 %d", len(pageSubmits))
		for _, submit := range pageSubmits {
			if submit.StatusDisplay != "Accepted" {
				continue
			}
			if _, ok := submitIDMap[submit.ID]; ok {
				continue
			}
			log.Printf("下载 %d, %s submit %d 代码 ", question.ID, question.Title, submit.ID)
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

func checkQuestion(question *models.Question) *models.Question {
	flag := true
	q, ok := questionIDMap[question.ID]
	if question.Status == "ac" {
		if ok {
			if q.Status == "ac" {
				flag = false
			} else {
				q.Status = "ac"
			}
		} else {
			questionIDMap[question.ID] = question
			questionSlugMap[question.TitleSlug] = question
			q = question
		}
	} else {
		return nil
	}
	if settings.Setting.Enter == "cn" {
		q.TranslatedTitle = question.TranslatedTitle
		if q.TranslatedContent == "" {
			fillQuestionContent(q)
		}
	} else if q.Content == "" {
		fillQuestionContent(q)
	}
	if !flag {
		return nil
	}
	return q
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
	log.Println("开始数据恢复")
	file, err := os.Open(settings.Setting.SaveFile)
	if err != nil {
		log.Printf("%v", err)
		log.Println("跳过恢复")
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
