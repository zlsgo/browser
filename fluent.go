package browser

import (
	"fmt"
	"time"

	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
)

// FluentPage 流式页面操作接口
type FluentPage struct {
	page *Page
	err  error
}

// FluentElement 流式元素操作接口
type FluentElement struct {
	element *Element
	err     error
}

// Chain 为Page添加链式调用支持
func (p *Page) Chain() *FluentPage {
	return &FluentPage{page: p}
}

// Fluent 为Element添加链式调用支持
func (e *Element) Chain() *FluentElement {
	return &FluentElement{element: e}
}

// ===== FluentPage 方法 =====

// NavigateTo 导航到指定URL
func (fp *FluentPage) NavigateTo(url string) *FluentPage {
	if fp.err != nil {
		return fp
	}
	fp.err = fp.page.NavigateWaitLoad(url)
	return fp
}

// WaitForLoad 等待页面加载完成
func (fp *FluentPage) WaitForLoad() *FluentPage {
	if fp.err != nil {
		return fp
	}
	fp.err = fp.page.WaitLoad()
	return fp
}

// WaitForStable 等待DOM稳定
func (fp *FluentPage) WaitForStable(diff ...float64) *FluentPage {
	if fp.err != nil {
		return fp
	}
	fp.err = fp.page.WaitDOMStable(diff...)
	return fp
}

// WaitForElement 等待元素出现
func (fp *FluentPage) WaitForElement(selector string, timeout ...time.Duration) *FluentPage {
	if fp.err != nil {
		return fp
	}

	page := fp.page
	if len(timeout) > 0 {
		page = page.Timeout(timeout[0])
	}

	_, fp.err = page.Element(selector)
	return fp
}

// FindElement 查找元素并返回流式元素操作
func (fp *FluentPage) FindElement(selector string) *FluentElement {
	if fp.err != nil {
		return &FluentElement{err: fp.err}
	}

	element, err := fp.page.Element(selector)
	return &FluentElement{element: element, err: err}
}

// ClickOn 点击指定选择器的元素
func (fp *FluentPage) ClickOn(selector string) *FluentPage {
	if fp.err != nil {
		return fp
	}

	element, err := fp.page.Element(selector)
	if err != nil {
		fp.err = err
		return fp
	}

	fp.err = element.Click()
	return fp
}

// TypeInto 在指定选择器的元素中输入文本
func (fp *FluentPage) TypeInto(selector, text string) *FluentPage {
	if fp.err != nil {
		return fp
	}

	element, err := fp.page.Element(selector)
	if err != nil {
		fp.err = err
		return fp
	}

	fp.err = element.InputText(text)
	return fp
}

// FillForm 批量填充表单
func (fp *FluentPage) FillForm(data map[string]string) *FluentPage {
	if fp.err != nil {
		return fp
	}

	for selector, value := range data {
		element, err := fp.page.Element(selector)
		if err != nil {
			fp.err = fmt.Errorf("找不到元素 %s: %w", selector, err)
			return fp
		}

		if err := element.InputText(value, true); err != nil {
			fp.err = fmt.Errorf("输入文本到 %s 失败: %w", selector, err)
			return fp
		}
	}

	return fp
}

// SubmitForm 提交表单
func (fp *FluentPage) SubmitForm(formSelector ...string) *FluentPage {
	if fp.err != nil {
		return fp
	}

	selector := "form"
	if len(formSelector) > 0 {
		selector = formSelector[0]
	}

	form, err := fp.page.Element(selector)
	if err != nil {
		fp.err = err
		return fp
	}

	// 查找提交按钮或直接提交表单
	submitBtn, err := form.Element("input[type='submit'], button[type='submit'], button:not([type])")
	if err != nil {
		// 如果找不到提交按钮，尝试按Enter键
		fp.err = fp.page.page.KeyActions().Type(input.Enter).Do()
	} else {
		fp.err = submitBtn.Click()
	}

	return fp
}

// ScrollTo 滚动到指定元素
func (fp *FluentPage) ScrollTo(selector string) *FluentPage {
	if fp.err != nil {
		return fp
	}

	element, err := fp.page.Element(selector)
	if err != nil {
		fp.err = err
		return fp
	}

	fp.err = fp.page.NaturalScroll(element, time.Second)
	return fp
}

// WaitForText 等待页面包含指定文本
func (fp *FluentPage) WaitForText(text string, timeout ...time.Duration) *FluentPage {
	if fp.err != nil {
		return fp
	}

	page := fp.page
	if len(timeout) > 0 {
		page = page.Timeout(timeout[0])
	}

	_, fp.err = page.Search(text)
	return fp
}

// Sleep 等待指定时间
func (fp *FluentPage) Sleep(duration time.Duration) *FluentPage {
	if fp.err != nil {
		return fp
	}
	time.Sleep(duration)
	return fp
}

// Execute 执行自定义函数
func (fp *FluentPage) Execute(fn func(*Page) error) *FluentPage {
	if fp.err != nil {
		return fp
	}
	fp.err = fn(fp.page)
	return fp
}

// ===== FluentElement 方法 =====

// Click 点击元素
func (fe *FluentElement) Click() *FluentElement {
	if fe.err != nil {
		return fe
	}
	fe.err = fe.element.Click()
	return fe
}

// DoubleClick 双击元素
func (fe *FluentElement) DoubleClick() *FluentElement {
	if fe.err != nil {
		return fe
	}
	fe.err = fe.element.element.Click(proto.InputMouseButtonLeft, 2)
	return fe
}

