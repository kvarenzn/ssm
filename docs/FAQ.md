# 常见问题
## Windows下`hid`后端无法识别设备
可能是驱动问题。可尝试卸载设备驱动并安装 [Google提供的驱动](https://dl.google.com/android/repository/usb_driver_r13-windows.zip)
详细请参考 Genymobile/scrcpy 项目的 [相关issue](https://github.com/Genymobile/scrcpy/issues/3654)

或直接尝试`adb`后端

## 断触问题
有时会出现 **所有音符都miss** 但ssm仍然在发送触控事件的情况
- 日服较常见，国际服极少出现
- 或许是反作弊机制（可能性较低）或客户端bug
- 对于自由演出模式，重开一般可解决
- 对于协力演出模式，暂时无法解决，请谨慎使用
