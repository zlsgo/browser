# Browser 浏览器自动化库

🚀 强大且易用的 Go 浏览器自动化库，基于 rod 构建，提供流畅的 API 和链式调用支持。

## ✨ 特性

- 🏗️ **Builder 模式配置** - 简洁直观的浏览器配置
- 🔗 **流式链式调用** - 优雅的操作链，减少样板代码
- 📋 **智能错误处理** - 详细的错误信息和重试机制
- 🎯 **预设配置** - 开发、生产、测试等环境预设
- 📱 **设备模拟** - 内置移动设备、平板等模拟
- 🕵️ **隐形模式** - 反检测浏览器指纹
- ⚡ **高性能** - 基于成熟的 rod 库构建

## 🚀 快速开始

### 安装

```bash
go get github.com/sohaha/zlsgo/browser
```

### 基础使用

```go
package main

import (
    "fmt"
    "github.com/sohaha/zlsgo/browser"
)

func main() {
    // 使用新的 Builder API
    browserInstance := browser.NewBrowser().
        WithHeadless(false).
        WithTimeout(30 * time.Second).
        MustBuild()
    defer browserInstance.Close()

    // 流式页面操作
    err := browserInstance.Open("https://example.com", func(page *browser.Page) error {
        return page.Chain().
            WaitForLoad().
            ClickOn("#login-button").
            FillForm(map[string]string{
                "#username": "user@example.com",
                "#password": "password123",
            }).
            SubmitForm().
            WaitForText("登录成功").
            Error()
    })

    if err != nil {
        fmt.Printf("操作失败: %v\n", err)
    }
}
```

## 📖 API 文档

### 🏗️ Builder 模式创建浏览器

#### 基础配置

```go
// 创建浏览器构建器
browser := browser.NewBrowser().
    WithHeadless(true).                    // 无头模式
    WithUserAgent("CustomBot/1.0").        // 自定义 User-Agent
    WithTimeout(30 * time.Second).         // 超时设置
    WithProxy("http://proxy:8080").        // 代理设置
    MustBuild()
```

#### 预设配置

```go
// 开发环境 - 可见界面，开启调试
devBrowser := browser.NewBrowser().
    Preset(browser.PresetDevelopment).
    MustBuild()

// 生产环境 - 无头，隐形，内存优化
prodBrowser := browser.NewBrowser().
    Preset(browser.PresetProduction).
    MustBuild()

// 测试环境 - 快速，隐身
testBrowser := browser.NewBrowser().
    Preset(browser.PresetTesting).
    MustBuild()

// 隐形模式 - 反检测
stealthBrowser := browser.NewBrowser().
    Preset(browser.PresetStealth).
    MustBuild()
```

#### 设备模拟

```go
// 移动设备模拟
mobileBrowser := browser.NewBrowser().
    WithMobileDevice().
    MustBuild()

// 平板设备模拟
tabletBrowser := browser.NewBrowser().
    WithTabletDevice().
    MustBuild()

// 自定义设备
customBrowser := browser.NewBrowser().
    WithDevice(devices.IPhoneX).
    MustBuild()
```

### 🔗 流式页面操作

#### 基础页面操作

```go
err := page.Chain().
    NavigateTo("https://example.com").     // 导航到页面
    WaitForLoad().                         // 等待加载完成
    WaitForElement("#content").            // 等待元素出现
    ScrollTo("#footer").                   // 滚动到元素
    Error()                               // 获取错误
```

#### 表单操作

```go
// 单个输入
err := page.Chain().
    TypeInto("#search", "Go语言").
    PressEnter().
    Error()

// 批量表单填充
err := page.Chain().
    FillForm(map[string]string{
        "#name":     "张三",
        "#email":    "zhangsan@example.com",
        "#message":  "这是测试消息",
    }).
    SubmitForm().
    Error()
```

#### 元素交互

```go
// 链式元素操作
element := page.Chain().
    FindElement("#button").
    Hover().                    // 悬停
    Click().                    // 点击
    WaitVisible().              // 等待可见
    MustComplete()              // 必须成功
```

#### 等待条件

```go
// 多种等待条件
err := page.Chain().
    WaitForLoad().                         // 等待页面加载
    WaitForStable().                       // 等待 DOM 稳定
    WaitForElement(".result").             // 等待元素出现
    WaitForText("操作完成").                 // 等待文本出现
    Error()
```

### 🎯 便利方法

```go
// 快速填充表单
err := browser.QuickFill("https://example.com/form", map[string]string{
    "#username": "admin",
    "#password": "password",
}, "#submit")

// 快速点击
err := browser.QuickClick("https://example.com", "#download")

// 快速搜索
err := browser.QuickSearch("https://example.com", "#search", "关键词")
```

### 📋 错误处理

#### 链式错误处理

