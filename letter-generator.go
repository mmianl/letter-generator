package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/aquilax/truncate"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

var Version = "development"
var (
	apiRequestsInFlightGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "letter_generator_http_requests_in_flight",
			Help: "Number of concurrent HTTP api requests currently handled.",
		},
	)
	apiRequestsTotalCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "letter_generator_http_requests_total",
			Help: "Total number of api requests.",
		},
		[]string{"code", "method"},
	)
	apiResponseSizeSummary = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "letter_generator_http_response_size_bytes",
			Help: "Api HTTP response size in bytes.",
		},
		[]string{},
	)
	apiRequestsDurationSummary = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "letter_generator_http_request_duration_seconds",
			Help: "Duration of api requests in seconds.",
		},
		[]string{"handler", "method"},
	)
	apiVersionGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "letter_generator_build_info",
			Help: "Metric with a constant '1' value labeled by version and goversion from which letter_generator was built.",
		},
		[]string{"version", "goversion"},
	)
	apiLettersGeneratedCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "letter_generator_generated_total",
			Help: "Total number of generated letters.",
		},
		[]string{},
	)
	apiLettersGeneratedFailedCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "letter_generator_failed_total",
			Help: "Total number of failed letters.",
		},
		[]string{},
	)
)

type LetterContent struct {
	Subject             string
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
	replacer := strings.NewReplacer("%", " ",
		"&", " ",
		"{", " ",
		"}", " ",
		"\\", " ",
		"<", " ",
		">", " ",
		"_", " ",
		"\"", "“",
		"'", "‘")

	l.Recipient = truncate.Truncate(replacer.Replace(l.Recipient), 200, "...", truncate.PositionEnd)
	l.RecipientStreet = truncate.Truncate(replacer.Replace(l.RecipientStreet), 200, "...", truncate.PositionEnd)
	l.RecipientPostalCode = truncate.Truncate(replacer.Replace(l.RecipientPostalCode), 200, "...", truncate.PositionEnd)
	l.RecipientCity = truncate.Truncate(replacer.Replace(l.RecipientCity), 200, "...", truncate.PositionEnd)
	l.Sender = truncate.Truncate(replacer.Replace(l.Sender), 200, "...", truncate.PositionEnd)
	l.SenderStreet = truncate.Truncate(replacer.Replace(l.SenderStreet), 200, "...", truncate.PositionEnd)
	l.SenderPostalCode = truncate.Truncate(replacer.Replace(l.SenderPostalCode), 200, "...", truncate.PositionEnd)
	l.SenderCity = truncate.Truncate(replacer.Replace(l.SenderCity), 200, "...", truncate.PositionEnd)
	l.Date = truncate.Truncate(replacer.Replace(l.Date), 200, "...", truncate.PositionEnd)
	l.Opening = truncate.Truncate(replacer.Replace(l.Opening), 200, "...", truncate.PositionEnd)
	l.Opening = truncate.Truncate(replacer.Replace(l.Opening), 200, "...", truncate.PositionEnd)
	l.Closing = truncate.Truncate(replacer.Replace(l.Closing), 200, "...", truncate.PositionEnd)
	l.Content = truncate.Truncate(replacer.Replace(l.Content), 10000, "...", truncate.PositionEnd)
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

	for i := 1; i <= 2; i++ {
		cmnd := exec.Command("pdflatex",
			fmt.Sprintf("-output-directory=./%s", dirName),
			"-synctex=1", "-no-shell-escape", "-interaction=nonstopmode",
			fmt.Sprintf("./%s/%s", dirName, baseFileName))

		var errb, erro bytes.Buffer
		cmnd.Stderr = &errb
		cmnd.Stdout = &erro

		err = cmnd.Run()
		if err != nil {
			log.Error().Msgf("%s; %s; %s", err.Error(), errb.String(), erro.String())
			return nil, err
		}

		log.Info().Msgf("Successfully ran pdflatex for the %dth time at %s.", i, dirName)
	}

	log.Info().Msgf("Successfully rendered pdf file at %s/%s.pdf", dirName, baseFileName)

	fileBytes, err := os.ReadFile(fmt.Sprintf("%s/%s.pdf", dirName, baseFileName))
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
	return fmt.Sprintf("%d. %s %d", t.Day(), months[t.Month()-1], t.Year())
}

func ReturnError(error string, w http.ResponseWriter) {
	apiLettersGeneratedFailedCounter.WithLabelValues().Inc()

	w.WriteHeader(http.StatusInternalServerError)
	log.Error().Msgf(error)

	t, err := template.ParseFiles("templates/error.html.tmpl")
	if err != nil {
		log.Error().Msg(err.Error())
		error = fmt.Sprintf("%s\n%s", err.Error(), error)
		if _, err := w.Write([]byte(error)); err != nil {
			log.Panic().Msg(err.Error())
		}
		return
	}

	e := LetterError{
		Error: error,
	}

	err = t.Execute(w, e)
	if err != nil {
		log.Error().Msg(err.Error())
		error = fmt.Sprintf("%s\n%s", err.Error(), error)
		if _, err := w.Write([]byte(error)); err != nil {
			log.Panic().Msg(err.Error())
		}
		return
	}
}

var formHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	apiLettersGeneratedCounter.WithLabelValues().Inc()

	err := r.ParseForm()
	if err != nil {
		ReturnError(err.Error(), w)
		return
	}

	subject := r.PostFormValue("subject")
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
		ReturnError(err.Error(), w)
		return
	}

	l := LetterContent{
		Subject:             subject,
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
		ReturnError(err.Error(), w)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=letter.pdf")

	_, err = w.Write(fileBytes)
	if err != nil {
		ReturnError(err.Error(), w)
		return
	}

	w.WriteHeader(http.StatusOK)
})

func main() {
	fs := http.FileServer(http.Dir("./web"))

	rootChain := promhttp.InstrumentHandlerInFlight(apiRequestsInFlightGauge,
		promhttp.InstrumentHandlerDuration(apiRequestsDurationSummary.MustCurryWith(prometheus.Labels{"handler": "/"}),
			promhttp.InstrumentHandlerCounter(apiRequestsTotalCounter,
				promhttp.InstrumentHandlerResponseSize(apiResponseSizeSummary, fs),
			),
		),
	)
	generateChain := promhttp.InstrumentHandlerInFlight(apiRequestsInFlightGauge,
		promhttp.InstrumentHandlerDuration(apiRequestsDurationSummary.MustCurryWith(prometheus.Labels{"handler": "/generate"}),
			promhttp.InstrumentHandlerCounter(apiRequestsTotalCounter,
				promhttp.InstrumentHandlerResponseSize(apiResponseSizeSummary, formHandler),
			),
		),
	)

	apiVersionGauge.WithLabelValues(Version, runtime.Version()).Set(1)

	http.Handle("/", rootChain)
	http.Handle("/generate", generateChain)

	log.Info().Msgf("Running letter-generator version %s", Version)
	go func() {
		err := http.ListenAndServe(":8081", promhttp.Handler())
		if err != nil {
			log.Fatal().Msgf(err.Error())
		}
	}()

	log.Info().Msgf("Listening on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal().Msgf(err.Error())
	}
}
