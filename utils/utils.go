package utils

import (
	"bytes"
	"fmt"
	"io"
	"leetcode-spider-go/settings"
	"mime/multipart"
	"os"
	"os/exec"
)

func GetMultipartForm(values map[string]io.Reader) (buffer bytes.Buffer, header string, err error) {
	w := multipart.NewWriter(&buffer)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return
		}

	}
	header = w.FormDataContentType()
	w.Close()
	return
}

func GetLangSuffix(lang string) string {
	var suffix string
	switch lang {
	case "python3", "python":
		suffix = "py"
	case "go":
		suffix = "go"
	case "mysql":
		suffix = "sql"
	case "c++":
		suffix = "cpp"
	case "c":
		suffix = "c"
	case "java":
		suffix = "java"
	case "JavaScript":
		suffix = "js"
	}
	return suffix
}

func ExecCommend(name string, args ...string) (err error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = settings.Setting.Out
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
	return
}
