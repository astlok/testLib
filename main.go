package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/GoWebProd/multipart"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"io"
	"net/http"
	"time"
)

func createTLSConnect(certFile string, keyFile string, timeout time.Duration) (*fasthttp.Client, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, errors.Wrap(err, "Can't load cert and key")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	client := &fasthttp.Client{TLSConfig: tlsConfig}

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

	//client, err := createTLSConnect("/home/ssl/kata.pem", "/home/ssl/kata.key", 5*time.Second)
	if err != nil {
		panic(err)
	}

	writer := multipart.NewWriter()

	err = writer.CreateFormFileReader("content", "very_important_file_name", fileResp.Body)
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

	req := fasthttp.AcquireRequest()
	req.SetRequestURI(PostAPIPath)
	req.Header.SetMethod("POST")

	var buf bytes.Buffer

	_, err = io.Copy(&buf, writer)

	req.SetBody(buf.Bytes())

	req.Header.Add("Content-Type", writer.FormDataContentType())

	//resp := fasthttp.AcquireResponse()

	//err = client.Do(req, resp)
	//if err != nil {
	//	panic(err)
	//}

	fmt.Println(string(req.Body()))
}
