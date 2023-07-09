package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"runtime"

	"github.com/sqweek/dialog"
)

func getFileNames(files []*multipart.FileHeader) []string {
	filenames := make([]string, len(files))
	for i, file := range files {
		filenames[i] = file.Filename
	}
	return filenames
}

type DialogData struct {
	message    string
	resultChan chan bool
}

func main() {
	var dialogChan = make(chan DialogData)

	go func() {

	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "upload.html")
	})

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		// Check if a POST request is made
		if r.Method == http.MethodPost {
			r.ParseMultipartForm(10 << 20)

			files := r.MultipartForm.File["myFiles"]
			names := getFileNames(files)
			log.Printf("%s want to upload %s", r.RemoteAddr, names)

			resultChan := make(chan bool)
			dialogChan <- DialogData{
				message:    fmt.Sprintf("%s wants to upload files %s. Do you want to accept?", r.RemoteAddr, names),
				resultChan: resultChan,
			}

			confirm := <-resultChan

			if !confirm {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				log.Printf("Rejected to upload files")
				return
			}

			for _, file := range files {
				log.Printf("%s receiving file %s", r.RemoteAddr, file.Filename)

				src, err := file.Open()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer src.Close()

				dst, err := os.Create(file.Filename)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer dst.Close()

				if _, err := io.Copy(dst, src); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			fmt.Fprintf(w, "Files uploaded successfully!")
		}
	})

	go func() {
		log.Printf("Serving on :8765")
		http.ListenAndServe(":8765", nil)
	}()

	for data := range dialogChan {
		runtime.LockOSThread()
		log.Printf("UI thread received message: %s", data.message)
		confirm := dialog.Message(data.message).Title("Upload Confirmation").YesNo()
		runtime.UnlockOSThread()
		data.resultChan <- confirm

	}
}
