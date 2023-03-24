package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type LetterContent struct {
	Recipient           string
	RecipientStreet     string
	RecipientPostalCode string
	RecipientCity       string
	Sender              string
	SenderStreet        string
	SenderPostalCode    string
	SenderCity          string
	Date                string
	Opening             string
	Closing             string
	Content             string
	SignatureSpace      bool
}

type LetterError struct {
	Error string
}

func (l *LetterContent) Sanitize() {
	l.Recipient = strings.ReplaceAll(l.Recipient, "%", "")
	l.Recipient = strings.ReplaceAll(l.Recipient, "&", "")
	l.Recipient = strings.ReplaceAll(l.Recipient, "{", "")
	l.Recipient = strings.ReplaceAll(l.Recipient, "}", "")
	l.Recipient = strings.ReplaceAll(l.Recipient, "\\", "")

	l.RecipientStreet = strings.ReplaceAll(l.RecipientStreet, "%", "")
	l.RecipientStreet = strings.ReplaceAll(l.RecipientStreet, "&", "")
	l.RecipientStreet = strings.ReplaceAll(l.RecipientStreet, "{", "")
	l.RecipientStreet = strings.ReplaceAll(l.RecipientStreet, "}", "")
	l.RecipientStreet = strings.ReplaceAll(l.RecipientStreet, "\\", "")

	l.RecipientPostalCode = strings.ReplaceAll(l.RecipientPostalCode, "%", "")
	l.RecipientPostalCode = strings.ReplaceAll(l.RecipientPostalCode, "&", "")
	l.RecipientPostalCode = strings.ReplaceAll(l.RecipientPostalCode, "{", "")
	l.RecipientPostalCode = strings.ReplaceAll(l.RecipientPostalCode, "}", "")
	l.RecipientPostalCode = strings.ReplaceAll(l.RecipientPostalCode, "\\", "")

	l.RecipientCity = strings.ReplaceAll(l.RecipientCity, "%", "")
	l.RecipientCity = strings.ReplaceAll(l.RecipientCity, "&", "")
	l.RecipientCity = strings.ReplaceAll(l.RecipientCity, "{", "")
	l.RecipientCity = strings.ReplaceAll(l.RecipientCity, "}", "")
	l.RecipientCity = strings.ReplaceAll(l.RecipientCity, "\\", "")

	l.Sender = strings.ReplaceAll(l.Sender, "%", "")
	l.Sender = strings.ReplaceAll(l.Sender, "&", "")
	l.Sender = strings.ReplaceAll(l.Sender, "{", "")
	l.Sender = strings.ReplaceAll(l.Sender, "}", "")
	l.Sender = strings.ReplaceAll(l.Sender, "\\", "")

	l.SenderStreet = strings.ReplaceAll(l.SenderStreet, "%", "")
	l.SenderStreet = strings.ReplaceAll(l.SenderStreet, "&", "")
	l.SenderStreet = strings.ReplaceAll(l.SenderStreet, "{", "")
	l.SenderStreet = strings.ReplaceAll(l.SenderStreet, "}", "")
	l.SenderStreet = strings.ReplaceAll(l.SenderStreet, "\\", "")

	l.SenderPostalCode = strings.ReplaceAll(l.SenderPostalCode, "%", "")
	l.SenderPostalCode = strings.ReplaceAll(l.SenderPostalCode, "&", "")
	l.SenderPostalCode = strings.ReplaceAll(l.SenderPostalCode, "{", "")
	l.SenderPostalCode = strings.ReplaceAll(l.SenderPostalCode, "}", "")
	l.SenderPostalCode = strings.ReplaceAll(l.SenderPostalCode, "\\", "")

	l.SenderCity = strings.ReplaceAll(l.SenderCity, "%", "")
	l.SenderCity = strings.ReplaceAll(l.SenderCity, "&", "")
	l.SenderCity = strings.ReplaceAll(l.SenderCity, "{", "")
	l.SenderCity = strings.ReplaceAll(l.SenderCity, "}", "")
	l.SenderCity = strings.ReplaceAll(l.SenderCity, "\\", "")

	l.Date = strings.ReplaceAll(l.Date, "%", "")
	l.Date = strings.ReplaceAll(l.Date, "&", "")
	l.Date = strings.ReplaceAll(l.Date, "{", "")
	l.Date = strings.ReplaceAll(l.Date, "}", "")
	l.Date = strings.ReplaceAll(l.Date, "\\", "")

	l.Opening = strings.ReplaceAll(l.Opening, "%", "")
	l.Opening = strings.ReplaceAll(l.Opening, "&", "")
	l.Opening = strings.ReplaceAll(l.Opening, "{", "")
	l.Opening = strings.ReplaceAll(l.Opening, "}", "")
	l.Opening = strings.ReplaceAll(l.Opening, "\\", "")

	l.Closing = strings.ReplaceAll(l.Closing, "%", "")
	l.Closing = strings.ReplaceAll(l.Closing, "&", "")
	l.Closing = strings.ReplaceAll(l.Closing, "{", "")
	l.Closing = strings.ReplaceAll(l.Closing, "}", "")
	l.Closing = strings.ReplaceAll(l.Closing, "\\", "")

	l.Content = strings.ReplaceAll(l.Content, "%", "")
	l.Content = strings.ReplaceAll(l.Content, "&", "")
	l.Content = strings.ReplaceAll(l.Content, "{", "")
	l.Content = strings.ReplaceAll(l.Content, "}", "")
	l.Content = strings.ReplaceAll(l.Content, "\\", "")
}

