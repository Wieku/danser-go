# osu-vs-player #

osu-vs-player 用于同屏回放多个 osu replay，并实时显示replay信息，例如player名、按键、acc、评级、pp等。

灵感来源于 youtube 的那几个 50 Top plays on xxx 的视频。

本项目改造自 osu可视化渲染工具 [https://github.com/Wieku/danser](https://github.com/Wieku/danser) 。

<small>（顺便提一句，原项目似乎也在开发相关功能）</small>

## 历史版本同屏回放视频展示 ##

- 0.2.4 - [【50光标同屏】50 Top plays on Snow Drive(0123)](https://www.bilibili.com/video/av39908000)

- 0.3.0 - [【50光标同屏】50 Top plays on Road of Resistance [Crimson Rebellion]](https://www.bilibili.com/video/av40701715)

- 0.3.1 - [【50光标同屏】50 Top plays on Tengaku [Uncompressed Fury of a Raging Japanese God]](https://www.bilibili.com/video/av40891023)

- 0.3.2 - [【50光标同屏】50 Top plays on Halozy - 源流懐古 [Higan Torrent]](https://www.bilibili.com/video/av41365133)

## 如何运行（仅 Windows） ##

1. 从 [releases](https://github.com/wasupandceacar/osu-vs-player/releases) 下载压缩包。

2. 调整 ```settings.json``` 的参数。

3. 运行 ```main.exe``` 即可。

##	关于运行参数 ##

在 ```settings.json``` 中有大量的参数，其中大部分参数来自于原项目，请自行试验其效果。

本项目加入的参数全部在 ```VSplayer``` 之下，其中的参数含义如下：

### PlayerInfo ###

- ```Players``` - 同屏 replay 的个数
- ```SpecifiedPlayers``` - 是否指定播放特定 replay，指定时上一个参数失效，将重新计算 replay 个数
- ```SpecifiedLine```- 指定的 replay 索引，为数字，从 1 开始计数，以 ```,``` 隔开

### PlayerInfoUI ###

- ```BaseSize``` - 左侧信息的基准大小（相当于一个标准单位，各种基准具体效果请自己尝试）
- ```BaseX``` - 左侧信息的基准 X 坐标
- ```BaseY``` - 左侧信息的基准 Y 坐标
- ```ShowMouse1``` - 左侧是否显示第一个鼠标按键（M1）
- ```ShowMouse2``` - 左侧是否显示第二个鼠标按键（M2）

### RecordInfoUI ###

- ```Recorder``` - 右下角录制信息的录制人
- ```RecordTime``` - 右下角录制信息的录制时间
- ```RecordBaseX``` - 右下角录制信息的基准 X 坐标
- ```RecordBaseY``` - 右下角录制信息的基准 Y 坐标
- ```RecordBaseSize``` - 右下角录制信息的基准大小
- ```RecordAlpha``` - 右下角录制信息的透明度（0-1）

### PlayerFieldUI ###

- ```HitFadeTime``` - 每个 object 判定后显示的时间
- ```SpinnerMinusTime``` - 转盘提前显示时间
- ```SpinnerMult``` - 转盘大小参数
- ```ReverseFadeMult``` - 滑条总体时间与滑条折返开始显示时间之比
- ```CursorColorNum``` - object（note、滑条、转盘）的颜色索引 X，为第 X 个 replay 光标的颜色
- ```CursorColorSkipNum``` - 光标颜色跳选参数。设为0时，无跳选，光标会颜色依次渐变，可能会造成区分度过低。自行调整以获得最佳效果

### MapInfo ###

- ```Title``` - 地图名
- ```Difficulty``` - 难度名

### Mods ###

- ```EnableDT``` - 开启 ```DT``` 模式

### BreakandQuit ###

- ```EnableBreakandQuit``` - 是否开启大逃杀模式（即断连就淘汰，从画面中淡出）
- ```PlayerFadeTime``` - 断连后淡出的时间
- ```SameTimeOffset``` - 在断连位置显示 player 名时，如果两个人同时断连，第二个人的 player 名会在 Y 方向进行偏移（正为向下偏移）
- ```MissMult``` - 显示断连 player 名的基准大小

### ReplayandCache ###

- ```ReplayDir``` - replay 的目录
- ```CacheDir``` - replay 分析结果缓存的目录
- ```SaveResultCache``` - 是否缓存 replay 分析结果（acc、评级、pp等）。因为如果不修改判定逻辑，replay 每次分析结果应该一样，而分析大量 replay 需要一定时间（数分钟），所以可以事先保存 replay 的分析结果然后直接读取（几秒），从而更有效率地测试总体效果
- ```ReadResultCache``` - 是否读取 replay 分析结果缓存。这个开启时将不会重新缓存 replay 分析结果
- ```ReplayDebug``` - 开启只 debug replay 分析。在这种情况下，只会载入地图和 replay，并分析结果，不会绘制任何窗口。并且会强制分析 replay，且不会缓存结果

### ErrorFix ###

- ```EnableErrorFix``` - 由于 replay 分析的误差，当存在一些较大的误差时，暂时可以人工修正，来获得更准确的结果
- ```ErrorFixFile``` - 修正所读取的文件


## 关于自行编译 ##

由于修改了调用的用 ```Go``` 重写的oppai库（因为需要加入 pp nerf 和**实时**计算 pp），直接下载源码预计**大概可能肯定一定**无法编译成功。

当然如果你能够大概猜出我修改了什么（上面我已经提了），并且修改正确，应该就能编译得出来（不。