# 常见问题

## 未找到谱面/Musicscore not found

首先，可能是因为没有做提取谱面这一步。请遵循使用说明中的描述提取谱面。

如果你确定完成了这一步，那么可能与你使用的操作系统有关。不过很抱歉，具体是什么原因导致的这个问题尚不明确。详见 [issues #10](https://github.com/kvarenzn/ssm/issues/10) 和 [issues #11](https://github.com/kvarenzn/ssm/issues/11) 。请检查ssm所在的文件夹内，有没有路径为`assets/star/forassetbundle/startapp/musicscore`的文件夹（对于Windows操作系统，路径为`assets\star\forassetbundle\startap\musicscore`）

无论有没有，请开一个issue，然后提供以下信息（格式无所谓，能表示清楚就行）：

- 有没有musicscore文件夹？
- 游戏服务器是？（日服/EN服/台服）
- 操作系统及其版本是？（提供大版本号即可，比如Windows 11）
- 解包得到的`extract.json`。这个文件会在你执行完`ssm -e`后，在ssm所在的文件夹内生成。请将其作为附件上传。

**但是，这并不意味着ssm完全不能使用，你可以尝试到bestdori.com上下载谱面：**

1. 浏览器打开bestdori.com
2. 在网页左栏展开`工具`折叠项，点击`数据包浏览器`
3. 你应该会看到五个文件夹：`jp`、`en`、`tw`、`cn`、`kr`，点击你玩的服务器对应的项（比如B服对应`cn`）
4. 点击`musicscore`
5. 你会看到一个长列表，列表每一项都是`musicscore+数字`的格式。后边的数字表示该数据包内最大的歌曲ID的值，并且一般10个歌曲打一个包。比如`musicscore10`包含id`1`到`10`的歌曲。按照这个规律点击要找的谱面所在的数据包。例如：EXIST的ID是325，那么你应该找`musicscore330`，点击
6. 点击`BMS`选项卡
7. 你会看到一堆`.txt`文件，它们都是以`{歌曲ID}_{歌曲简称}_{歌曲难度}.txt`格式命名的。点击你要找的谱面。比如EXIST的EXPERT难度就是`325_exist_expert.txt`
8. 之后会弹出一个下载对话框（某些浏览器，比如chrome，会直接下载，不会弹确认框），选择下载路径，确认下载
9. 找到刚刚下载的文件的路径，然后将这个路径传递给ssm（`ssm -p {路径}`）。注意要传递完整路径。例如刚才的EXIST EXPERT难度谱面下载到了`C:\Users\user\Downloads\`内，那么对应的命令是

```
ssm -p C:\Users\user\Downloads\325_exist_expert.txt
```

## Windows下`hid`后端无法识别设备

可能是驱动问题。可尝试卸载设备驱动并安装 [Google提供的驱动](https://dl.google.com/android/repository/usb_driver_r13-windows.zip)
详细请参考 Genymobile/scrcpy 项目的 [相关issue](https://github.com/Genymobile/scrcpy/issues/3654)

或直接尝试`adb`后端

## 使用`hid`后端报错`libusb: not found [code -5]`

很多情况都可能导致这个问题。需要按顺序排错。

1. 将`-g`选项传给ssm并重新运行。这时ssm会输出一些以`DEBUG`开始的行，注意其中的`Recognized devices:`，表示ssm识别到的设备。这是一个列表，列表中每一项都是一个设备的序列号(device serial)
2. 看看你的设备的序列号在不在其中。如果你不知道设备序列号，你可以按照下边的步骤查询：
   1. 确保`adb`已经正确安装，并且`adb version`可以正确运行，不报错
   2. 在你的游戏设备上开启USB调试，连接到计算机上
   3. 运行`adb devices`，这一命令会输出类似下边的内容
      ```
      List of devices attached
      XXXXXXXX	device
      ```
      前边的`XXXXXXXX`就是你的设备序列号
   4. 关闭游戏设备的USB调试选项，在计算机上运行`adb kill-server`
3. 根据实际情况分类：
   - 如果在
     - 看看你的设备序列号是不是列表中的第一个
       - 如果不是
         - 你只需要用`-s {设备序列号}`选项告诉ssm连接到你的设备即可解决问题
       - 如果是
         1. 请关闭计算机上任何正在与你设备通信的软件，比如各种手机管家、USB抓包软件和adb，然后重试
         2. 如果还是无法解决，你可以试试上一个问题的解决方案（安装驱动），再重试
         3. 如果还是不行，开启设备的USB调试，然后使用`adb`后端，重试
         4. 如果`adb`后端也无法工作，放弃
   - 如果不在
     1. 开启设备的USB调试后，重试最外层的第1、2步。如果这时候设备序列号在列表中，那么保持设备的USB调试开启，尝试最外层的第3步
     2. 如果还是没有，尝试上一个问题的解决方案，重试
     3. 如果还是不行，放弃

```mermaid
graph TD;
done[完成];
giveup[放弃];
id1[传递`-g`选项给ssm] --> id2[观察识别到的序列号列表];
id2 --> id3{你的设备在不在？};
id3 -->|在| id4{是不是第一个？};
id4 -->|是| id5[关闭冲突的软件，重试];
id5 --> id6{解决了吗？};
id6 -->|是|done;
id6 -->|否| id7[尝试安装驱动，重试];
id7 --> id8{解决了吗？};
id8 -->|是|done;
id8 -->|否| id9[使用`adb`后端，重试];
id9 --> id10{解决了吗？};
id10 -->|是|done;
id10 -->|否|giveup;
id4 -->|不是| id11[传递`-s`给ssm];
id11 --> done;
id3 -->|不在| id12[开启USB调试];
id12 --> id13{现在在吗？};
id13 -->|在| id4;
id13 -->|不在| id14[安装驱动];
id14 --> id15{现在在吗？};
id15 -->|在| id4;
id15 -->|不在|giveup;
```

## 断触问题

有时会出现 **所有音符都miss** 但ssm仍然在发送触控事件的情况

- 日服较常见，国际服极少出现
- 或许是反作弊机制（可能性较低）或客户端bug
- 对于自由演出模式，重开一般可解决
- 对于协力演出模式，暂时无法解决，请谨慎使用
