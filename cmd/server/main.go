package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/sqweek/dialog"
)

func getFileNames(files []*multipart.FileHeader) []string {
	filenames := make([]string, len(files))
	for i, file := range files {
		filenames[i] = file.Filename
	}
	return filenames
}

func main() {
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

			// Display a dialog asking for confirmation before saving files
			confirm := dialog.Message("%s wants to upload files %s. Do you want to accept?", r.RemoteAddr, names).
				Title("Upload Confirmation").
				YesNo()

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

	log.Printf("Serving on :8765")
	http.ListenAndServe(":8765", nil)
}
