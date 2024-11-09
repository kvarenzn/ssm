# (WIP) ssm - Star Stone Miner
(正在开发中) *BanG Dream! 少女乐团派对！* 星石挖掘机

## 简介
- 适用于支持AoA v2协议的Android设备
- 代码及实现思路继承自我的另一个（已停止更新的）项目 [phisap](https://github.com/kvarenzn/phisap) ，使用go重写
- 解包部分参考 [Perfare/AssetStudio](https://github.com/Perfare/AssetStudio.git) 和 [nesrak1/AssetsTools.NET](https://github.com/nesrak1/AssetsTools.NET.git)，在此致谢
- Texture2D解码部分参考 [Perfare/AssetStudio](https://github.com/Perfare/AssetStudio.git) 和 [AssetRipper/AssetRipper](https://github.com/AssetRipper/AssetRipper.git) ，在此致谢
- **自用**

## 特点
- 内置资源解包模块，可以提取游戏中的图像资源和乐谱数据
- （一定程度上）规避检测
	- 非侵入式设计，不会干预游戏进程
	- 无需root权限
	- 无需启用USB调试
	- 触点可以通过“显示点按操作反馈”显示
- 采用图着色算法分配指针ID，使用尽可能少的触点完成演奏
	- 理论上95%以上的谱面可仅使用两根手指完成
- 使用golang编写
	- *“可以和我GO!!!!!一辈子吗”*

## FAQ
- Q: 如何直连BanG Dream(TM) Girls Band Party日服
	- A: 通过修改hosts或建立私人DNS服务器的方式将游戏设备的域名`api.garupa.jp`解析到`150.230.215.37`

## 开源协议
GPLv3
