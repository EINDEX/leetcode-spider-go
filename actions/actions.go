package actions

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"io/ioutil"
	"leetcode-spider-go/models"
	"leetcode-spider-go/settings"
	"leetcode-spider-go/utils"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

var (
	baseURL  string
	graphQL  string
	loginURL string
	User     *user
)

type user struct {
	status int
	client *http.Client
}

func init() {
	cookieJar, _ := cookiejar.New(nil)

	User = &user{
		status: 0,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Jar: cookieJar,
		},
	}
	if settings.Setting.Enter == "cn" {
		baseURL = "https://leetcode-cn.com/"
	} else {
		baseURL = "https://leetcode.com/"
	}
	graphQL = baseURL + "graphql/"
	loginURL = baseURL + "accounts/login/"
	resp, _ := User.client.Get(graphQL)
	defer resp.Body.Close()
}

func (u *user) geneGraphQLRequest(query io.Reader) (request *http.Request) {
	request, err := http.NewRequest("POST", graphQL, query)
	//fmt.Print(request.Header)
	if err != nil || request == nil {
		log.Fatalf("gene request fatal")
		return
	}

	request.Header.Add("content-type", "application/json")
	request.Header.Add("x-csrftoken", u.getCsrfToken())
	request.Header.Add("referer", baseURL)
	request.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36")

	uri, _ := url.Parse(baseURL)
	for _, cookie := range u.client.Jar.Cookies(uri) {
		request.AddCookie(cookie)
	}
	return
}

func (u *user) getCsrfToken() string {
	uri, _ := url.Parse(baseURL)
	for _, cookie := range u.client.Jar.Cookies(uri) {
		if cookie.Name == "csrftoken" {
			return cookie.Value
		}
	}
	return ""
}

func (u *user) Login(username, password string) {
	values := map[string]io.Reader{
		"csrfmiddlewaretoken": strings.NewReader(u.getCsrfToken()),
		"login":               strings.NewReader(username),
		"password":            strings.NewReader(password),
		"next":                strings.NewReader("/problems"),
	}
	buffer, header, err := utils.GetMultipartForm(values)
	if err != nil {
		log.Fatal(err)
	}

	request, err := http.NewRequest("POST", loginURL, &buffer)
	if err != nil || request == nil {
		log.Fatalf("%s Login Failed! %v\n", username, err)
		return
	}
	request.Header.Add("x-csrftoken", u.getCsrfToken())
	request.Header.Add("x-requested-with", "XMLHttpRequest")
	request.Header.Add("Referer", loginURL)
	request.Header.Add("Origin", baseURL)
	request.Header.Add("content-type", header)

	resp, err := u.client.Do(request)
	if err != nil {
		log.Fatalf("%s Login Failed! %v\n", username, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "LEETCODE_SESSION" {
				log.Printf("%s Login Success\n", username)
				u.status = 1
				return
			}
		}
	}
	log.Fatalf("%s Login Failed! \n", username)
	return
}

func (u *user) GetAllQuestionStatus() (data []*models.Question, err error) {
	query := "{\"query\":\"query allQuestions {\\n  allQuestions {\\n    ...questionSummaryFields\\n    __typename\\n  }\\n}\\n\\nfragment questionSummaryFields on QuestionNode {\\n  title\\n  titleSlug\\n  translatedTitle\\n  questionId\\n  questionFrontendId\\n  status\\n  difficulty\\n  isPaidOnly\\n  __typename\\n}\\n\",\"variables\":{},\"operationName\":\"allQuestions\"}"
	request := u.geneGraphQLRequest(strings.NewReader(query))
	resp, err := u.client.Do(request)
	if err != nil {
		log.Fatalf("Get Questions Status Failed %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Parse Questions Status Failed %v\n", err)
		return nil, err
	}
	res := gjson.GetBytes(body, "data.allQuestions")
	if err := json.Unmarshal([]byte(res.String()), &data); err != nil {
		return nil, err
	}
	return
}

