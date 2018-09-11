# danser
danser is a visualiser for osu! maps written in Go.

Application is in dev phase so only few things work. But if you want to test it, you should follow steps at the end of this readme.

## Dance examples
* [Omoi - Chiisana Koi no Uta (Synth Rock Cover) [Kroytz's EX EX] - TAG2 Mirror Collage](https://youtu.be/Vo0Pbpu113Y)
* [Sex Whales & Fraxo - Dead To Me (feat. Lox Chatterbox) [extrad1881 (ar 10)] Mirror Collage](https://youtu.be/KCHqrVGdXrk)
* [Halozy - Genryuu Kaiko [Higan Torrent] Mirror Collage](https://youtu.be/HCVIBQh4ljI)
* [Nightcore - Flower Dance [Amachoco ARX.7] Mandala Mirror Collage](https://youtu.be/HBC89S-UwFc)

## How to download it

### Windows
You can download windows binaries from [releases](https://github.com/Wieku/danser/releases).

### Linux/Unix/Windows

If you want to build it yourself, you would need 64bit [go](https://golang.org/dl/), `gcc/mingw` and libgl (on Linux/Unix).

In your terminal type:
```bash
go get -u github.com/wieku/danser
```
This will automatically download and compile needed libraries.

## How to run it

### Windows executable
```bash
danser***.exe <arguments>
```

### Project
You have to be in the project directory.

If you're running it for the first time type:
```bash
go build
```

Then type:
```bash
danser <arguments>
```

#### Arguments
* `-artist="NOMA"`
* `-title="Brain Power"`
* `-difficulty="Overdrive"`
* `-creator="Skystar"`
* `-cursors=2` - number of cursors used in mirror collage
* `-tag=2` - number of TAG cursors
* `-speed=1.5` - music speed. Value of 1.5 equals to osu!'s DoubleTime
* `-mover=flower` - cursor mover. Movers available now: linear, bezier, flower (default), circular.
* `-settings=1` - if number given is bigger than 0 (e.g. 1) then app will try to load `settings-1.json` instead of `settings.json`

Example:
```bash
<executable> -title="Brain Power" -difficulty="Overdrive" -tag=2
```


## Credits

Original game was made by Dean Herbert ([@ppy](https://github.com/ppy)) and [osu!](https://osu.ppy.sh/) community.

Map assets were made by [Haskorion](https://osu.ppy.sh/users/3252321): [Redd Glass HD](https://osu.ppy.sh/community/forums/topics/211396)
