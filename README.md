<a href="https://mangadex.org/title/22631"><img src="./.github/header.jpg" alt="Header Image" width="80%"></a>

<h1>
  <span>Kojirou</span>
  <a href="https://goreportcard.com/report/github.com/leotaku/kojirou">
    <img src="https://goreportcard.com/badge/github.com/leotaku/kojirou?style=flat-square" alt="Go Report Card">
  </a>
  <a href="https://github.com/leotaku/kojirou/actions">
    <img src="https://img.shields.io/github/workflow/status/leotaku/kojirou/check?label=check&logo=github&logoColor=white&style=flat-square" alt="Github CI Status">
  </a>
  <a href="https://github.com/leotaku/kojirou/wiki/Home">
    <img src="https://img.shields.io/github/workflow/status/leotaku/kojirou/wiki?label=wiki&color=blue&logo=gitbook&logoColor=white&style=flat-square" alt="GitHub Wiki Status">
  <a/>
</h1>

> Generate perfectly formatted Kindle EBooks from MangaDex manga

## Features

### Download manga and generate Kindle EBooks

Kojirou will automatically download the series for the specified ID and language while outputting a folder with all the downloaded volumes.

``` shell
kojirou 22631 -l en
```

### Generate Kindle folder structure for easy synchronization

Kojirou can also output a folder structure matching that of any modern Kindle devices to allow for easy synchronization using e.g. rsync.

``` shell
kojirou 22631 -l en --kindle-folder-mode
rsync kindle/ /run/media/user/Kindle/
```

### Customize ranking for better scantlations

Kojirou has the ability to use different [ranking algorithms](https://github.com/leotaku/kojirou/wiki/Ranking) in order to always dowload the highest-quality scantlations.
You can preview what would be downloaded by running in dry-run mode.

``` shell
kojirou 22631 -l en --rank views --dry-run
kojirou 22631 -l en --rank most
```

## License

[MIT](./LICENSE) © Leo Gaskin 2020-2021
