// Markdown 渲染封装。
//
// 为什么要单独抽出来并做「净化」：
// 文章正文是用户在后台写的 Markdown，渲染成 HTML 后会插进页面（v-html）。
// 如果直接渲染，作者可以写 <img src=x onerror="alert(1)"> 之类的 XSS 代码，
// 或者写 <script> 偷 token。DOMPurify 会把危险标签和属性全部剥掉。
import { Marked } from 'marked'
import { markedHighlight } from 'marked-highlight'
import DOMPurify from 'dompurify'

// ❗ 不要写 import hljs from 'highlight.js'
// 那会把 190 多种语言的语法定义全打进包里（构建产物多出约 1MB）。
// 按需引入核心 + 本项目实际用到的语言，体积从 ~1MB 降到 ~40KB。
import hljs from 'highlight.js/lib/core'
import bash from 'highlight.js/lib/languages/bash'
import css from 'highlight.js/lib/languages/css'
import go from 'highlight.js/lib/languages/go'
import javascript from 'highlight.js/lib/languages/javascript'
import json from 'highlight.js/lib/languages/json'
import markdown from 'highlight.js/lib/languages/markdown'
import python from 'highlight.js/lib/languages/python'
import sql from 'highlight.js/lib/languages/sql'
import typescript from 'highlight.js/lib/languages/typescript'
import xml from 'highlight.js/lib/languages/xml'
import yaml from 'highlight.js/lib/languages/yaml'

hljs.registerLanguage('bash', bash)
hljs.registerLanguage('sh', bash)
hljs.registerLanguage('shell', bash)
hljs.registerLanguage('css', css)
hljs.registerLanguage('go', go)
hljs.registerLanguage('golang', go)
hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('js', javascript)
hljs.registerLanguage('json', json)
hljs.registerLanguage('markdown', markdown)
hljs.registerLanguage('md', markdown)
hljs.registerLanguage('python', python)
hljs.registerLanguage('py', python)
hljs.registerLanguage('sql', sql)
hljs.registerLanguage('typescript', typescript)
hljs.registerLanguage('ts', typescript)
hljs.registerLanguage('html', xml)
hljs.registerLanguage('xml', xml)
hljs.registerLanguage('vue', xml) // highlight.js 没有 vue 语法，用 xml 兜底

// 属性值转义：markdown 里的 alt/title 是作者可控文本，
// 直接拼进 HTML 属性会被 " 逃逸出标签（DOMPurify 虽然会兜底，但没必要让它兜）。
function escapeAttr(text = '') {
  return String(text)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

// 图片渲染：<figure> + alt 当图注。
//
// 这个博客的插图都是「讲解图」而不是配图，图本身要能独立说清一件事。
// 所以约定 alt 不写「图1」，而是写一句完整结论（20-45 字），
// 渲染时把它同时用作无障碍描述和可见图注——一张图配一句话，扫一眼就懂。
// 外层再包一个链接，点开可以在新标签页看原始 SVG 大图。
function renderFigure({ href, title, text }) {
  const alt = (text || '').trim()
  const src = escapeAttr(href || '')
  const altAttr = escapeAttr(alt)
  const titleAttr = title ? ` title="${escapeAttr(title)}"` : ''

  return [
    '<figure class="md-figure">',
    `<a class="md-figure-link" href="${src}" target="_blank" rel="noopener">`,
    `<img src="${src}" alt="${altAttr}"${titleAttr} loading="lazy" decoding="async" />`,
    '</a>',
    alt ? `<figcaption>${escapeAttr(alt)}</figcaption>` : '',
    '</figure>',
  ].join('')
}

// 代码高亮。
//
// 注意：marked v5 之后删掉了 setOptions({ highlight })，老教程里的写法已经失效，
// 必须用 marked-highlight 扩展挂上去。这也是为什么之前代码块渲染出来了却没有颜色。
const marked = new Marked(
  markedHighlight({
    langPrefix: 'hljs language-', // 生成的 class 要能被 highlight.js 的主题样式命中
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
  {
    breaks: true, // 单个换行也当 <br>，符合博客写作习惯
    gfm: true, // 支持表格、删除线等 GitHub 风格
    renderer: {
      image: renderFigure,
    },
  },
)

/**
 * 把 Markdown 文本渲染成可以安全插入页面的 HTML
 * @param {string} raw Markdown 原文
 * @returns {string} 净化后的 HTML
 */
export function renderMarkdown(raw) {
  if (!raw) return ''
  const html = marked.parse(raw)
  // marked 会把独占一段的图片包进 <p>，但 <figure> 不能嵌在 <p> 里。
  // 浏览器虽然会自动纠正，但会留下一个空 <p> 撑出多余间距，这里先摘掉。
  const normalized = html.replace(/<p>\s*(<figure[\s\S]*?<\/figure>)\s*<\/p>/g, '$1')
  // 允许的标签/属性由 DOMPurify 默认白名单决定，足够覆盖博客排版
  return DOMPurify.sanitize(normalized, {
    ADD_ATTR: ['target', 'rel'],
  })
}
