// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package locale

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func regSimplifiedChinese() {
	message.SetString(language.SimplifiedChinese, "copyright info", `
Copyright (C) 2024, 2025 kvarenzn
授权协议 GPLv3+：GNU 通用公共许可证第 3 版或更新版本 <https://gnu.org/licenses/gpl.html>。
这是自由软件：您可以自由修改和重新发布它。
在法律允许的范围内，没有任何担保。`)
	message.SetString(language.SimplifiedChinese, "Usage of %s:", "ssm 的用法：")
	message.SetString(language.SimplifiedChinese, "usage.b", "指定 ssm 后端，可选值：`hid`，`adb`")
	message.SetString(language.SimplifiedChinese, "usage.n", "歌曲 ID")
	message.SetString(language.SimplifiedChinese, "usage.d", "歌曲难度")
	message.SetString(language.SimplifiedChinese, "usage.e", "从资源路径中解包资源")
	message.SetString(language.SimplifiedChinese, "usage.r", "设备方向，可选值：`left` （↺, 逆时针），`right`（↻, 顺时针）。注：当使用`adb`后端时本选项被忽略")
	message.SetString(language.SimplifiedChinese, "usage.p", "指定谱面路径（如果本选项被提供，歌曲 ID 和歌曲难度都会被忽略）")
	message.SetString(language.SimplifiedChinese, "usage.s", "指定设备序列号（如果未提供，ssm 会使用第一个检索到的设备序列号）")
	message.SetString(language.SimplifiedChinese, "usage.k", "切换到PJSK模式")
	message.SetString(language.SimplifiedChinese, "usage.g", "显示调试信息")
	message.SetString(language.SimplifiedChinese, "usage.v", "显示 ssm 的版本信息并退出")
	message.SetString(language.SimplifiedChinese, "ssm version: %s", "ssm 版本：%s")
	message.SetString(language.SimplifiedChinese, "(unknown)", "(未指定)")
	message.SetString(language.SimplifiedChinese, "To use adb as the backend, the third-party component `scrcpy-server` (version %s) is required.", "要使用adb作为后端，需要第三方组件`scrcpy-server` (%s 版本)。")
	message.SetString(language.SimplifiedChinese, "This component is developed by Genymobile and licensed under Apache License 2.0.", "这个组件是由Genymobile开发的，并以Apache License 2.0协议开源。")
	message.SetString(language.SimplifiedChinese, "Please download it from the official release page and place it in the same directory as `ssm.exe`.", "请从官方的release页面下载，然后将它放到'ssm'（或者'ssm.exe'）所在的文件夹。")
	message.SetString(language.SimplifiedChinese, "Download link:", "下载链接：")
	message.SetString(language.SimplifiedChinese, "Alternatively, ssm can automatically handle this process for you.", "或者，ssm也可以帮你自动完成这些。")
	message.SetString(language.SimplifiedChinese, "Proceed with automatic download? [Y/n]: ", "需要自动下载吗？ [Y/n]: ")
	message.SetString(language.SimplifiedChinese, "Failed to get input from user:", "尝试获取输入失败：")
	message.SetString(language.SimplifiedChinese, "`scrcpy-server` is required. To use `adb` as the backend, you should download it manually.", "需要`scrcpy-server`。如果希望选择`adb`作为后端，你需要手动下载它。")
	message.SetString(language.SimplifiedChinese, "Downloading... Please wait.", "正在下载...请稍等。")
	message.SetString(language.SimplifiedChinese, "Failed to download `scrcpy-server`.", "`scrcpy-server`下载失败。")
	message.SetString(language.SimplifiedChinese, "You may try again later, download it manually, or use `hid` backend instead.", "你可以稍后重试、手动下载，或使用`hid`作为后端。")
	message.SetString(language.SimplifiedChinese, "Failed to calculate sha256 of `scrcpy-server`:", "计算`scrcpy-server`的sha256失败：")
	message.SetString(language.SimplifiedChinese, "Checksum mismatch. Please try again later.", "校验和不匹配。请稍后重试。")
	message.SetString(language.SimplifiedChinese, "Failed to save `scrcpy-server` to disk:", "将`scrcpy-server`保存到硬盘失败：")
	message.SetString(language.SimplifiedChinese, "Failed to locate server file:", "查找`scrcpy-server`文件失败：")
	message.SetString(language.SimplifiedChinese, "Failed to read the content of `scrcpy-server`:", "读取`scrcpy-server`内容失败：")
	message.SetString(language.SimplifiedChinese, "Failed to calculate sha256 of `scrcpy-server`:", "计算`scrcpy-server`的sha256失败：")
	message.SetString(language.SimplifiedChinese, "Checksum mismatch. File may be corrupted.", "校验和不匹配。文件可能损坏。")
	message.SetString(language.SimplifiedChinese, "Failed to load music jacket: %s", "加载歌曲封面失败：%s")
	message.SetString(language.SimplifiedChinese, "ui line 0", "\x1b[7m\x1b[1m 回车/空格 \x1b[0m GO!!!!!")
	message.SetString(language.SimplifiedChinese, "Offset: %d ms", "偏移：%d 毫秒")
	message.SetString(language.SimplifiedChinese, "Failed to get key from stdin: %s", "从标准输入读取按键失败：%s")
	message.SetString(language.SimplifiedChinese, "ui line 1", "\x1b[7m\x1b[1m ← \x1b[0m -10ms   \x1b[7m\x1b[1m Shift-← \x1b[0m -50ms   \x1b[7m\x1b[1m Ctrl-← \x1b[0m -100ms   \x1b[7m\x1b[1m Ctrl-C \x1b[0m 停止")
	message.SetString(language.SimplifiedChinese, "ui line 2", "\x1b[7m\x1b[1m → \x1b[0m +10ms   \x1b[7m\x1b[1m Shift-→ \x1b[0m +50ms   \x1b[7m\x1b[1m Ctrl-→ \x1b[0m +100ms                ")
	message.SetString(language.SimplifiedChinese, "ADB devices:", "ADB设备：")
	message.SetString(language.SimplifiedChinese, "No authorized devices.", "没有授权的设备。")
	message.SetString(language.SimplifiedChinese, "No device has serial `%s`", "没有设备拥有序列号 `%s`")
	message.SetString(language.SimplifiedChinese, "Found device with serial number `%s`, but that device is not authorized.", "找到了拥有设备序列号`%s`的设备，但该设备并未授权。")
	message.SetString(language.SimplifiedChinese, "Selected device:", "选中的设备：")
	message.SetString(language.SimplifiedChinese, "Failed to connect to device:", "连接该设备失败：")
	message.SetString(language.SimplifiedChinese, "Recognized devices:", "已识别的设备：")
	message.SetString(language.SimplifiedChinese, "Song id and difficulty are both required", "歌曲ID和歌曲难度都需要提供")
	message.SetString(language.SimplifiedChinese, "Failed to find musicscore file:", "无法找到谱面文件：")
	message.SetString(language.SimplifiedChinese, "Musicscore not found", "未找到谱面")
	message.SetString(language.SimplifiedChinese, "Musicscore loaded:", "已加载谱面：")
	message.SetString(language.SimplifiedChinese, "Failed to load musicscore:", "加载谱面失败：")
	message.SetString(language.SimplifiedChinese, "Unknown backend: %q", "未知后端：%q")
	message.SetString(language.SimplifiedChinese, "%d pointers used.", "使用了%d个触点。")
	message.SetString(language.SimplifiedChinese, "[FATAL]", "\033[1;41m 错误 \033[0m")
	message.SetString(language.SimplifiedChinese, "[WARN]", "\033[1;45m 警告 \033[0m")
	message.SetString(language.SimplifiedChinese, "[INFO]", "\033[1;46m 信息 \033[0m")
	message.SetString(language.SimplifiedChinese, "[DEBUG]", "\033[1;44m 调试 \033[0m")
	message.SetString(language.SimplifiedChinese, "ssm: READY", "ssm: 已就绪")
	message.SetString(language.SimplifiedChinese, "ssm: Autoplaying %s (%s)", "ssm: 自动演奏 %s (%s)")
	message.SetString(language.SimplifiedChinese, "ssm: Autoplaying %s", "ssm: 自动演奏 %s")
}

