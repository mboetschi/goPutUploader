// Copyright 2024 LiveKit, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type PUTUploader struct {
	endpoint string
	timeout  time.Duration
}

func NewPUTUploader(endpoint string, timeout time.Duration) (*PUTUploader, error) {
	return &PUTUploader{
		endpoint: endpoint,
		timeout:  timeout,
	}, nil
}

func (u *PUTUploader) upload(localFilepath, storageFilepath string, _ OutputType) (string, int64, error) {
	file, err := os.Open(localFilepath)
	if err != nil {
		return "", 0, wrap("HTTP", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("File", filepath.Base(localFilepath))
	if err != nil {
		panic(err)
	}
	io.Copy(part, file)

	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)
	writer.WriteField("timestamp", timestampStr)

	writer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), u.timeout)
	defer cancel()

	requestURL := fmt.Sprintf("%s/%s", u.endpoint, storageFilepath)
	req, err := http.NewRequestWithContext(ctx, "PUT", requestURL, body)
	if err != nil {
		return "", 0, wrap("HTTP", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.ContentLength = int64(body.Len()) // Set the content length for the request

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, wrap("HTTP", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("HTTP upload failed with status code: %d", resp.StatusCode)
	}

	return requestURL, int64(body.Len()), nil
}

func wrap(name string, err error) error {
	return errors.Wrap(err, fmt.Sprintf("%s upload failed", name))
}
