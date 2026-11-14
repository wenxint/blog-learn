#!/usr/bin/env python3
"""把 docs/articles/*.md 导入博客数据库。

每篇文章用 frontmatter 描述元数据，正文是 markdown：

    ---
    title: 标题
    summary: 摘要
    category: 建站记
    tags: 架构, 后端
    status: 1
    is_top: 1
    ---
    正文从这里开始……

用法：
    python3 scripts/import_articles.py --password 你的后台密码
    python3 scripts/import_articles.py --password xxx --base-url http://127.0.0.1:8090
    python3 scripts/import_articles.py --password xxx --dry-run

脚本是幂等的：按标题匹配，已存在就更新，不存在才创建。
"""

import argparse
import hashlib
import json
import re
import sys
import urllib.error
import urllib.request
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
ARTICLES_DIR = ROOT / "docs" / "articles"


def parse_frontmatter(text):
    """拆出 frontmatter 和正文。没有 frontmatter 时返回空元数据。"""
    m = re.match(r"^---\n(.*?)\n---\n?(.*)$", text, re.S)
    if not m:
        return {}, text
    meta, body = {}, m.group(2)
    for line in m.group(1).splitlines():
        if not line.strip() or ":" not in line:
            continue
        k, v = line.split(":", 1)
        meta[k.strip()] = v.strip()
    return meta, body.lstrip("\n")


class Api:
    def __init__(self, base_url):
        self.base = base_url.rstrip("/")
        self.token = None

    def call(self, method, path, body=None):
        url = f"{self.base}{path}"
        data = None
        headers = {"Accept": "application/json"}
        if body is not None:
            data = json.dumps(body, ensure_ascii=False).encode("utf-8")
            headers["Content-Type"] = "application/json"
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"

        req = urllib.request.Request(url, data=data, headers=headers, method=method)
        try:
            with urllib.request.urlopen(req, timeout=20) as resp:
                payload = json.loads(resp.read().decode("utf-8"))
        except urllib.error.HTTPError as e:
            raise RuntimeError(f"{method} {path} -> HTTP {e.code}: {e.read()[:300]!r}")
        except urllib.error.URLError as e:
            raise RuntimeError(f"{method} {path} 连接失败: {e.reason}")

        if payload.get("code") != 0:
            raise RuntimeError(f"{method} {path} 业务失败: {payload.get('message')}")
        return payload.get("data")

    def login(self, username, password):
        data = self.call("POST", "/api/v1/auth/login", {"username": username, "password": password})
        self.token = data["token"]
        return data["user"]


def ensure_category(api, name, slug, description, dry):
    data = api.call("GET", "/api/v1/categories")
    for c in data.get("list", []):
        if c["name"] == name:
            return c["id"]
    if dry:
        print(f"    [dry-run] 将创建分类 {name}")
        return 0
    created = api.call(
        "POST",
        "/api/v1/admin/categories",
        {"name": name, "slug": slug, "description": description, "sort": 10},
    )
    print(f"    + 创建分类 {name} (id={created.get('id')})")
    return created.get("id") or 0


# 中文标签名 -> 英文 slug。未收录的退化成 tag-<hash>，保证唯一且可读性尚可。
TAG_SLUGS = {
    "架构": "architecture",
    "后端": "backend",
    "前端": "frontend",
    "数据库": "database",
    "部署": "deploy",
    "认证": "auth",
    "评论": "comment",
    "踩坑": "pitfall",
    "踩坑记录": "pitfall",
    "性能": "performance",
    "安全": "security",
    "测试": "test",
    "复盘": "retro",
    "上传": "upload",
    "运维": "ops",
    "Gin": "gin",
    "GORM": "gorm",
    "Nginx": "nginx",
    "JWT": "jwt",
    "MySQL": "mysql",
    "Go": "go",
    "Vue": "vue",
}


def tag_slug(name):
    if name in TAG_SLUGS:
        return TAG_SLUGS[name]
    # 用 md5 而不是内置 hash：Python 的 hash 每个进程都会加盐，
    # 同一标签两次导入会得到不同 slug，进而创建出重复标签。
    return "tag-" + hashlib.md5(name.encode("utf-8")).hexdigest()[:6]


def ensure_tag(api, name, dry):
    data = api.call("GET", "/api/v1/tags")
    slug = tag_slug(name)
    for t in data.get("list", []):
        # 同名直接复用；slug 相同也复用（例如「踩坑」与「踩坑记录」共享 pitfall），
        # 否则 POST 会因为 slug 唯一索引冲突返回 409。
        if t["name"] == name or t["slug"] == slug:
            return t["id"]
    if dry:
        print(f"    [dry-run] 将创建标签 {name} (slug={slug})")
        return 0
    try:
        created = api.call("POST", "/api/v1/admin/tags", {"name": name, "slug": slug})
        print(f"    + 创建标签 {name} (id={created.get('id')})")
        return created.get("id") or 0
    except RuntimeError:
        # 409 兜底：可能是并发创建，也可能是历史数据占用了同名/slug
        data = api.call("GET", "/api/v1/tags")
        for t in data.get("list", []):
            if t["name"] == name or t["slug"] == slug:
                print(f"    = 复用已有标签 {t['name']} (id={t['id']})")
                return t["id"]
        raise


def existing_articles(api):
    data = api.call("GET", "/api/v1/admin/articles?page=1&page_size=100")
    return {a["title"]: a["id"] for a in data.get("list", [])}


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--base-url", default="http://127.0.0.1:8090")
    ap.add_argument("--username", default="admin")
    ap.add_argument("--password", required=True)
    ap.add_argument("--dir", default=str(ARTICLES_DIR))
    ap.add_argument("--dry-run", action="store_true")
    args = ap.parse_args()

    files = sorted(Path(args.dir).glob("*.md"))
    if not files:
        print(f"{args.dir} 下没有 .md 文件", file=sys.stderr)
        return 1

    api = Api(args.base_url)
    user = api.login(args.username, args.password)
    print(f"已登录：{user.get('nickname')} ({user.get('role')})\n")

    category_ids = {}
    tag_ids = {}
    existing = existing_articles(api)

    for f in files:
        meta, body = parse_frontmatter(f.read_text(encoding="utf-8"))
        title = meta.get("title") or f.stem
        cat_name = meta.get("category", "建站记")
        tags = [t.strip() for t in meta.get("tags", "").split(",") if t.strip()]

        print(f"==> {f.name}")
        print(f"    标题: {title}  分类: {cat_name}  标签: {tags or '无'}")

        if cat_name not in category_ids:
            category_ids[cat_name] = ensure_category(
                api, cat_name, "build-log", "这个博客是怎么一点点搭起来的", args.dry_run
            )
        for t in tags:
            if t not in tag_ids:
                tag_ids[t] = ensure_tag(api, t, args.dry_run)

        payload = {
            "title": title,
            "summary": meta.get("summary", ""),
            "content": body,
            "status": int(meta.get("status", "1")),
            "is_top": int(meta.get("is_top", "0")),
            "category_id": category_ids[cat_name],
            "tag_ids": [tag_ids[t] for t in tags if tag_ids.get(t)],
        }

        if args.dry_run:
            print(f"    [dry-run] 正文 {len(body)} 字符，未写入\n")
            continue

        if title in existing:
            api.call("PUT", f"/api/v1/admin/articles/{existing[title]}", payload)
            print(f"    ↻ 已更新 (id={existing[title]})\n")
        else:
            created = api.call("POST", "/api/v1/admin/articles", payload)
            print(f"    + 已创建 (id={created.get('id')})\n")

    print("完成")
    return 0


if __name__ == "__main__":
    sys.exit(main())
