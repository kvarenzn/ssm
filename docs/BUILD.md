# 构建
1. 安装Go 1.24+
2. 安装gcc（推荐9.x+，Windows 用 MSYS2 + MinGW）
3. 安装pkgconf
4. 安装依赖库：`libusb-1.0`、`ffmpeg`、`libavcodec`、`libavformat`、`libavutil`
    - 部分系统可能打包在一起，注意安装`-dev`/`-devel`包
5. `git clone https://github.com/kvarenzn/ssm.git`
6. `cd ssm`
7. 构建
    ```bash
    go build -ldflags "-X main.SSM_VERSION=$VERSION"
    ```
    - 未指定版本号时，显示为`(unknown)`，不影响使用

具体请参考工作流程清单`.github/workflows/release.yml`
