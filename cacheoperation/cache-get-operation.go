package cacheoperation

import (
	"context"
	"fmt"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-cache-common/cachelks"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-cache-common/cachelksregistry"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-http-archive/har"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
	"time"
)

func Get(linkedServiceRef cachelks.CacheLinkedServiceRef, id string, cacheKey string, contentType string, opts ...cachelks.CacheOption) (*har.Entry, error) {
	const semLogContext = "cache-operation::get"

	var err error
	var options cachelks.CacheOptions
	for _, o := range opts {
		o(&options)
	}

	lks, err := cachelksregistry.GetLinkedServiceOfType(linkedServiceRef.Typ, linkedServiceRef.Name)
	if err != nil {
		return nil, err
	}

	req, err := newGetRequestDefinition(lks, id, []byte(cacheKey), contentType, options)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	harEntry := &har.Entry{
		Comment:         id,
		StartedDateTime: now.Format("2006-01-02T15:04:05.999999999Z07:00"),
		StartDateTimeTm: now,
		Request:         req,
	}

	v, err := lks.Get(context.Background(), cacheKey, options)
	elapsed := time.Since(now)
	if err != nil {
		log.Error().Err(err).Str(SemLogCacheKey, cacheKey).Dur("elapsed", elapsed).Msg(semLogContext)
		harEntry, _ = newGetResponseDefinition(harEntry, http.StatusNotFound, []byte(err.Error()), "text/plain", elapsed)
		return harEntry, err
	}

	if v != nil {
		log.Trace().Str(SemLogCacheKey, cacheKey).Msg(semLogContext + " cache hit")
		switch typedVal := v.(type) {
		case string:
			harEntry, _ = newGetResponseDefinition(harEntry, http.StatusOK, []byte(typedVal), contentType, elapsed)
		case []byte:
			harEntry, _ = newGetResponseDefinition(harEntry, http.StatusOK, typedVal, contentType, elapsed)
		default:
			err = fmt.Errorf("cache key %s resolves to %T", cacheKey, v)
			log.Error().Err(err).Msg(semLogContext)
			harEntry, _ = newGetResponseDefinition(harEntry, http.StatusUnsupportedMediaType, []byte(err.Error()), "text/plain", elapsed)
		}
	} else {
		log.Warn().Str(SemLogCacheKey, cacheKey).Msg(semLogContext + " cache miss")
		harEntry, err = newGetResponseDefinition(harEntry, http.StatusNotFound, []byte("cache miss"), "text/plain", elapsed)
	}

	return harEntry, nil
}

func newGetRequestDefinition(lks cachelks.LinkedService, id string, cacheKey []byte, contentType string, options cachelks.CacheOptions) (*har.Request, error) {

	var opts []har.RequestOption

	var pathBuilder strings.Builder
	pathBuilder.WriteString("/")
	pathBuilder.WriteString(id)
	if options.Namespace != "" {
		pathBuilder.WriteString("/")
		pathBuilder.WriteString(options.Namespace)
	}

	opts = append(opts, har.WithMethod("POST"))
	opts = append(opts, har.WithUrl(lks.Url(pathBuilder.String())))
	opts = append(opts, har.WithBody(cacheKey))

	req := har.Request{
		HTTPVersion: "1.1",
		Cookies:     []har.Cookie{},
		QueryString: []har.NameValuePair{},
		HeadersSize: -1,
		Headers:     []har.NameValuePair{},
		BodySize:    -1,
	}
	for _, o := range opts {
		o(&req)
	}

	return &req, nil
}

func newGetResponseDefinition(harEntry *har.Entry, sc int, resp []byte, contentType string, elapsed time.Duration) (*har.Entry, error) {

	harEntry.Time = float64(elapsed.Milliseconds())
	harEntry.Timings = &har.Timings{
		Blocked: -1,
		DNS:     -1,
		Connect: -1,
		Send:    -1,
		Wait:    harEntry.Time,
		Receive: -1,
		Ssl:     -1,
	}

	r := &har.Response{
		Status:      sc,
		HTTPVersion: "1.1",
		StatusText:  http.StatusText(sc),
		HeadersSize: -1,
		BodySize:    int64(len(resp)),
		Cookies:     []har.Cookie{},
		Headers:     []har.NameValuePair{},
		Content: &har.Content{
			MimeType: contentType,
			Size:     int64(len(resp)),
			Data:     resp,
		},
	}

	harEntry.Response = r
	return harEntry, nil
}
