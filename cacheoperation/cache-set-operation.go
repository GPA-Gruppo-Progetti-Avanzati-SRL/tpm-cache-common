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

func Set(ns, cacheKey string, v interface{}, linkedServiceRef cachelks.CacheLinkedServiceRef) error {
	const semLogContext = "cache-operation::set"
	var err error
	lks, err := cachelksregistry.GetLinkedServiceOfType(linkedServiceRef.Typ, linkedServiceRef.Name)
	if err != nil {
		return err
	}

	err = lks.Set(context.Background(), cacheKey, v)
	if err != nil {
		return err
	}

	log.Trace().Str(SemLogCacheKey, cacheKey).Msg(semLogContext)
	return nil
}
