# danser
danser is a visualiser for osu! maps written in Go.

Application is in dev phase so only few things work. But if you want to test it, you should follow steps at the end of this readme.

## Dance examples
* [Sex Whales & Fraxo - Dead To Me (feat. Lox Chatterbox) [extrad1881 (ar 10)] Mirror Collage](https://youtu.be/KCHqrVGdXrk)
* [Halozy - Genryuu Kaiko [Higan Torrent] Mirror Collage](https://youtu.be/HCVIBQh4ljI)
* [Nightcore - Flower Dance [Amachoco ARX.7] Mandala Mirror Collage](https://youtu.be/HBC89S-UwFc)

## How to run it
You should have 64-bit [go](https://golang.org/dl/) and `gcc` installed. Instead of cloning this repository, in your terminal type:
```bash
go get -u github.com/wieku/danser
```

To build the project (you have to be in project directory), run:
```bash
go build
```
then:
```bash
go run main.go
```

#### Run flags
* `-artist="NOMA"`
* `-title="Brain Power"`
* `-difficulty="Overdrive"`
* `-creator="Skystar"`
* `-cursors=2` - number of cursors used in mirror collage
* `-tag=2` - number of TAG cursors
* `-speed=1.5` - music speed. Value of 1.5 equals to osu!'s DoubleTime
* `-mover=flower` - cursor mover. Movers available now: linear, bezier, flower (default), circular.
* `-settings=1` - if number given is bigger than 0 (e.g. 2) then app will try to load `settings-2.json` instead of `settings.json`

Example:
```bash
go run main.go -title="Brain Power" -difficulty="Overdrive" -tag=2
```


## Credits

Original game was made by Dean Herbert ([@ppy](https://github.com/ppy)) and [osu!](https://osu.ppy.sh/) community.

Map assets were made by [Haskorion](https://osu.ppy.sh/users/3252321): [Redd Glass HD](https://osu.ppy.sh/community/forums/topics/211396)
