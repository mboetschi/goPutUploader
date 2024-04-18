package main

import (
	"fmt"
	"time"
)

func main() {

	uploader, err := NewPUTUploader("https://localhost:34567/api/v1/recordings/livevideo", 5*time.Second)
	if err != nil {
		fmt.Errorf("Creating PutUploder:", err)
	}
	_, _, err = uploader.upload("sample-8GB.mp4", "test6", OutputTypeMP4)
	if err != nil {
		fmt.Printf("Uploading:", err)
	}

}
