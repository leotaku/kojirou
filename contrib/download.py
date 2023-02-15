#!/usr/bin/env python3
import asyncio
import re
import os.path
import lxml.html

from typing import AsyncIterable, Awaitable, Iterable, Optional, TypeVar
from httpx import AsyncClient, AsyncHTTPTransport
from fs import open_fs
from pydantic import BaseModel


async def main():
    url = "https://readclaymore.com/"
    title = "Claymore"
    directory = "Claymore"
    authors = artists = ["Norihiro Yagi"]
    await download(url, title, directory)

    # url = "https://tsurezurechildren.com/"
    # title = "Tsurezure Children"
    # directory = "Tsurezure Children"
    # authors = artists = ["Wakabayashi Toshiya"]
    # await download(url, title, directory)

    # url = "https://readgokushufudou.com/"
    # title = "Gokushufudou: The Way of the House Husband"
    # directory = "Gokushufudou: The Way of the House Husband"
    # authors = artists = ["Kousuke Oono"]
    # await download(url, title, directory)

    # url = "https://komisanmanga.com/"
    # title = "Komi-san wa Komyushou Desu"
    # directory = "Komi-san wa Komyushou Desu."
    # authors = artists = ["Tomohito Oda"]
    # await download(url, title, directory)


async def download(url, title, directory):
    async with AsyncClient(
        timeout=100, transport=AsyncHTTPTransport(retries=10)
    ) as client:
        root = await client.get(url)

        with open_fs(directory, create=True) as root_fs:
            async for (html, chapter) in run_with_semaphore(
                asyncio.Semaphore(100),
                (
                    (client.get(url), chapter)
                    for (url, chapter) in get_chapter_urls(root.text, title)
                ),
            ):
                print(chapter)
                directory = chapter.volume_and_chapter_directory()
                if root_fs.exists(directory):
                    print("skipped")
                    continue
                with root_fs.makedir(directory, recreate=True) as chapter_fs:
                    async for (html, number) in run_with_semaphore(
                        asyncio.Semaphore(20),
                        (
                            (client.get(url), number)
                            for (number, url) in enumerate(get_page_urls(html.text))
                        ),
                    ):
                        (_, ext) = os.path.splitext(str(html.url))
                        filename = str(number).zfill(4) + ext
                        chapter_fs.writebytes(filename, html.content)
                        print(filename)


def get_chapter_urls(text: str, title: str) -> Iterable[tuple[str, "Chapter"]]:
    for link in reversed(lxml.html.fromstring(text).xpath("//main//a")):
        chapter = Chapter.from_title(link.text)
        if (chapter.title or "").lower() == title.lower():
            yield link.attrib["href"], chapter


def get_page_urls(text: str):
    for image in lxml.html.fromstring(text).xpath("//img"):
        source = image.attrib["src"]
        if source.startswith("http"):
            yield source


class Chapter(BaseModel):
    title: Optional[str]
    volume_id: Optional[str]
    chapter_id: Optional[str]
    chapter_title: Optional[str]

    @staticmethod
    def from_title(string: str) -> "Chapter":
        match = re.match("(.*), Vol.(.*) Chapter (.*): (.*)", string)
        if match != None:
            return Chapter(
                title=match.group(1),
                volume_id=match.group(2),
                chapter_id=match.group(3),
                chapter_title=match.group(4),
            )

        match = re.match("(.*), Vol.(.*) Chapter (.*)", string)
        if match != None:
            return Chapter(
                title=match.group(1),
                volume_id=match.group(2),
                chapter_id=match.group(3),
                chapter_title=None,
            )

        match = re.match("(.*), Chapter (.*): (.*)", string)
        if match != None:
            return Chapter(
                title=match.group(1),
                volume_id=None,
                chapter_id=match.group(2),
                chapter_title=match.group(3),
            )

        match = re.match("(.*), Chapter (.*)", string)
        if match != None:
            return Chapter(
                title=match.group(1),
                volume_id=None,
                chapter_id=match.group(2),
                chapter_title=None,
            )

        return Chapter(
            title=None,
            volume_id=None,
            chapter_id=None,
            chapter_title=None,
        )

    def volume_and_chapter_directory(self) -> str:
        filename = fill_id(self.volume_id) + "/" + fill_id(self.chapter_id)
        if self.chapter_title:
            filename += ": " + self.chapter_title

        return filename


def fill_id(num: Optional[str]) -> str:
    if num == None:
        return "Unknown"

    split = num.split(".")
    filename = split[0].zfill(4)
    if len(split) == 2:
        filename += "." + split[1].zfill(2)

    return filename


A = TypeVar("A")
T = TypeVar("T")


async def run_with_semaphore(
    semaphore: asyncio.Semaphore,
    generator: Iterable[tuple[Awaitable[A], T]],
) -> AsyncIterable[tuple[A, T]]:
    async def await_with_semaphore(semaphore, coroutine, context) -> tuple[A, T]:
        async with semaphore:
            return (await coroutine, context)

    for result in asyncio.as_completed(
        (
            asyncio.create_task(await_with_semaphore(semaphore, coroutine, context))
            for (coroutine, context) in generator
        )
    ):
        yield await result


if __name__ == "__main__":
    asyncio.run(main())