```go
// 使用 Error() 进行错误检查
err := page.Chain().
    NavigateTo("https://example.com").
    ClickOn("#button").
    Error()

if err != nil {
    // 检查错误类型
    if browser.IsTimeoutError(err) {
        fmt.Println("操作超时")
    }
    if browser.IsRetryableError(err) {
        fmt.Println("可以重试")
    }
}

// 使用 MustComplete() (失败时 panic)
page.Chain().
    NavigateTo("https://example.com").
    ClickOn("#button").
    MustComplete()
```

#### 重试机制

```go
// 使用默认重试策略
err := browser.WithRetry(browser.DefaultRetry, func() error {
    return page.Chain().
        NavigateTo("https://unstable-site.com").
        Error()
})

// 自定义重试策略
customRetry := browser.ErrorRetry{
    MaxAttempts: 5,
    Delay:       2 * time.Second,
    Backoff:     1.5,
}

err = browser.WithRetry(customRetry, func() error {
    return page.ClickOn("#flaky-button")
})
```

## 🔧 高级用法

### 条件操作

```go
// 检查元素是否存在
has, element := page.HasElement("#login-form")
if has {
    // 执行登录
    err := element.Chain().
        FindChild("#username").Type("admin").
        FindChild("#password").Type("password").
        FindChild("#submit").Click().
        Error()
}
```

### 自定义操作

```go
// 在链式调用中执行自定义逻辑
err := page.Chain().
    WaitForLoad().
    Execute(func(page *browser.Page) error {
        // 自定义操作
        return page.page.KeyActions().
            Press(input.ControlLeft).
            Type(input.KeyA).
            Do()
    }).
    Type("新内容").
    Error()
```

### 克隆配置

```go
// 创建基础配置
baseBuilder := browser.NewBrowser().
    Preset(browser.PresetTesting).
    WithTimeout(30 * time.Second)

// 克隆并修改
browser1 := baseBuilder.Clone().
    WithUserAgent("Bot1/1.0").
    MustBuild()

browser2 := baseBuilder.Clone().
    WithUserAgent("Bot2/1.0").
    WithHeadless(false).
    MustBuild()
```

## 🆚 新旧 API 对比

### 旧 API (仍然支持)

```go
browser, err := browser.New(func(o *browser.Options) {
    o.Headless = false
    o.UserAgent = "Chrome/120.0"
    o.Timeout = 30 * time.Second
    o.Flags = map[string]string{"no-sandbox": ""}
})

err = browser.Open("https://example.com", func(page *browser.Page) error {
    element, err := page.Element("#input")
    if err != nil { return err }
    return element.InputText("hello")
})
```

### 新 API (推荐)

```go
browser := browser.NewBrowser().
    WithHeadless(false).
    WithUserAgent("Chrome/120.0").
    WithTimeout(30 * time.Second).
    WithFlag("no-sandbox", "").
    MustBuild()

err := browser.Open("https://example.com", func(page *browser.Page) error {
    return page.Chain().
        FindElement("#input").
        Type("hello").
        Error()
})
```

## 🛠️ 最佳实践

### 1. 错误处理

```go
// 推荐：使用 Error() 进行显式错误处理
err := page.Chain().
    NavigateTo(url).
    Error()
if err != nil {
    log.Printf("导航失败: %v", err)
    // 处理错误或重试
}

// 开发调试：使用 MustComplete() 快速失败
page.Chain().
    NavigateTo(url).
    MustComplete() // 失败时 panic，便于调试
```

### 2. 性能优化

```go
// 生产环境使用隐形模式
browser := browser.NewBrowser().
    Preset(browser.PresetProduction).
    WithStealth(true).
    MustBuild()

// 合理设置超时
browser := browser.NewBrowser().
    WithTimeout(10 * time.Second).  // 全局超时
    MustBuild()

// 页面级超时
page.WithTimeout(5 * time.Second, func(p *browser.Page) error {
    return p.Chain().FindElement("#fast-element").Click().Error()
})
```

### 3. 资源管理

```go
// 确保浏览器正确关闭
browser := browser.NewBrowser().MustBuild()
defer browser.Close()  // 重要：释放资源

// 页面资源清理
err := browser.Open(url, func(page *browser.Page) error {
    defer func() {
        // 页面特定的清理工作
    }()
    return page.Chain()./* 操作 */.Error()
})
```

## 🐛 故障排除

### 常见问题

1. **浏览器启动失败**
   ```bash
   # Linux 安装依赖
   sudo apt-get install -y libnss3 libxss1 libasound2t64 libxtst6 libgtk-3-0 libgbm1
   ```

2. **元素找不到**
   ```go
   // 增加等待时间
   err := page.Chain().
       WaitForElement("#slow-element", 10*time.Second).
       Error()
   ```

3. **操作太快**
   ```go
   // 添加延迟
   err := page.Chain().
       Click("#button").
       Sleep(2 * time.Second).
       Error()
   ```

## 📚 更多资源

- [完整 API 文档](https://docs.73zls.com/zlsgo/#/c9e16ee075214cf2a9df1f7093aece58)

