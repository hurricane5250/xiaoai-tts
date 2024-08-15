# go-xiaoai
基于golang 的小爱音箱api 控制接口。简单、易用。


## 安装

```bash
go get github.com/YoungBreezeM/xiaoai-tts
```

## 例子

```golang

//新建小爱音响客户端
xiaoai, err := xiaoaitts.New("xxxx", "xxxx")
if err != nil {
	return
}


//控制小爱说话
xiaoai.Say("hello")

//获取账号下设备
xiaoai.GetDevice() []models.DeviceInfo

//使用设备
xiaoai.UseDevice(index int16)

//控制说话
xiaoai.Say(text string)

//播放音量
xiaoai.SetVolume(volume int8)

//获取音量
xiaoai.GetVolume() string

//播放
xiaoai.Play()

//暂停
xiaoai.Pause()

//上一首
xiaoai.Prev()

//下一首
xiaoai.Next()

//
xiaoai.TogglePlayState()

//获取音响状态
xiaoai.GetStatus() *models.Info

//通过http 路径播放音乐
xiaoai.PlayUrl(url string)

//获取对话记录
xiaoai.GetLastAsk()
```

