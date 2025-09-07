# 使用方法
整体流程和 [phisap](https://github.com/kvarenzn/phisap) 类似。

## 准备工作
1. **准备一台电脑**  
   - 台式机、笔记本、树莓派都行，对性能无硬性要求  
   - 支持 Windows / Linux / macOS  
   - 自动构建版本支持以下架构（注意区分 **amd64** 与 **arm64**）：  
     - Linux amd64 (x86_64)  
     - Linux arm64 (aarch64)  
     - macOS x86_64  
     - macOS arm64  
     - Windows x86_64  
   - 其他系统/架构可尝试 [自行构建](./docs/BUILD.md)  

2. **安装依赖**  
   - FFmpeg 和 Libusb  
     - **Windows**：release 包已自带所需 dll，解压后可直接运行  
     - **macOS**：推荐使用 `brew install ffmpeg libusb`  
     - **Linux**：需 `libusb-1.0.so`、`libavcodec.so`、`libavformat.so`、`libavutil.so`，具体命令因发行版不同而异  
   - 若使用 `adb` 后端，请确保 `adb version` 可正常运行  

3. **测量屏幕参数**  
   - 约定：短边为宽，长边为高  
   - 获取方法：查看系统设置或截屏测量  

4. **测量判定线位置**（单位：像素，允许 ±20px 误差，整数即可）  
   - 判定线 → 顶端距离 = `Y`  
   - 判定线最左点 → 屏幕左边距 = `X1`  
   - 判定线最右点 → 屏幕左边距 = `X2`  
   ![测量数据示意图](./imgs/scales.jpg)

5. **导入游戏素材**  
   - 将游戏设备的 `/sdcard/Android/data/{游戏包名}/files/data/` 整个目录复制到电脑  
   - 可用 adb 命令：`adb pull /sdcard/Android/data/jp.co.craftegg.band/files/data/`  
   - 每次更新（新歌/新难度）后需重新导入  

6. **解包**  
   - 运行 `ssm -e {数据文件夹}`（Windows 为 `ssm.exe`）  
   - 解包完成后会在可执行文件同目录生成 `assets/`

## 开始打歌
1. 在 [bestdori](https://bestdori.com) 或类似网站查到歌曲 ID  
   - 例：《EXIST》= `325`  
2. 将设备连上电脑  
3. 启动游戏（横屏模式）  
4. 命令行运行：  
   ```bash
   ./ssm -d {难度} -n {歌曲ID} -r {旋转方向} -b {后端}
   ```
   - `{难度}`：`easy` / `normal` / `hard` / `expert` / `special`
   - `{旋转方向}`: `left` / `right`，默认为`left`，可省略
   - `{后端}`：`hid` / `adb`，默认为`hid`，可省略
   - 示例：`ssm -d expert -n 325`
   - 首次运行需输入准备阶段的测量数据
   - 控制台提示`ENTER/SPACE GO!!!!!`后，准备完成
5. 在游戏中进入曲目
6. 当第一个音符即将达到判定线时，在控制台中按 **ENTER** 或 **空格**
7. ssm将接管剩下的演奏
8. 若偏早/偏晚，可用方向键调整延迟
    - ← = -10ms / → = +10ms
    - Shift+方向键 = ±50ms，Ctrl+方向键 = ±100ms
9. 若要中断，控制台中输入 **Ctrl-C**
