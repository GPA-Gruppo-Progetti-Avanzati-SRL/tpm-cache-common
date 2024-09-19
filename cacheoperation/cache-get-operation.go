package cacheoperation

import (
	"context"
	"fmt"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-cache-common/cachelks"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-cache-common/cachelksregistry"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-http-archive/har"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

func Get(id string, ns, cacheKey string, contentType string, linkedServiceRef cachelks.CacheLinkedServiceRef) (*har.Entry, error) {
	const semLogContext = "cache-operation::get"

	var err error

	lks, err := cachelksregistry.GetLinkedServiceOfType(linkedServiceRef.Typ, linkedServiceRef.Name)
	if err != nil {
		return nil, err
	}

	req, err := newRequestDefinition(lks, []byte(cacheKey), contentType)
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

	v, err := lks.Get(context.Background(), cacheKey)
	elapsed := time.Since(now)
	if err != nil {
		log.Error().Err(err).Str(SemLogCacheKey, cacheKey).Dur("elapsed", elapsed).Msg(semLogContext)
		harEntry, _ = newResponseDefinition(harEntry, http.StatusNotFound, []byte(err.Error()), "text/plain", elapsed)
		return harEntry, err
	}

	if v != nil {
		if b, ok := v.(string); ok {
			log.Trace().Str(SemLogCacheKey, cacheKey).Msg(semLogContext + " cache hit")
			harEntry, _ = newResponseDefinition(harEntry, http.StatusOK, []byte(b), contentType, elapsed)
			return harEntry, nil
		}

		err = fmt.Errorf("cache key %s resolves to %T", cacheKey, v)
		log.Error().Err(err).Msg(semLogContext)
		harEntry, _ = newResponseDefinition(harEntry, http.StatusUnsupportedMediaType, []byte(err.Error()), "text/plain", elapsed)
	} else {
		log.Warn().Str(SemLogCacheKey, cacheKey).Msg(semLogContext + " cache miss")
		harEntry, err = newResponseDefinition(harEntry, http.StatusNotFound, []byte("cache miss"), "text/plain", elapsed)
	}

	return harEntry, nil
}

func newRequestDefinition(lks cachelks.LinkedService, cacheKey []byte, contentType string) (*har.Request, error) {
	var opts []har.RequestOption

	ub := har.UrlBuilder{}
	ub.WithPort(80)
	ub.WithScheme("http")
	ub.WithHostname("localhost")
	ub.WithPath("myPath")

	opts = append(opts, har.WithMethod("POST"))
	opts = append(opts, har.WithUrl(ub.Url()))
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

func newResponseDefinition(harEntry *har.Entry, sc int, resp []byte, contentType string, elapsed time.Duration) (*har.Entry, error) {

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
