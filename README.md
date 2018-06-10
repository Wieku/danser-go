# danser
danser is a visualiser for osu! maps written in Go.

Application is in dev phase so only few things work. But if you want to test it, you should follow steps at the end of this readme.

## Dance examples
* [Halozy - Genryuu Kaiko [Higan Torrent] Mirror Collage](https://youtu.be/HCVIBQh4ljI)
* [Nightcore - Flower Dance [Amachoco ARX.7] Mandala Mirror Collage](https://youtu.be/HBC89S-UwFc)

## How to run it
You should have go sdk installed. Instead of cloning this repository, in your terminal type:
```$xslt
go get -u github.com/wieku/danser
```

To build the project (you have to be in project directory), run:
```$xslt
go build
```
then:
```$xslt
go run main.go
```

#### Run flags
* `-artist="NOMA"`
* `-title="Brain Power"`
* `-difficulty="Overdrive"`
* `-maker="Skystar"`
* `-cursors=2` - number of cursors used in mirror collage
* `-settings=1` - if number given is bigger than 0 (e.g. 2) then app will try to load `settings-2.json` instead of `settings.json`

Example:
```$xslt
go run main.go -title="Brain Power" -difficulty="Overdrive"
```


## Credits

Original game was made by Dean Herbert ([@ppy](https://github.com/ppy)) and osu! community.

Map assets were made by [Haskorion](https://osu.ppy.sh/users/3252321): [Redd Glass HD](https://osu.ppy.sh/community/forums/topics/211396)
