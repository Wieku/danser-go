<p align="center">
  <img width="500px" src="assets/textures/coinbig.png"/>
</p>

# danser-go

[![GitHub release](https://img.shields.io/github/release/wieku/danser-go.svg)](https://github.com/Wieku/danser-go/releases/latest)
[![CodeFactor](https://www.codefactor.io/repository/github/wieku/danser-go/badge)](https://www.codefactor.io/repository/github/wieku/danser-go)
[![Discord server](https://img.shields.io/discord/713705871758065685.svg?label=&logo=discord&logoColor=ffffff&color=7389D8&labelColor=6A7EC2)](https://discord.gg/UTPvbe8)

danser-go is a CLI visualisation tool for osu!standard maps.

As danser is in development phase, some things may break. If that happens please fill an issue with as much detail as possible.

**WARNING**: Because of MacOS' poor OpenGL support, danser-go won't run on that platform. Please use dual-booted Windows/Linux instead.

## Examples
* [Omoi - Chiisana Koi no Uta (Synth Rock Cover) [Kroytz's EX EX] - TAG2 Mirror Collage](https://youtu.be/Vo0Pbpu113Y)
* [Sex Whales & Fraxo - Dead To Me (feat. Lox Chatterbox) [extrad1881 (ar 10)] Mirror Collage](https://youtu.be/KCHqrVGdXrk)
* [Nightcore - Flower Dance [Amachoco ARX.7] Mandala Mirror Collage](https://youtu.be/HBC89S-UwFc)
* [Flower Dance (osu! cursordance)](https://youtu.be/lcnnz3fN3bs)
* [osu! top 50 replays knockout | xi - FREEDOM DiVE [ENDLESS DiMENSiONS]](https://www.youtu.be/kzr_Sr0Shuc)
* [osu! top 50 knockout | YURRY CANNON - Suicide Parade [Sakase]](https://youtu.be/GS_yoq5MJMU)
* [osu! top 50 replays knockout | Kobaryo - Bookmaker [Corrupt The World]](https://youtu.be/SJqkP1IDUq0)

## Running Danser

You can download the newest Windows/Linux 64-bit binaries from [releases](https://github.com/Wieku/danser-go/releases).

After unpacking it to your desired directory, you need to run it using a command-line application/terminal:

##### Windows cmd:
```bash
danser*** <arguments>
```

##### Linux / Unix / git bash / Powershell:
```bash
./danser*** <arguments>
```

If you try to run Danser without any arguments there's a surprise waiting for you ;)

## Run arguments
* `-artist="NOMA"` or `-a="NOMA"`
* `-title="Brain Power"` or `-t="Brain Power"`
* `-difficulty="Overdrive"` or `-d="Overdrive"`
* `-creator="Skystar"` or `-c="Skystar"`
* `-md5=hash` - overrides all map selection arguments and attempts to find `.osu` file matching the specified MD5 hash
* `-cursors=2` - number of cursors used in mirror collage
* `-tag=2` - number of cursors in TAG mode
* `-speed=1.5` - music speed. Value of 1.5 is equal to osu!'s DoubleTime mod
* `-pitch=1.5` - music pitch. Value of 1.5 is equal to osu!'s Nightcore pitch. To recreate osu!'s Nightcore mod, use with speed 1.5
* `-settings=name` - settings filename - for example `settings-name.json` instead of `settings.json`
* `-debug` - shows additional info when running Danser, overrides `Graphics.DrawFPS` setting
* `-play` - play through the map in osu!standard mode
* `-skip` - skips map's intro like in osu!
* `-start=20.5` - start the map at a given time (in seconds)
* `-end=30.5` - end the map at the given time (in seconds)
* `-knockout` - knockout mode
* `-record` - Records danser's output to a video file. Needs a globally accessible [ffmpeg](https://ffmpeg.org/download.html) installation.
* `-out=abcd` - overrides `-record` flag, records to a given filename instead of auto-generating it. Extension of the file is set in settings.
* `-replay="path_to_replay.osr"` or `-r="path_to_replay.osr"` - plays a given replay file. Be sure to replace `\` with `\\` or `/`. Overrides all map selection arguments
* `-mods=HDHR` - displays the map with given mods. Overrides `-speed` and `-pitch` arguments if DT/NC/HT/DC mods are given
* `-skin` - overrides `Skin.CurrentSkin` in settings

Since danser 0.4.0b artist, creator, difficulty names and titles don't have to exactly match the `.osu` file. 

Examples which should give the same result:

```bash
<executable> -d="Overdrive" -tag=2 //Assuming that there is only ONE map with "Overdrive" as its difficulty name

<executable> -t="Brain Power" -d="Overdrive" -tag=2

<executable> -t "Brain Power" -d Overdrive -tag 2

<executable> -t="ain pow" -difficulty="rdrive" -tag=2

<executable> -md5=59f3708114c73b2334ad18f31ef49046 -tag=2
```

Settings and knockout usage are detailed in the [wiki](https://github.com/Wieku/danser-go/wiki).

## Building the project
You need to clone it or download as a .zip (and unpack it to desired directory)

#### Prerequisites

* [64-bit go (1.16 at least)](https://golang.org/dl/)
* gcc (Linux/Unix), [mingw-w64](http://mingw-w64.org/) or [WinLibs](http://winlibs.com/) (Windows, TDM-GCC won't work)
* OpenGL library (shipped with drivers)
* xorg-dev (Linux)

#### Building and running the project

First, enter the cloned/downloaded repository.

When you're running it for the first time or if you made any changes type:
```bash
go build
```

This will automatically download and build needed dependencies.

Afterwards type:
```bash
./danser-go <arguments>
```

## Credits

[osu!](https://osu.ppy.sh/) was created by osu! team ([@ppy](https://github.com/ppy)) and osu! community.

Default skin was created by [Haskorion](https://osu.ppy.sh/users/3252321): [Redd Glass HD](https://osu.ppy.sh/community/forums/topics/211396)

Uses [Exo2](https://fonts.google.com/specimen/Exo+2) font under [SIL Open Font License](http://scripts.sil.org/cms/scripts/page.php?site_id=nrsi&id=OFL_web)

Uses [Ubuntu](https://fonts.google.com/specimen/Ubuntu) font under [Ubuntu Font License](https://ubuntu.com/legal/font-licence)
