<a href="https://mangadex.org/title/22631"><img src="./.github/header.jpg" alt="Header Image" width="80%"></a>

<h1>
  <span>Kojirou</span>
  <a href="https://goreportcard.com/report/github.com/leotaku/kojirou">
    <img src="https://goreportcard.com/badge/github.com/leotaku/kojirou?style=flat-square" alt="Go Report Card">
  </a>
  <a href="https://github.com/leotaku/kojirou/actions">
    <img src="https://img.shields.io/github/actions/workflow/status/leotaku/kojirou/check.yml?branch=master&label=check&logo=github&logoColor=white&style=flat-square" alt="Github CI Status">
  </a>
  <a href="https://github.com/leotaku/kojirou/wiki/Home">
    <img src="https://img.shields.io/github/actions/workflow/status/leotaku/kojirou/wiki.yml?branch=master&label=wiki&color=blue&logo=gitbook&logoColor=white&style=flat-square" alt="GitHub Wiki Status">
  <a/>
</h1>

> Generate perfectly formatted Kindle EBooks from MangaDex manga

## Features

### Download manga and generate Kindle EBooks

Kojirou will automatically download the series for the specified ID and language while outputting a folder with all the downloaded volumes.

``` shell
kojirou d86cf65b-5f6c-437d-a0af-19a31f94ec55 -l en
```

### Generate Kindle folder structure for easy synchronization

Kojirou can also output a folder structure matching that of any modern Kindle device to allow for easy synchronization using e.g. rsync.

``` shell
kojirou d86cf65b-5f6c-437d-a0af-19a31f94ec55 -l en --kindle-folder-mode
udisksctl mount -b /dev/sdb
rsync kindle/ /run/media/user/Kindle/
```

### Customize ranking for better scantlations

Kojirou has the ability to use different [ranking algorithms](https://github.com/leotaku/kojirou/wiki/Ranking) in order to always dowload the highest-quality scantlations.
You can preview what would be downloaded by running in dry-run mode.

**Note:** Currently, the views and views-total ranking algorithms are broken because MangaDex no longer provides the required viewcount information.

``` shell
kojirou d86cf65b-5f6c-437d-a0af-19a31f94ec55 -l en --rank newest --dry-run
kojirou d86cf65b-5f6c-437d-a0af-19a31f94ec55 -l en --rank most
```

### Crop whitespace from pages automatically

Kojirou has the ability to crop whitespace from the borders of manga pages.
This may be useful if your e-reader has a small screen.

``` shell
kojirou d86cf65b-5f6c-437d-a0af-19a31f94ec55 -l en --autocrop
```

### Change reading direction

Kojirou, by default, generates e-books with right-to-left reading direction, as this is the default convention for most manga.
Also note that right-to-left reading does not seem to be supported on all Kindle devices.

``` shell
kojirou d86cf65b-5f6c-437d-a0af-19a31f94ec55 -l en --left-to-right
```

## Prebuilt binaries

Prebuilt binaries for Linux, Windows and MacOS on x86 and ARM processors are provided.
Visit the [release tab](https://github.com/leotaku/kojirou/releases) to download the archive for your respective setup.

On Linux and MacOS you will have to make the provided binary executable after extracting it from the archive.

``` shell
chmod u+x ./kojirou.exe
```

Afterwards, verify your installation succeeded by executing the application on the command line.

``` shell
./kojirou.exe --version
```

## Install from source

Kojirou can be installed easily if you already have Go installed, using the following command.
Otherwise, follow the [Go installation instructions](https://go.dev/doc/install) for your operating system and then run the command.

``` shell
go install github.com/leotaku/kojirou@latest
```

Afterwards, verify your installation succeeded by executing the application on the command line.

``` shell
kojirou --version
```

On many systems, the Go binary directory is not added to the list of directories searched for executables by default.
If you get a "command not found" or similar error after the previous command, run the following command and try again.
If you are using Windows, please find out how to add directories to the lookup path yourself, as there does not seem to be any quality documentation that I could link here.

``` shell
export PATH=$PATH:$(go env GOPATH)/bin
```

## License

[MIT](./LICENSE) © Leo Gaskin 2020-2023
