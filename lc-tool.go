package main

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"leetcode-tools/actions"
	"leetcode-tools/models"
	"leetcode-tools/settings"
	"log"
	"os"
	"time"
)

var (
	questionIDMap   = make(map[int64]*models.Question)
	questionSlugMap = make(map[string]*models.Question)
	submitIDMap     = make(map[int64]*models.Submit)
)

func main() {
	// recovery local data to map
	err := recovery()

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
		q, ok := questionIDMap[question.ID]
		if question.Status == "ac" {
			if ok {
				if q.Status == "ac" {
					continue
				} else {
					q.Status = "ac"
					question = q
				}
			} else {
				questionIDMap[question.ID] = question
				questionSlugMap[question.TitleSlug] = question
			}
		} else {
			continue
		}

		log.Printf("处理题目: %d, %s", question.ID, question.Title)
		if question.Content == "" {
			q, err := actions.User.GetQuestionDetail(question.TitleSlug)
			if err == nil {
				question.Content = q.Content
				if q.TranslatedContent != "" {
					question.TranslatedContent = q.TranslatedContent
				}
				if q.TranslatedTitle != "" {
					question.TranslatedTitle = q.TranslatedTitle
				}
				if question.Tags == nil {
					question.Tags = q.Tags
				}
			}
		}

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
				submit.Code = code
				time.Sleep(1 * time.Second)
			}
			time.Sleep(2 * time.Second)
		}
	}
	save()
	// gene code to file

	// gene commit each file

}

func save() {
	data := make(map[string]interface{})
	data["questions"] = make([]*models.Question, 0)
	data["submits"] = make([]*models.Submit, 0)
	for _, q := range questionIDMap {
		data["questions"] = append(data["questions"].([]*models.Question), q)
	}
	for _, s := range submitIDMap {
		data["submits"] = append(data["submits"].([]*models.Submit), s)
	}
	// save data to file
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("leetcode-data.json", jsonData, 0644)
	if err != nil {
		log.Fatalln("save error")
	}
}

func recovery() error {
	file, err := os.Open("leetcode-data.json")
	if err != nil {
		log.Printf("%v", err)
		log.Println("跳过恢复")
		return err
	}
	defer file.Close()
	recoveryData, _ := ioutil.ReadAll(file)
	for _, q := range gjson.GetBytes(recoveryData, "questions").Array() {
		var question models.Question
		if err := json.Unmarshal([]byte(q.String()), &question); err != nil {
			questionIDMap[question.ID] = &question
			questionSlugMap[question.TitleSlug] = &question
		}
	}
	for _, q := range gjson.GetBytes(recoveryData, "submits").Array() {
		var submit models.Submit
		if err := json.Unmarshal([]byte(q.String()), &submit); err != nil {
			submitIDMap[submit.ID] = &submit
		}
	}
	return err
}
