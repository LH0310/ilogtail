package sql

import (
	"errors"
	"strings"

	"github.com/alibaba/ilogtail/pkg/models"
)

func toStringLogContents(logContents models.LogContents) (stringLogContents, error) {
	slc := newStringLogContents()
	for key, value := range logContents.Iterator() {
		value, ok := value.(string)
		if !ok {
			return nil, errors.New("not stringLogContents")
		}
		slc.Add(key, value)
	}
	return slc, nil
}

func newStringLogContents() stringLogContents {
	return models.NewKeyValues[string]()
}

func SQLLikeToRegexp(sqlLike string) string {
	regexpLike := strings.ReplaceAll(sqlLike, "%", ".*")
	regexpLike = strings.ReplaceAll(regexpLike, "_", ".")
	return "^" + regexpLike + "$"
}