func pdfLatex(l *LetterContent) ([]byte, error) {
	dirName := uuid.New().String()
	err := os.Mkdir(fmt.Sprintf("./%s", dirName), 0755)
	if err != nil {
		log.Error().Msg(err.Error())
		return nil, err
	}
	defer os.RemoveAll(fmt.Sprintf("./%s", dirName))

	baseFileName := "letter-de"
	fileName := fmt.Sprintf("%s.tex", baseFileName)
	file, err := os.Create(fmt.Sprintf("./%s/%s", dirName, fileName))
	if err != nil {
		log.Error().Msg(err.Error())
		return nil, err
	} else {
		defer file.Close()
	}

	t, err := template.ParseFiles("templates/letter-de.tex.tmpl")
	if err != nil {
		log.Error().Msg(err.Error())
		return nil, err
	}

	err = t.Execute(file, l)
	if err != nil {
		log.Error().Msg(err.Error())
		return nil, err
	}

	log.Info().Msgf("Successfully rendered tex file at ./%s/%s", dirName, fileName)

	cmnd := exec.Command("pdflatex",
		fmt.Sprintf("-output-directory=./%s", dirName),
		// fmt.Sprintf("-aux-directory=%s", dirName),
		"-synctex=1", "-no-shell-escape", "-interaction=nonstopmode",
		fmt.Sprintf("./%s/%s", dirName, baseFileName))

    var errb bytes.Buffer
    // cmnd.Stdout = &outb
    cmnd.Stderr = &errb

	err = cmnd.Run()
	if err != nil {
		log.Error().Msgf("%s; %s", err.Error(), errb.String())
		return nil, err
	}

	log.Info().Msgf("Successfully rendered pdf file at %s/%s.pdf", dirName, baseFileName)

	fileBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s.pdf", dirName, baseFileName))
	if err != nil {
		log.Error().Msg(err.Error())
		return nil, err
	}

	return fileBytes, nil
}

var months = [...]string{
	"Jänner", "Februar", "März", "April", "Mai", "Juni",
	"Juli", "August", "September", "Oktober", "November", "Dezember",
}

func GermanDate(t time.Time) string {
	return fmt.Sprintf("%d %s %d", t.Day(), months[t.Month()-1], t.Year())
}

func ReturnError(error string, w http.ResponseWriter) {
	t, err := template.ParseFiles("templates/error.html.tmpl")
	if err != nil {
		log.Error().Msg(err.Error())
		error = fmt.Sprintf("%s\n%s", err, error)
	}

	e := LetterError{
		Error: error,
	}

	err = t.Execute(w, e)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
}

func handleForm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Msgf(r.FormValue(err.Error()))
		ReturnError(err.Error(), w)
		return
	}

	recipient := r.PostFormValue("recipient")
	recipientStreet := r.PostFormValue("recipient_street")
	recipientPostalCode := r.PostFormValue("recipient_postal_code")
	recipientCity := r.PostFormValue("recipient_city")
	sender := r.PostFormValue("sender")
	senderStreet := r.PostFormValue("sender_street")
	senderPostalCode := r.PostFormValue("sender_postal_code")
	senderCity := r.PostFormValue("sender_city")
	date := r.PostFormValue("date")
	opening := r.PostFormValue("opening")
	closing := r.PostFormValue("closing")
	content := r.PostFormValue("content")

	signatureSpace := false
	if r.PostFormValue("signature_space") == "on" {
		signatureSpace = true
	}

	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Msgf(err.Error())
		ReturnError(err.Error(), w)
		return
	}

	l := LetterContent{
		Recipient:           recipient,
		RecipientStreet:     recipientStreet,
		RecipientPostalCode: recipientPostalCode,
		RecipientCity:       recipientCity,
		Sender:              sender,
		SenderStreet:        senderStreet,
		SenderPostalCode:    senderPostalCode,
		SenderCity:          senderCity,
		Date:                GermanDate(d),
		Opening:             opening,
		Closing:             closing,
		Content:             content,
		SignatureSpace:      signatureSpace,
	}

	l.Sanitize()
	fileBytes, err := pdfLatex(&l)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ReturnError(err.Error(), w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(fileBytes)
}

func main() {
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)
	http.HandleFunc("/generate", handleForm)

	log.Info().Msgf("Listening on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Error().Msgf(err.Error())
	}
}
