<p align="center">
  <img width="500px" src="assets/textures/coinbig.png"/>
</p>

# danser-go

[![GitHub release](https://img.shields.io/github/release/wieku/danser-go.svg)](https://github.com/Wieku/danser-go/releases/latest)
[![CodeFactor](https://www.codefactor.io/repository/github/wieku/danser-go/badge)](https://www.codefactor.io/repository/github/wieku/danser-go)
[![Discord server](https://img.shields.io/discord/713705871758065685.svg?label=&logo=discord&logoColor=ffffff&color=7389D8&labelColor=6A7EC2)](https://discord.gg/UTPvbe8)

danser-go is a visualiser for osu! maps written in Go.

Application is in dev phase so only few things work. But if you want to test it, you should follow steps at the end of this readme.

## Dance examples
* [REOL - YoiYoi Kokon [Yoi] - Dance comparison](https://youtu.be/QZ6-MaWWyA8)
* [Omoi - Chiisana Koi no Uta (Synth Rock Cover) [Kroytz's EX EX] - TAG2 Mirror Collage](https://youtu.be/Vo0Pbpu113Y)
* [Sex Whales & Fraxo - Dead To Me (feat. Lox Chatterbox) [extrad1881 (ar 10)] Mirror Collage](https://youtu.be/KCHqrVGdXrk)
* [Halozy - Genryuu Kaiko [Higan Torrent] Mirror Collage](https://youtu.be/HCVIBQh4ljI)
* [Nightcore - Flower Dance [Amachoco ARX.7] Mandala Mirror Collage](https://youtu.be/HBC89S-UwFc)
* [MDK - Press Start [bhop_start_collab] - Co-Op Mirror Collage](https://youtu.be/P5mYXvH48Uk)

## How to download it

### Windows
You can download windows binaries from [releases](https://github.com/Wieku/danser-go/releases).

### Linux/Unix/Windows

If you want to build it yourself, you need 64bit [go](https://golang.org/dl/), `gcc/mingw64` and libgl (on Linux/Unix).

In your terminal type:
```bash
go get -u github.com/wieku/danser-go
```
This will automatically download and compile needed libraries.

## How to run it

### Windows executable
```bash
danser***.exe <arguments>
```

### Project
You have to be in the project directory.

If you're running it for the first time or when you made some changes type:
```bash
go build
```

Then type:
```bash
danser-go <arguments>
```

#### Arguments
* `-artist="NOMA"` or `-a="NOMA"`
* `-title="Brain Power"` or `-t="Brain Power"`
* `-difficulty="Overdrive"` or `-d="Overdrive"`
* `-creator="Skystar"` or `-c="Skystar"`
* `-cursors=2` - number of cursors used in mirror collage
* `-tag=2` - number of TAG cursors
* `-speed=1.5` - music speed. Value of 1.5 equals to osu!'s DoubleTime
* `-pitch=1.5` - music pitch. Value of 1.5 equals to osu!'s Nightcore pitch. To recreate osu!'s Nightcore mod, use with 1.5 speed
* `-mover=flower` - cursor mover. Movers available now: linear, bezier, flower (default), circular, aggressive.
* `-settings=1` - if number given is bigger than 0 (e.g. 1) then app will try to load `settings-1.json` instead of `settings.json`
* `-fps` - shows fps in the lower-left corner 
* `-debug` - shows more info during the map

Example:
```bash
<executable> -title="Brain Power" -difficulty="Overdrive" -tag=2
```

## Credits

Original game was made by Dean Herbert ([@ppy](https://github.com/ppy)) and [osu!](https://osu.ppy.sh/) community.

Map assets were made by [Haskorion](https://osu.ppy.sh/users/3252321): [Redd Glass HD](https://osu.ppy.sh/community/forums/topics/211396)
