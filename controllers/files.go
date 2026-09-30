package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/example/placeholder-webapp/context"
	"github.com/example/placeholder-webapp/models"
	"github.com/go-chi/chi/v5"
)

type Files struct {
	Templates struct {
		List Template
	}
	FilesService *models.FileService
}

func (f Files) List(w http.ResponseWriter, r *http.Request) {
	var data struct{ FullName string }
	if user := context.User(r.Context()); user != nil {
		data.FullName = user.FullName
	}

	f.Templates.List.Execute(w, r, data)
}

func (f Files) DownloadGeneral(w http.ResponseWriter, r *http.Request) {
	filename := f.filename(w, r)

	file, err := f.FilesService.GetGeneral(filename)
	if err != nil {
		log.Printf("failed to get general file %q: %v", filename, err)
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	http.ServeFile(w, r, file.Path)
}

func (f Files) DownloadSafety(w http.ResponseWriter, r *http.Request) {
	filename := f.filename(w, r)

	file, err := f.FilesService.GetSafety(filename)
	if err != nil {
		log.Printf("failed to get safety file %q: %v", filename, err)
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	http.ServeFile(w, r, file.Path)
}

func (f Files) DownloadSchedule(w http.ResponseWriter, r *http.Request) {
	filename := f.filename(w, r)

	user := context.User(r.Context())
	if user == nil {
		http.Error(w, "User not found", http.StatusForbidden)
		return
	}

	file, err := f.FilesService.GetSchedule(user.SFID, filename)
	if err != nil {
		log.Printf("failed to get schedule file %q for %q: %v", filename, user.SFID, err)
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		if errors.Is(err, models.ErrUnauthorized) {
			http.Error(w, "Not authorized to access this file", http.StatusUnauthorized)
			return
		}

		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	http.ServeFile(w, r, file.Path)
}

func (f Files) filename(_ http.ResponseWriter, r *http.Request) string {
	filename := chi.URLParam(r, "filename")
	if unescaped, err := url.PathUnescape(filename); err == nil {
		filename = unescaped
	}
	filename = filepath.Base(filename)
	return filename
}

func byteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
