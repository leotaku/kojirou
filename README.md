<a href="https://mangadex.org/title/22631"><img src="./.github/header.jpg" alt="Header Image" width="75%"></a>

<h1>
  <span>Manki</span>
  <a href="https://goreportcard.com/report/github.com/leotaku/manki">
    <img src="https://goreportcard.com/badge/github.com/leotaku/manki" alt="Go Report Card">
  </a>
  <a href="https://github.com/leotaku/manki/actions">
    <img src="https://github.com/leotaku/manki/workflows/check/badge.svg" alt="Github Actions CI">
  </a>
</h1>

> Generate perfectly formatted Kindle EBooks from MangaDex manga

## Features

### Download manga and generate Kindle EBooks

Manki will automatically download the series for the specified ID and language while outputting a folder with all the downloaded volumes.

``` shell
manki 22631 -l en
```

### Generate Kindle folder structure for easy synchronization

Manki can also output a folder structure matching that of any modern Kindle devices to allow for easy synchronization using e.g. rsync.

``` shell
manki 22631 -l en --kindle-folder-mode
rsync kindle/ /run/media/user/Kindle/
```

### Customize ranking for better scantlations

Manki has the ability to use different [ranking algorithms](https://github.com/leotaku/manki/wiki/Ranking) in order to always dowload the highest-quality scantlations.
You can preview what would be downloaded by running in dry-run mode.

``` shell
manki 22631 -l en --rank views --dry-run
manki 22631 -l en --rank most
```

## License

[MIT](./LICENSE) © Leo Gaskin 2020-2021
