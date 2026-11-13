package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// 允许的图片类型：MIME -> 文件扩展名。
//
// 关键点：扩展名由「文件真实内容」决定，不由用户上传的文件名决定。
// 用户把木马命名为 logo.png 传上来，嗅探出来的 MIME 是 application/x-php 或
// application/octet-stream，直接不在白名单里，被拒。
// 文件名参数完全不可信——它可以叫 ../../../etc/passwd.png。
var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// UploadService 处理文件上传
type UploadService struct {
	dir   string // 存储根目录，例如 ./uploads
	maxMB int    // 单文件大小上限
}

func NewUploadService(dir string, maxMB int) *UploadService {
	if dir == "" {
		dir = "./uploads"
	}
	if maxMB <= 0 {
		maxMB = 5
	}
	return &UploadService{dir: dir, maxMB: maxMB}
}

// Dir 暴露存储目录给 router 挂静态文件服务
func (s *UploadService) Dir() string { return s.dir }

// UploadResult 上传成功后返回给前端的信息
type UploadResult struct {
	URL      string `json:"url"`       // 可直接放进 <img src>，也能存进文章 cover 字段
	FileName string `json:"file_name"` // 实际落盘的文件名
	Size     int64  `json:"size"`      // 字节数
	MIME     string `json:"mime"`      // 嗅探出的真实类型
}

// Save 保存上传的图片，返回访问路径
func (s *UploadService) Save(fh *multipart.FileHeader) (*UploadResult, error) {
	if fh.Size <= 0 {
		return nil, fmt.Errorf("%w: 文件为空", ErrInvalidInput)
	}
	if fh.Size > int64(s.maxMB)<<20 {
		return nil, fmt.Errorf("%w: 图片不能超过 %d MB", ErrInvalidInput, s.maxMB)
	}

	src, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer src.Close()

	// 读前 512 字节嗅探真实类型（http.DetectContentType 的约定长度）
	head := make([]byte, 512)
	n, err := io.ReadFull(src, head)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, fmt.Errorf("读取文件头失败: %w", err)
	}
	head = head[:n]

	mime := http.DetectContentType(head)
	ext, ok := allowedImageTypes[mime]
	if !ok {
		return nil, fmt.Errorf("%w: 只允许 jpg/png/gif/webp 图片，识别到的类型是 %s", ErrInvalidInput, mime)
	}

	// 嗅探用的前 512 字节要从头再读一遍，否则存下来的文件会少一截
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("回退文件读取位置失败: %w", err)
	}

	// 按年月分目录：uploads/2026/11/xxxx.png
	// 单目录文件过多时，ls、备份、rsync 都会变慢
	subDir := time.Now().Format("2006/01")
	dstDir := filepath.Join(s.dir, subDir)
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return nil, fmt.Errorf("创建上传目录失败: %w", err)
	}

	// 文件名完全由服务端随机生成：不用用户文件名，就彻底没有路径穿越问题
	name := randomFileName() + ext
	dstPath := filepath.Join(dstDir, name)

	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		_ = os.Remove(dstPath) // 写失败就把半截文件删掉，别留垃圾
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	return &UploadResult{
		// URL 用正斜杠拼，Windows 上 path/filepath 会给出反斜杠，浏览器认不了
		URL:      "/uploads/" + subDir + "/" + name,
		FileName: name,
		Size:     fh.Size,
		MIME:     mime,
	}, nil
}

// randomFileName 生成 16 位十六进制随机名。
// 用 crypto/rand 而不是 math/rand：math/rand 的种子可预测，可能被猜出文件名。
func randomFileName() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		// 极端情况兜底：用纳秒时间戳，保证不 panic
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}
