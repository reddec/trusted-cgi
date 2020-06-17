package application

import (
	"bytes"
	"context"
	"fmt"
	"github.com/reddec/trusted-cgi/internal"
	"github.com/reddec/trusted-cgi/stats"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Handler for incoming requests
func (project *Project) Handler(ctx context.Context, tracker stats.Recorder) (http.HandlerFunc, error) {
	return func(writer http.ResponseWriter, request *http.Request) {
		sections := strings.SplitN(request.URL.Path, "/", 2)
		appName := sections[0]
		app := project.FindApp(appName)

		if app == nil {
			http.NotFound(writer, request)
			return
		}

		start := time.Now()
		app.Run(ctx, tracker, project.config.Environment, writer, request)
		end := time.Now()
		log.Println("[INFO]", "("+appName+")", end.Sub(start))
	}, nil
}

// Handler for incoming requests
func (project *Project) HandlerAlias(ctx context.Context, tracker stats.Recorder) (http.HandlerFunc, error) {
	return func(writer http.ResponseWriter, request *http.Request) {
		sections := strings.SplitN(request.URL.Path, "/", 2)
		appName := sections[0]
		app := project.FindAppByAlias(appName)

		if app == nil {
			http.NotFound(writer, request)
			return
		}

		start := time.Now()
		app.Run(ctx, tracker, project.config.Environment, writer, request)
		end := time.Now()
		log.Println("[INFO]", "("+appName+")", end.Sub(start))
	}, nil
}

// Run application with parameters defined in manifest in directory
//
func (app *App) Run(ctx context.Context,
	tracker stats.Recorder,
	env map[string]string,
	w http.ResponseWriter,
	r *http.Request) {
	requestBody := r.Body
	defer requestBody.Close()

	var record = stats.Record{
		UID:    app.UID,
		Method: r.Method,
		Remote: r.RemoteAddr,
		Origin: r.Header.Get("Origin"),
		URI:    r.RequestURI,
		Token:  r.Header.Get("Authorization"),
		Begin:  time.Now(),
	}
	defer func() {
		record.End = time.Now()
		tracker.Track(record)
	}()

	if !app.passSecurityCheck(r) {
		record.Code = http.StatusForbidden
		record.Err = "security checks failed"
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if app.Manifest.Static != "" && (r.Method == http.MethodGet || r.Method == http.MethodHead) {
		dir, err := app.File(app.Manifest.Static)
		if err != nil {
			record.Code = http.StatusForbidden
			record.Err = err.Error()
			w.WriteHeader(http.StatusForbidden)
			return
		}
		prefix := strings.SplitN(r.URL.Path, "/", 2)[0]
		http.StripPrefix(prefix, http.FileServer(http.Dir(dir))).ServeHTTP(w, r)
		return
	}

	if len(app.Manifest.Run) == 0 {
		record.Code = http.StatusNotFound
		record.Err = "run is not defined in manifest"
		http.NotFound(w, r)
		return
	}

	if app.Manifest.Method != "" && r.Method != app.Manifest.Method {
		record.Code = http.StatusMethodNotAllowed
		record.Err = "method not allowed"
		http.Error(w, "method nod allowed", http.StatusMethodNotAllowed)
		return
	}

	if app.Manifest.TimeLimit > 0 {
		cctx, cancel := context.WithTimeout(ctx, time.Duration(app.Manifest.TimeLimit))
		defer cancel()
		ctx = cctx
	}

	var input io.Reader = r.Body

	if app.Manifest.MaximumPayload > 0 {
		input = io.LimitReader(input, app.Manifest.MaximumPayload)
	}

	inputData, err := ioutil.ReadAll(input)
	if err != nil {
		record.Code = http.StatusBadRequest
		record.Err = fmt.Sprintf("failed read input data: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	record.Input = inputData
	r.Body = ioutil.NopCloser(bytes.NewReader(inputData))

	var result bytes.Buffer

	cmd := exec.CommandContext(ctx, app.Manifest.Run[0], app.Manifest.Run[1:]...)
	cmd.Dir = app.location
	cmd.Stdin = bytes.NewReader(inputData)
	cmd.Stdout = &result
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: app.creds,
	}
	internal.SetFlags(cmd)
	var environments = os.Environ()
	for header, mapped := range env {
		environments = append(environments, header+"="+mapped)
	}
	for header, mapped := range app.Manifest.InputHeaders {
		environments = append(environments, mapped+"="+r.Header.Get(header))
	}
	for query, mapped := range app.Manifest.Query {
		environments = append(environments, mapped+"="+r.FormValue(query))
	}
	if app.Manifest.MethodEnv != "" {
		environments = append(environments, app.Manifest.MethodEnv+"="+r.Method)
	}
	if app.Manifest.PathEnv != "" {
		environments = append(environments, app.Manifest.PathEnv+"="+r.URL.Path)
	}
	for k, v := range app.Manifest.Environment {
		environments = append(environments, k+"="+v)
	}
	cmd.Env = environments
	err = cmd.Run()
	record.Output = result.Bytes()
	if err != nil {
		record.Code = http.StatusBadGateway
		record.Err = fmt.Sprintf("run failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	for k, v := range app.Manifest.OutputHeaders {
		w.Header().Set(k, v)
	}
	w.Header().Set("Content-Length", strconv.Itoa(result.Len()))
	w.WriteHeader(http.StatusOK)
	record.Code = http.StatusOK
	_, _ = w.Write(result.Bytes())
}

func (app *App) passSecurityCheck(req *http.Request) bool {
	host, _, _ := net.SplitHostPort(req.RemoteAddr)
	if len(app.Manifest.AllowedIP) > 0 && !app.Manifest.AllowedIP.Has(host) {
		return false
	}
	if len(app.Manifest.AllowedOrigin) > 0 && !app.Manifest.AllowedOrigin.Has(req.Header.Get("Origin")) {
		return false
	}

	if !app.Manifest.Public {
		_, ok := app.Manifest.Tokens[req.Header.Get("Authorization")]
		if !ok {
			return false
		}
	}
	return true
}
