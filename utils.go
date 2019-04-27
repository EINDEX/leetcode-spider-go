package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"
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
