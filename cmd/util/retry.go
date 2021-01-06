package util

import (
	"context"
	"errors"
	"time"

	"github.com/leotaku/manki/mangadex"
	"golang.org/x/time/rate"
)

const retryArg = 4

var limiter = rate.NewLimiter(rate.Inf, 1)

func Retry(f func() error) error {
	err := error(nil)
	for i := 0; i <= retryArg; i++ {
		limiter.Wait(context.TODO())
		err = f()
		if err == nil {
			return nil
		} else {
			status := new(mangadex.HttpStatusError)
			errors.As(err, &status)
			if 500 <= status.Status() || status.Status() < 600 {
				limiter.SetLimit(rate.Every(time.Second))
				defer limiter.SetLimit(rate.Inf)
				continue
			}
			return err
		}
	}
	return err
}

func RetryFetch(mangaID int) (*mangadex.MangaInfo, mangadex.ChapterList, []mangadex.PathInfo, error) {
	var base *mangadex.MangaInfo
	var list mangadex.ChapterList
	var covers []mangadex.PathInfo
	var inner error
	err := Retry(func() error {
		base, list, covers, inner = mangadex.Fetch(mangaID)
		if inner != nil {
			return inner
		}
		return nil
	})

	return base, list, covers, err
}
