package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/GoWebProd/multipart"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

func createTLSConnect(certFile string, keyFile string, timeout time.Duration) (*http.Client, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, errors.Wrap(err, "Can't load cert and key")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	return client, nil
}

func main() {
	fileURL := "https://secure.eicar.org/eicar.com.txt"
	fileReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fileURL, nil)
	if err != nil {
		panic(err)
	}

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	fileResp, err := httpClient.Do(fileReq)
	if err != nil {
		panic(err)
	}

	defer fileResp.Body.Close()

	client, err := createTLSConnect("/home/ssl/kata.pem", "/home/ssl/kata.key", 5*time.Second)
	//if err != nil {
	//	panic(err)
	//}

	writer := multipart.NewWriter()

	err = writer.CreateFormFileReader("content", "very_important_file_name", multipart.NewReader(fileResp.Body, int(fileResp.ContentLength)))
	if err != nil {
		panic(err)
	}

	scanID, _ := uuid.NewUUID()
	err = writer.CreateFormField("scanId", []byte(scanID.String()))
	if err != nil {
		panic(err)
	}

	err = writer.CreateFormField("objectType", []byte("file"))
	if err != nil {
		panic(err)
	}

	PostAPIPath := "https://cn.kata.im-sandbox.devmail.ru/kata/scanner/v1/sensors/06c834a7-5e6d-4d61-b8af-4ff3ff415dc0/scans"

	var buf bytes.Buffer

	_, err = io.Copy(&buf, writer)

	u, err := url.Parse(PostAPIPath)
	if err != nil {
		panic(err)
	}
	//req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, PostAPIPath, nil)
	req := &http.Request{
		Method:     http.MethodPost,
		URL:        u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Body:       writer,
		Host:       u.Host,
	}

	req.Body = writer
	req.ContentLength = int64(writer.Len())
	req.GetBody = func() (io.ReadCloser, error) {
		return writer, nil
	}

	//req.Body = ioutil.NopCloser(writer)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	fmt.Println(resp.StatusCode)
}
