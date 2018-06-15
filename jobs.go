package perfm

import (
	"bytes"
	"crypto/tls"
	"errors"
	"net/http"
)

var ErrReturn error = errors.New("error in return")

type HTTPJob struct {
	client http.Client
	req    http.Request
}

func NewHTTPJob(url, method, body string, t *tls.Config) Job {
	req, _ := http.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
	return &HTTPJob{client: http.Client{}, req: *req}
}

func (j *HTTPJob) Do() error {
	resp, err := j.client.Do(&j.req)
	if err != nil {
		return err
	}
	if resp.StatusCode > 400 {
		return ErrReturn
	}
	return nil
}
func (j *HTTPJob) Copy() (Job, error) {
	jb := *j
	jb.client = http.Client{}
	return &jb, nil
}
func (j *HTTPJob) Pre() error {
	return nil
}
func (j *HTTPJob) After() {
	return
}