func (u *user) GetSubmitHistory(page int, lastKey string, slug string) (data []*models.Submit, nextLastKey string, err error) {
	var query string

	pageSize := 20
	offset := (page - 1) * pageSize
	if lastKey == "" {
		lastKey = "null"
	} else {
		lastKey = "\n" + lastKey + "\n"
	}
	if slug == "" {
		query = "{\"query\":\"query submissions($offset: Int!, $limit: Int!, $lastKey: String) {\\n  submissionList(offset: $offset, limit: $limit, lastKey: $lastKey) {\\n    lastKey\\n    hasNext\\n    submissions {\\n      id\\n      statusDisplay\\n      lang\\n      runtime\\n      timestamp\\n      url\\n      isPending\\n      memory\\n      __typename\\n    }\\n    __typename\\n  }\\n}\\n\",\"variables\":{\"offset\":%d,\"limit\":%d,\"lastKey\":%s},\"operationName\":\"submissions\"}"
		query = fmt.Sprintf(query, offset, pageSize, lastKey)
	} else {
		query = "{\"query\":\"query submissions($offset: Int!, $limit: Int!, $lastKey: String, $questionSlug: String!, $markedOnly: Boolean, $lang: String) {\\n  submissionList(offset: $offset, limit: $limit, lastKey: $lastKey, questionSlug: $questionSlug, markedOnly: $markedOnly, lang: $lang) {\\n    lastKey\\n    hasNext\\n    \\n    submissions {\\n      id\\n      statusDisplay\\n      lang\\n      runtime\\n      timestamp\\n      url\\n      title\\n\\n      isPending\\n      memory\\n      __typename\\n      \\n    }\\n    __typename\\n  }\\n}\\n\",\"variables\":{\"offset\":%d,\"limit\":%d,\"lastKey\":%s,\"questionSlug\":\"%s\"},\"operationName\":\"submissions\"}"
		query = fmt.Sprintf(query, offset, pageSize, lastKey, slug)
	}
	req := u.geneGraphQLRequest(strings.NewReader(query))
	resp, err := u.client.Do(req)
	if err != nil {
		log.Fatalf("Get Submit History Failed %v\n", err)
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	res := gjson.GetBytes(body, "data.submissionList.submissions")
	nextLastKey = gjson.GetBytes(body, "data.submissionList.lastKey").Str
	if err := json.Unmarshal([]byte(res.String()), &data); err != nil {
		return nil, "", err
	}
	return data, nextLastKey, nil
}

func (u *user) GetQuestionDetail(questionSlug string) (question *models.Question, err error) {
	query := "{\"query\":\"query questionData($titleSlug: String!) {\\n  question(titleSlug: $titleSlug) {\\n    questionId\\n    questionFrontendId\\n    boundTopicId\\n    title\\n    titleSlug\\n    content\\n    translatedTitle\\n    translatedContent\\n    isPaidOnly\\n    difficulty\\n    likes\\n    dislikes\\n    isLiked\\n    similarQuestions\\n    contributors {\\n      username\\n      profileUrl\\n      avatarUrl\\n      __typename\\n    }\\n    langToValidPlayground\\n    topicTags {\\n      name\\n      slug\\n      translatedName\\n      __typename\\n    }\\n    companyTagStats\\n    codeSnippets {\\n      lang\\n      langSlug\\n      code\\n      __typename\\n    }\\n    stats\\n    hints\\n    solution {\\n      id\\n      canSeeDetail\\n      __typename\\n    }\\n    status\\n    sampleTestCase\\n    metaData\\n    judgerAvailable\\n    judgeType\\n    mysqlSchemas\\n    enableRunCode\\n    enableTestMode\\n    envInfo\\n    libraryUrl\\n    __typename\\n    submitUrl\\n  }\\n}\\n\",\"variables\":{\"titleSlug\":\"%s\"},\"operationName\":\"questionData\"}"
	query = fmt.Sprintf(query, questionSlug)
	req := u.geneGraphQLRequest(strings.NewReader(query))
	resp, err := u.client.Do(req)
	if err != nil {
		log.Fatalf("Get Submit History Failed %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	question = &models.Question{
		Tags:    make([]*models.Tag, 0),
		Submits: make(map[int64]*models.Submit),
	}
	data := gjson.GetBytes(body, "data.question").String()
	if err := json.Unmarshal([]byte(data), question); err != nil {
		return nil, err
	}
	return question, nil
}

func (u *user) GetSubmitDetail(submitID int64) (data string, err error) {
	query := "{\"operationName\":\"mySubmissionDetail\",\"variables\":{\"id\":\"%d\"},\"query\":\"query mySubmissionDetail($id: ID!) {\\n  submissionDetail(submissionId: $id) {\\n    id\\n    code\\n    runtime\\n    memory\\n    rawMemory\\n    statusDisplay\\n    timestamp\\n    lang\\n    passedTestCaseCnt\\n    totalTestCaseCnt\\n    sourceUrl\\n    question {\\n      titleSlug\\n      title\\n      translatedTitle\\n      questionId\\n      __typename\\n    }\\n    ... on GeneralSubmissionNode {\\n      outputDetail {\\n        codeOutput\\n        expectedOutput\\n        input\\n        compileError\\n        runtimeError\\n        lastTestcase\\n        __typename\\n      }\\n      __typename\\n    }\\n    submissionComment {\\n      comment\\n      flagType\\n      __typename\\n    }\\n    __typename\\n  }\\n}\\n\"}"
	query = fmt.Sprintf(query, submitID)
	req := u.geneGraphQLRequest(strings.NewReader(query))
	resp, err := u.client.Do(req)
	if err != nil {
		log.Fatalf("Get Submit Detail Failed %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	data = gjson.GetBytes(body, "data.submissionDetail.code").String()
	return

	//todo runtimeDistributionFormatted what best leetcode
	// re := regexp.MustCompile("runtimeDistributionFormatted:\\s('[^']+')")

	//todo memoryDistributionFormatted waat best leetcode
	// re := regexp.MustCompile("memoryDistributionFormatted:\\s('[^']+')")
}
