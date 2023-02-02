package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/rivo/tview"
)

func main() {
	HOFApp()
}

func HOFApp() {
	getManga := func() *MangaInfo {
		manga, err := ScrapVyvy()
		if err != nil {
			panic("[Fatal error]: cannot retrieve yaoi infos")
		}
		return manga
	}

	app := tview.NewApplication()
	nextMangas := []*MangaInfo{}

	for {
		onNext := make(chan bool)

		var currManga *MangaInfo
		if len(nextMangas) == 0 {
			currManga = getManga()
		} else {
			currManga = nextMangas[0]
			nextMangas = append([]*MangaInfo{}, nextMangas[1:]...)
		}

		go displayManga(app, currManga, onNext)
		go func() {
			nextMangas = append(nextMangas, getManga())
		}()

		<-onNext
	}
}

func displayManga(app *tview.Application, manga *MangaInfo, onNext chan bool) {
	imgBuf, _, err := fetchImage(manga.coverUrl)
	if err != nil {
		log.Fatalln("[Fatal error]: cannot retrieve cover", err)
	}

	image := tview.NewImage()
	image.SetImage(imgBuf)

	image.SetColors(tview.TrueColor)
	image.SetDithering(tview.DitheringFloydSteinberg)

	list := tview.NewList().
		AddItem("Title", manga.title, '*', nil).
		AddItem("Views", fmt.Sprintf("%d view", manga.views), '*', nil).
		AddItem("Genres", strings.Join(manga.genres, ", "), '*', nil).
		AddItem("Rating", manga.rating, '*', nil).
		AddItem("Status", StatusToText(manga.completed), '*', nil).
		AddItem("Length", fmt.Sprintf("%d chapters", manga.length), '*', nil).
		AddItem("See", "", '👀', func() {
			exec.Command("xdg-open", manga.url).Run()
		}).
		AddItem("Next Manga", "", '⏭', func() {
			onNext <- true
		})

	grid := tview.NewGrid().
		SetBorders(true).
		AddItem(list, 0, 0, 1, 2, 0, 0, false).
		AddItem(image, 0, 2, 1, 3, 0, 0, false)

	if err := app.SetRoot(grid, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
