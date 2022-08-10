package kindle

import (
	"fmt"
	"image/jpeg"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/leotaku/kojirou/cmd/formats"
	md "github.com/leotaku/kojirou/mangadex"
)

type NormalizedDirectory struct {
	target       string
	title        string
	kindleFolder bool
}

func NewNormalizedDirectory(target, title string, kindleFolder bool) NormalizedDirectory {
	return NormalizedDirectory{target, title, kindleFolder}
}

func (n *NormalizedDirectory) Has(identifier md.Identifier) bool {
	filename := identifier.StringFilled(4, 2, false) + ".azw3"
	if n.kindleFolder {
		return exists(path.Join(n.target, "documents", pathnameFromTitle(n.title), filename))
	} else {
		return exists(path.Join(n.target, pathnameFromTitle(n.title), filename))
	}
}

func (n *NormalizedDirectory) Write(part md.Manga, p formats.Progress) error {
	if part.Info.Title != n.title {
		return fmt.Errorf("unsupported configuration: title changed")
	}
	if len(part.Volumes) != 1 {
		return fmt.Errorf("unsupported configuration: multiple volumes")
	}
	volume := part.Sorted()[0]
	pathname := path.Join(pathnameFromTitle(n.title), volume.Info.Identifier.StringFilled(4, 2, false)+".azw3")
	if n.kindleFolder {
		pathname = path.Join("documents", pathname)
	}

	f, err := create(path.Join(n.target, pathname))
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	mobi := GenerateMOBI(part)
	if err := mobi.Realize().Write(p.NewProxyWriter(f)); err != nil {
		f.Close()
		return fmt.Errorf("write: %w", err)
	}
	f.Close()
	if n.kindleFolder && volume.Cover != nil {
		f, err := create(path.Join(n.target, "system", "thumbnails", mobi.GetThumbFilename()))
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		if err := jpeg.Encode(p.NewProxyWriter(f), volume.Cover, nil); err != nil {
			f.Close()
			return fmt.Errorf("write: %w", err)
		}
		f.Close()
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
	if os.IsNotExist(err) {
		return false
	} else if os.IsExist(err) {
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
