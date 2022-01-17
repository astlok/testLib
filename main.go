package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"io"
	"mime/multipart"
	"net/http"
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

	client, err := createTLSConnect("/home/ssl/kata.pem", "/home/ssl/kata.key", 5*time.Second)
	//if err != nil {
	//	panic(err)
	//}

	body, writer := io.Pipe()

	PostAPIPath := "https://cn.kata.im-sandbox.devmail.ru/kata/scanner/v1/sensors/06c834a7-5e6d-4d61-b8af-4ff3ff415dc0/scans"

	mwriter := multipart.NewWriter(writer)

	go func() {
		part, err := mwriter.CreateFormFile("content", "very_important_file_name")
		if err != nil {
			panic(err)
		}

		_, err = io.Copy(part, fileResp.Body)

		//defer fileResp.Body.Close()

		scanID, _ := uuid.NewUUID()
		err = mwriter.WriteField("scanId", scanID.String())
		if err != nil {
			panic(err)
		}

		err = mwriter.WriteField("objectType", "file")
		if err != nil {
			panic(err)
		}

		if err = mwriter.Close(); err != nil {
			panic(err)
		}
		writer.Close()
	}()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, PostAPIPath, body)

	req.ContentLength = 601

	req.Header.Add("Content-Type", mwriter.FormDataContentType())

	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	fmt.Println(resp.StatusCode)
}
