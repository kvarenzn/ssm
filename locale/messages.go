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

var langTomoriZhHans = language.MustParse("x-tomori-zh-hans")

func regTomoriZh() {
	message.SetString(langTomoriZhHans, "copyright info", `
那个...ssm...是从2024年开始，到2025年为止，由kvarenzn...一点点拼凑起来的东西。
然后是...关于它的约定...

它遵循着 GNU 通用公共许可证...第3版...或者之后更新的版本...
更多的文字...在这里：<https://gnu.org/licenses/gpl.html>

这个软件...是“自由”的。
这意味着...您可以把它拆开，看看里面的构造；也可以把它变成新的样子，分享给其他人...
就像...一起拼凑迷失的拼图一样。

但是...
我...我们...无法向您承诺任何事。
它就像是被放在这里...“就这样存在着”。
在法律允许的范围内...这就是它的全部了。`)
	message.SetString(langTomoriZhHans, "Usage of %s:", "关于 ssm 的用法... 那个... 我想这样解释可能会清楚一点...")
	message.SetString(langTomoriZhHans, "usage.b", "选择...和设备对话的方式。可以试试`hid`，或者...`adb`")
	message.SetString(langTomoriZhHans, "usage.n", "想要完成的...那首歌的编号")
	message.SetString(langTomoriZhHans, "usage.d", "告诉ssm...想要挑战的那首歌的难度")
	message.SetString(langTomoriZhHans, "usage.e", "从那里...把我们需要的东西...找出来")
	message.SetString(langTomoriZhHans, "usage.r", "告诉ssm...设备当前的方向。可以是`left` （↺, 逆时针）...或者`right`（↻, 顺时针）。（如果与设备对话的方式是`adb`...这个选项...就不起作用了）")
	message.SetString(langTomoriZhHans, "usage.p", "也可以...直接告诉ssm到哪里去找谱面（如果告诉了这个...编号和难度...ssm就不会去看了）")
	message.SetString(langTomoriZhHans, "usage.s", "如果连接的设备很多...可以用这个指定其中一个（如果不告诉ssm...ssm会使用找到的第一个...希望没有选错）")
	message.SetString(langTomoriZhHans, "usage.g", "如果加上这个...ssm就会展示更多内心的、细微的纠结和迷茫。也许能知道哪里出了问题...")
	message.SetString(langTomoriZhHans, "usage.v", "让ssm...简单地介绍自己")
	message.SetString(langTomoriZhHans, "ssm version: %s", "ssm 的版本是... %s")
	message.SetString(langTomoriZhHans, "(unknown)", "(unknown) ...あのん...小爱...一辈子")
	message.SetString(langTomoriZhHans, "To use adb as the backend, the third-party component `scrcpy-server` (version %s) is required.", "如果想借助`adb`的力量...还需要一个叫`scrcpy-server` (%s 版) 的组件...")
	message.SetString(langTomoriZhHans, "This component is developed by Genymobile and licensed under Apache License 2.0.", "它是由Genymobile团队制作的...遵循着名为Apache License 2.0的约定...")
	message.SetString(langTomoriZhHans, "Please download it from the official release page and place it in the same directory as `ssm.exe`.", "可以从他们发布它的地方找到...然后...请把它放到ssm所在的文件夹里...")
	message.SetString(langTomoriZhHans, "Download link:", "就是这里：")
	message.SetString(langTomoriZhHans, "Alternatively, ssm can automatically handle this process for you.", "或者...那个...也可以交给ssm...帮您把这些事情做完...")
	message.SetString(langTomoriZhHans, "Proceed with automatic download? [Y/n]: ", "需要...ssm的帮助吗？ [Y/n]: ")
	message.SetString(langTomoriZhHans, "Failed to get input from user:", "抱歉...没能听清您的回答...：")
	message.SetString(langTomoriZhHans, "`scrcpy-server` is required. To use `adb` as the backend, you should download it manually.", "没能找到`scrcpy-server`...如果您想借助`adb`的力量...可能只能由您亲自把它带回来了...")
	message.SetString(langTomoriZhHans, "Downloading... Please wait.", "正在下载...请再等一会...一会就好...")
	message.SetString(langTomoriZhHans, "Failed to download `scrcpy-server`.", "失败了...没能拿到`scrcpy-server`...")
	message.SetString(langTomoriZhHans, "You may try again later, download it manually, or use `hid` backend instead.", "那个...或许可以之后再试试...或者，不用ssm的帮助，您亲自去把它带回来...也可以选择用`hid`的方式...")
	message.SetString(langTomoriZhHans, "Failed to calculate sha256 of `scrcpy-server`:", "没能计算出`scrcpy-server`的验证信息...")
	message.SetString(langTomoriZhHans, "Checksum mismatch. Please try again later.", "拿到的东西...和预期的对不上...请之后再试试吧")
	message.SetString(langTomoriZhHans, "Failed to save `scrcpy-server` to disk:", "抱歉...拿到了`scrcpy-server`...但没能好好地珍藏起来...：")
	message.SetString(langTomoriZhHans, "Failed to locate server file:", "抱歉...哪里都找不到`scrcpy-server`...：")
	message.SetString(langTomoriZhHans, "Failed to read the content of `scrcpy-server`:", "找到了`scrcpy-server`...但看不到它的内容...：")
	message.SetString(langTomoriZhHans, "Checksum mismatch. File may be corrupted.", "抱歉...跟预想的对不上...可能是在收藏的过程中损坏了...")
	message.SetString(langTomoriZhHans, "Failed to load music jacket: %s", "没能把那张回忆的封面...找出来...：%s")
	message.SetString(langTomoriZhHans, "ui line 0", "\x1b[7m\x1b[1m 回车/空格 \x1b[0m 来吧 一同奏响吧")
	message.SetString(langTomoriZhHans, "Offset: %d ms", "偏移：%d 毫秒")
	message.SetString(langTomoriZhHans, "Failed to get key from stdin: %s", "没能接收到您的信号...：%s")
	message.SetString(langTomoriZhHans, "ui line 1", "\x1b[7m\x1b[1m ← \x1b[0m -10ms   \x1b[7m\x1b[1m Shift-← \x1b[0m -50ms   \x1b[7m\x1b[1m Ctrl-← \x1b[0m -100ms   \x1b[7m\x1b[1m Ctrl-C \x1b[0m 停止")
	message.SetString(langTomoriZhHans, "ui line 2", "\x1b[7m\x1b[1m → \x1b[0m +10ms   \x1b[7m\x1b[1m Shift-→ \x1b[0m +50ms   \x1b[7m\x1b[1m Ctrl-→ \x1b[0m +100ms                ")
	message.SetString(langTomoriZhHans, "ADB devices:", "通过ADB找到的设备...")
	message.SetString(langTomoriZhHans, "No authorized devices.", "没有...愿意回应ssm的设备...")
	message.SetString(langTomoriZhHans, "No device has serial `%s`", "找不到名叫`%s`的设备...")
	message.SetString(langTomoriZhHans, "Found device with serial number `%s`, but that device is not authorized.", "虽然找到了`%s`...但它好像...不想理ssm...")
	message.SetString(langTomoriZhHans, "Selected device:", "就决定是...这个了：")
	message.SetString(langTomoriZhHans, "Failed to connect to device:", "没能跟它...建立联系...：")
	message.SetString(langTomoriZhHans, "Recognized devices:", "这些是...ssm找到的设备：")
	message.SetString(langTomoriZhHans, "Song id and difficulty are both required", "那个...还需要告诉ssm...是哪一首歌...和它的难度...")
	message.SetString(langTomoriZhHans, "Failed to find musicscore file:", "找不到乐谱文件...：")
	message.SetString(langTomoriZhHans, "Musicscore not found", "哪里都找不到乐谱...")
	message.SetString(langTomoriZhHans, "Musicscore loaded:", "乐谱...ssm读懂了...：")
	message.SetString(langTomoriZhHans, "Failed to load musicscore:", "抱歉...ssm看不懂这份乐谱...：")
	message.SetString(langTomoriZhHans, "Unknown backend: %q", "不清楚...该怎么用`%q`这种方式...建立联系...")
	message.SetString(langTomoriZhHans, "%d pointers used.", "ssm...需要用%d根手指完成这首歌...")
	message.SetString(langTomoriZhHans, "[FATAL]", "\033[1;41m 无法继续了... \033[0m")
	message.SetString(langTomoriZhHans, "[WARN]", "\033[1;45m 有点担心... \033[0m")
	message.SetString(langTomoriZhHans, "[INFO]", "\033[1;46m 请注意... \033[0m")
	message.SetString(langTomoriZhHans, "[DEBUG]", "\033[1;44m（碎碎念）\033[0m")
	message.SetString(langTomoriZhHans, "ssm: READY", "ssm: ...准备好了！")
	message.SetString(langTomoriZhHans, "ssm: Autoplaying %s (%s)", "ssm: ...正在全力演奏 %s (%s)")
	message.SetString(langTomoriZhHans, "ssm: Autoplaying %s", "ssm: ...正在全力演奏 %s")
}

var LanguageString string

var P *message.Printer

func init() {
	regSimplifiedChinese()
	regEnglish()
	regTomoriZh()

	LanguageString, _ = GetSystemLocale()

	matcher := language.NewMatcher([]language.Tag{
		language.English,
		language.SimplifiedChinese,
		langTomoriZhHans,
	})

	lang, _, _ := matcher.Match(language.Make(LanguageString))

	P = message.NewPrinter(lang)
}
