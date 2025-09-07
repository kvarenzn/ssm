<p align="center">
    <img src="imgs/icon.svg" width="128" height="128" alt="ssm"/>
</p>

# (WIP) ssm - Star Stone Miner
(开发中) *BanG Dream! 少女乐团派对！* 星石挖掘机

## 免责声明
> [!IMPORTANT]
> 本项目为个人学习与研究用途开发，不保证功能的稳定性与适用性。
>
> 本项目与 Craft Egg、Bushiroad、BanG Dream! 少女乐团派对！ 及其他相关公司、组织无任何从属或合作关系。
>
> 使用本项目可能会违反游戏的服务条款，甚至导致账号封禁或数据损坏。
>
> 作者不对因使用本项目所造成的任何后果承担责任，请自行评估风险并谨慎使用。

## 简介
- 仅适用于支持 **Android开放配件(AOA) 2.0** 协议的Android设备
	 - 2011年以后出厂的Android设备基本都支持AOA 2.0

### 优势
- 内置解包功能，可提取游戏中的图像和乐谱数据
- 尽量减少对游戏环境的干扰
  - 非侵入式设计，不干预游戏进程
  - 无需 root 权限
  - 不使用 `adb` 后端时，无需开启 USB 调试
  - 不修改游戏安装包或数据
  - 触点可通过「显示点按操作反馈」看到
- 智能指针分配：采用图着色算法，用尽可能少的触点完成演奏
  - 理论上 99% 的谱面可仅用两个触点完成
- 使用 Go 编写
  - *“可以和我 GO!!!!! 一辈子吗”*

### 缺陷
- 目前仅有 **命令行界面**
- 必须使用 USB 数据线连接设备
  - 使用 `adb` 后端时可尝试无线调试，但对局域网质量要求较高，延迟可能影响打歌准确度
- 必须手动触发开局

## 用法
```
Usage of ./ssm:
  -b hid
    	Specify ssm backend, possible values: hid, `adb` (default "hid")
  -d string
    	Difficulty of song
  -e string
    	Extract assets from assets folder <path>
  -g	Display useful information for debugging
  -n int
    	Song ID (default -1)
  -p string
    	Custom chart path (if this is provided, song ID and difficulty will be ignored)
  -r left
    	Device orientation, options: left (↺, counter-clockwise), `right` (↻, clockwise). Note: ignored when using `adb` backend (default "left")
  -s string
    	Specify the device serial (if not provided, ssm will use the first device serial)
  -v	Show ssm's version number and exit
```
更详细的安装步骤与使用说明，请参见[USAGE.md](./docs/USAGE.md)

## 常见问题
详见[FAQ.md](./docs/FAQ.md)

## TODO
- [ ] 图形化控制界面
- [X] 移植`scrcpy-server`控制功能
	- [ ] 读取游戏设备屏幕内容
        - [ ] 识别选中歌曲及难度
        - [ ] 自动开始
        - [ ] 自动重复

## 参考及引用
感谢以下项目的作者及维护者：
- 解包： [Perfare/AssetStudio](https://github.com/Perfare/AssetStudio.git) 、 [nesrak1/AssetsTools.NET](https://github.com/nesrak1/AssetsTools.NET.git)
- Texture2D解码： [Perfare/AssetStudio](https://github.com/Perfare/AssetStudio.git) 、 [AssetRipper/AssetRipper](https://github.com/AssetRipper/AssetRipper.git)
- `adb`后端依赖 [Genymobile/scrcpy](https://github.com/Genymobile/scrcpy) 的`scrcpy-server`
- 歌曲/乐队信息来自 [bestdori](https://bestdori.com) ：
    - 歌曲 ID ↔ 歌曲名
    - 歌曲 ID ↔ 乐队 ID
    - 歌曲 ID ↔ 封面 ID 列表
    - 乐队 ID ↔ 乐队名

## 开源协议
GPLv3
