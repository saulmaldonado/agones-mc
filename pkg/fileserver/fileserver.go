package fileserver

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
)

const MaxFileSize int64 = 5 << 10

// Wrapper for http.ResponseWritter for intercepting
// the written body buffer and headers before being sent back
type appResWriter struct {
	http.ResponseWriter
	buf        *bytes.Buffer
	statusCode int
}

// Overwrites the http.ReponseWriter.Write method to write to a local buffer
// buffer can then be used to intercept the written response body
func (rw *appResWriter) Write(b []byte) (int, error) {
	return rw.buf.Write(b)
}

// Overwrites the http.ReponseWriter.WriteHeader method to write to the status code to a local field
// writing to statusCode can then be delayed or overwritten
func (rw *appResWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

func GetFile(rw http.ResponseWriter, r *http.Request) error {

	// ResponseWriter wrapper
	dummyRw := appResWriter{
		ResponseWriter: rw,
		buf:            &bytes.Buffer{},
	}

	http.FileServer(http.Dir(".")).ServeHTTP(&dummyRw, r)

	if dummyRw.statusCode == 204 || dummyRw.statusCode >= 300 {
		rw.WriteHeader(dummyRw.statusCode)
		return nil
	}

	if rw.Header().Get("Accept-Ranges") != "bytes" {
		rw.Header().Set("Content-Type", "application/json")
		rw.Header().Del("Last-Modified")

		files := []map[string]string{}

		scanner := bufio.NewScanner(dummyRw.buf)
		for scanner.Scan() {

			text := scanner.Text()

			reg := regexp.MustCompile(`^<a href=\"([a-zA-z0-9\-\_\.\/]+)\">([a-zA-z0-9\-\_\.\/]+)<\/a>`)

			if matches := reg.FindStringSubmatch(text); len(matches) > 0 {
				files = append(files, map[string]string{"name": matches[1]})
			}
		}

		if err := json.NewEncoder(rw).Encode(files); err != nil {
			http.Error(rw, "error encoding JSON", http.StatusInternalServerError)
			return err
		}
	}

	_, err := io.Copy(rw, dummyRw.buf)
	return err
}

func UploadFile(rw http.ResponseWriter, r *http.Request, vol string) error {
	r.Body = http.MaxBytesReader(rw, r.Body, MaxFileSize)

	if err := r.ParseMultipartForm(MaxFileSize); err != nil {
		http.Error(rw, err.Error(), http.StatusRequestEntityTooLarge)
		return err
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return err
	}

	defer file.Close()

	f, err := os.Create(path.Join(vol, header.Filename))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return err
	}

	defer f.Close()

	if _, err := io.Copy(f, file); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}

func DeleteFile(rw http.ResponseWriter, r *http.Request, vol string, path string) error {
	if err := os.Remove(path); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}
