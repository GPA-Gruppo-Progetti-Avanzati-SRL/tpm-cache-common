package cacheoperation

import (
	"context"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-cache-common/cachelks"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-cache-common/cachelksregistry"
	"github.com/rs/zerolog/log"
)

const (
	SemLogCacheKey = "cache-key"
)

func Set(linkedServiceRef cachelks.CacheLinkedServiceRef, cacheKey string, v interface{}, opts ...cachelks.CacheOption) error {
	const semLogContext = "cache-operation::set"
	var err error

	var options cachelks.CacheOptions
	for _, o := range opts {
		o(&options)
	}

	lks, err := cachelksregistry.GetLinkedServiceOfType(linkedServiceRef.Typ, linkedServiceRef.Name)
	if err != nil {
		return err
	}

	err = lks.Set(context.Background(), cacheKey, v, options)
	if err != nil {
		return err
	}

	log.Trace().Str(SemLogCacheKey, cacheKey).Msg(semLogContext)
	return nil
}