func regEnglish() {
	message.SetString(language.English, "copyright info", `
Copyright (C) 2024, 2025 kvarenzn
License GPLv3+: GNU GPL version 3 or later <https://gnu.org/licenses/gpl.html>.
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.`)
	message.SetString(language.English, "usage.b", "Specify ssm backend, possible values")
	message.SetString(language.English, "usage.n", "Song ID")
	message.SetString(language.English, "usage.d", "Difficulty of song")
	message.SetString(language.English, "usage.e", "Extract assets from assets foler path")
	message.SetString(language.English, "usage.r", "Device orientation, options: `left` (↺, counter-clockwise), `right` (↻, clockwise). Note: ignored when using `adb` backend")
	message.SetString(language.English, "usage.p", "Custom chart path (if this is provided, song ID and difficulty will be ignored)")
	message.SetString(language.English, "usage.s", "Specify the device serial (if not provided, ssm will use the first device serial)")
	message.SetString(language.English, "usage.k", "Switch to PJSK mode")
	message.SetString(language.English, "usage.g", "Show debug info")
	message.SetString(language.English, "usage.v", "Show ssm's version information and exit")
	message.SetString(language.English, "ui line 0", "\x1b[7m\x1b[1m ENTER/SPACE \x1b[0m GO!!!!!")
	message.SetString(language.English, "ui line 1", "\x1b[7m\x1b[1m ← \x1b[0m -10ms   \x1b[7m\x1b[1m Shift-← \x1b[0m -50ms   \x1b[7m\x1b[1m Ctrl-← \x1b[0m -100ms   \x1b[7m\x1b[1m Ctrl-C \x1b[0m Stop")
	message.SetString(language.English, "ui line 2", "\x1b[7m\x1b[1m → \x1b[0m +10ms   \x1b[7m\x1b[1m Shift-→ \x1b[0m +50ms   \x1b[7m\x1b[1m Ctrl-→ \x1b[0m +100ms                ")
	message.SetString(language.English, "[FATAL]", "\033[1;41m FATAL \033[0m")
	message.SetString(language.English, "[WARN]", "\033[1;45m WARN \033[0m")
	message.SetString(language.English, "[INFO]", "\033[1;46m INFO \033[0m")
	message.SetString(language.English, "[DEBUG]", "\033[1;44m DEBUG \033[0m")
}

var LanguageString string

var P *message.Printer

func init() {
	regSimplifiedChinese()
	regEnglish()

	LanguageString, _ = GetSystemLocale()

	matcher := language.NewMatcher([]language.Tag{
		language.English,
		language.SimplifiedChinese,
	})

	lang, _, _ := matcher.Match(language.Make(LanguageString))

	P = message.NewPrinter(lang)
}