// RightClick 右键点击元素
func (fe *FluentElement) RightClick() *FluentElement {
	if fe.err != nil {
		return fe
	}
	fe.err = fe.element.Click(proto.InputMouseButtonRight)
	return fe
}

// Type 输入文本
func (fe *FluentElement) Type(text string) *FluentElement {
	if fe.err != nil {
		return fe
	}
	fe.err = fe.element.InputText(text)
	return fe
}

// Clear 清空元素内容
func (fe *FluentElement) Clear() *FluentElement {
	if fe.err != nil {
		return fe
	}
	fe.err = fe.element.element.SelectAllText()
	if fe.err == nil {
		fe.err = fe.element.element.Input("")
	}
	return fe
}

// ClearAndType 清空并输入新文本
func (fe *FluentElement) ClearAndType(text string) *FluentElement {
	return fe.Clear().Type(text)
}

// PressEnter 按回车键
func (fe *FluentElement) PressEnter() *FluentElement {
	if fe.err != nil {
		return fe
	}
	fe.err = fe.element.InputEnter()
	return fe
}

// PressKey 按指定键
func (fe *FluentElement) PressKey(keys ...input.Key) *FluentElement {
	if fe.err != nil {
		return fe
	}
	fe.err = fe.element.page.page.KeyActions().Press(keys...).Do()
	return fe
}

// Focus 聚焦元素
func (fe *FluentElement) Focus() *FluentElement {
	if fe.err != nil {
		return fe
	}
	fe.err = fe.element.Focus()
	return fe
}

// Hover 悬停在元素上
func (fe *FluentElement) Hover() *FluentElement {
	if fe.err != nil {
		return fe
	}
	fe.err = fe.element.element.Hover()
	return fe
}

// ScrollIntoView 滚动到元素可见位置
func (fe *FluentElement) ScrollIntoView() *FluentElement {
	if fe.err != nil {
		return fe
	}
	fe.err = fe.element.element.ScrollIntoView()
	return fe
}

// WaitVisible 等待元素可见
func (fe *FluentElement) WaitVisible() *FluentElement {
	if fe.err != nil {
		return fe
	}
	fe.err = fe.element.element.WaitVisible()
	return fe
}

// WaitInvisible 等待元素不可见
func (fe *FluentElement) WaitInvisible() *FluentElement {
	if fe.err != nil {
		return fe
	}
	fe.err = fe.element.element.WaitInvisible()
	return fe
}

// FindChild 查找子元素
func (fe *FluentElement) FindChild(selector string) *FluentElement {
	if fe.err != nil {
		return &FluentElement{err: fe.err}
	}

	child, err := fe.element.Element(selector)
	return &FluentElement{element: child, err: err}
}

// Execute 在元素上执行自定义函数
func (fe *FluentElement) Execute(fn func(*Element) error) *FluentElement {
	if fe.err != nil {
		return fe
	}
	fe.err = fn(fe.element)
	return fe
}

// ===== 结果获取方法 =====

// Error 获取操作中的错误
func (fp *FluentPage) Error() error {
	return fp.err
}

// Page 获取页面实例
func (fp *FluentPage) Page() *Page {
	return fp.page
}

// MustComplete 必须成功完成，否则panic
func (fp *FluentPage) MustComplete() *Page {
	if fp.err != nil {
		panic(fp.err)
	}
	return fp.page
}

// Complete 完成操作链，返回结果
func (fp *FluentPage) Complete() (*Page, error) {
	return fp.page, fp.err
}

// Error 获取操作中的错误
func (fe *FluentElement) Error() error {
	return fe.err
}

// Element 获取元素实例
func (fe *FluentElement) Element() *Element {
	return fe.element
}

// MustComplete 必须成功完成，否则panic
func (fe *FluentElement) MustComplete() *Element {
	if fe.err != nil {
		panic(fe.err)
	}
	return fe.element
}

// Complete 完成操作链，返回结果
func (fe *FluentElement) Complete() (*Element, error) {
	return fe.element, fe.err
}

// ===== 便利方法 =====

// QuickFill 快速填充表单的便利方法
func (b *Browser) QuickFill(url string, formData map[string]string, submitSelector ...string) error {
	return b.Open(url, func(p *Page) error {
		chain := p.Chain().
			WaitForLoad().
			FillForm(formData)

		if len(submitSelector) > 0 {
			chain = chain.ClickOn(submitSelector[0])
		} else {
			chain = chain.SubmitForm()
		}

		return chain.Error()
	})
}

// QuickClick 快速点击的便利方法
func (b *Browser) QuickClick(url, selector string) error {
	return b.Open(url, func(p *Page) error {
		return p.Chain().
			WaitForLoad().
			ClickOn(selector).
			Error()
	})
}

// QuickSearch 快速搜索的便利方法
func (b *Browser) QuickSearch(url, searchSelector, query string) error {
	return b.Open(url, func(p *Page) error {
		return p.Chain().
			WaitForLoad().
			TypeInto(searchSelector, query).
			PressEnter().
			Error()
	})
}

// PressEnter 为FluentPage添加回车键支持
func (fp *FluentPage) PressEnter() *FluentPage {
	if fp.err != nil {
		return fp
	}
	fp.err = fp.page.page.KeyActions().Type(input.Enter).Do()
	return fp
}