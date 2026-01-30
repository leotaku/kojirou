package kindle

import (
	"errors"
	"fmt"
	"image/jpeg"
	"io/fs"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/leotaku/kojirou/cmd/formats"
	md "github.com/leotaku/kojirou/mangadex"
	"github.com/leotaku/mobi"
)

type NormalizedDirectory struct {
	bookDirectory      string
	thumbnailDirectory string
}

func NewNormalizedDirectory(target, title string, kindleFolder bool) NormalizedDirectory {
	switch {
	case kindleFolder && target == "":
		return NormalizedDirectory{
			bookDirectory:      path.Join("kindle", "documents", pathnameFromTitle(title)),
			thumbnailDirectory: path.Join("kindle", "system", "thumbnails"),
		}
	case kindleFolder:
		return NormalizedDirectory{
			bookDirectory:      path.Join(target, "documents", pathnameFromTitle(title)),
			thumbnailDirectory: path.Join(target, "system", "thumbnails"),
		}
	case target == "":
		return NormalizedDirectory{
			bookDirectory: pathnameFromTitle(title),
		}
	default:
		return NormalizedDirectory{
			bookDirectory: target,
		}
	}
}

func (n *NormalizedDirectory) Has(identifier md.Identifier) bool {
	filename := identifier.StringFilled(4, 2, false) + ".azw3"
	return exists(path.Join(n.bookDirectory, filename))
}

func (n *NormalizedDirectory) Write(identifier md.Identifier, mobi mobi.Book, p formats.Progress) error {
	if n.bookDirectory == "" {
		return fmt.Errorf("unsupported configuration: no book output")
	}
	filename := identifier.StringFilled(4, 2, false) + ".azw3"

	f, err := create(path.Join(n.bookDirectory, filename))
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	defer f.Close() //nolint:errcheck
	if err := mobi.Realize().Write(p.NewProxyWriter(f)); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	if n.thumbnailDirectory != "" && mobi.CoverImage != nil {
		f, err := create(path.Join(n.thumbnailDirectory, mobi.GetThumbFilename()))
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		defer f.Close() //nolint:errcheck
		if err := jpeg.Encode(p.NewProxyWriter(f), mobi.CoverImage, nil); err != nil {
			return fmt.Errorf("write: %w", err)
		}
	}

	return nil
}

func pathnameFromTitle(filename string) string {
	switch runtime.GOOS {
	case "windows":
		filename = strings.ReplaceAll(filename, "\"", "＂")
		filename = strings.ReplaceAll(filename, "\\", "＼")
		filename = strings.ReplaceAll(filename, "<", "＜")
		filename = strings.ReplaceAll(filename, ">", "＞")
		filename = strings.ReplaceAll(filename, ":", "：")
		filename = strings.ReplaceAll(filename, "|", "｜")
		filename = strings.ReplaceAll(filename, "?", "？")
		filename = strings.ReplaceAll(filename, "*", "＊")
		filename = strings.TrimRight(filename, ". ")
	case "darwin":
		filename = strings.ReplaceAll(filename, ":", "：")
	}

	return strings.ReplaceAll(filename, "/", "／")
}

func exists(pathname string) bool {
	_, err := os.Stat(pathname)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	} else if errors.Is(err, fs.ErrExist) {
		return true
	} else if err != nil {
		return false
	} else {
		return true
	}
}

func create(pathname string) (*os.File, error) {
	if err := os.MkdirAll(path.Dir(pathname), os.ModePerm); err != nil {
		return nil, fmt.Errorf("directory: %w", err)
	}
	if f, err := os.Create(pathname); err != nil {
		return nil, fmt.Errorf("file: %w", err)
	} else {
		return f, nil
	}
}
