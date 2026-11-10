// Package branding loads the instance's custom name, logo and favicon.
package branding

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"image"
	_ "image/png"
	"net/http"

	"bereaucat/internal/store"
)

const DefaultName = "Bureaucat"

type Asset string

const (
	AssetLogo    Asset = "logo"
	AssetFavicon Asset = "favicon"
)

const (
	MaxLogoBytes    = 256 << 10
	MaxFaviconBytes = 64 << 10
)

var ErrInvalidImage = errors.New("invalid image")

// Settings is the stored "branding" row.
type Settings struct {
	Enabled bool   `json:"enabled"`
	AppName string `json:"app_name"`
}

// Image is a stored logo or favicon. A removed image is stored as {}.
type Image struct {
	ContentType string `json:"content_type,omitempty"`
	Data        []byte `json:"data,omitempty"`
	Hash        string `json:"hash,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
}

// Brand is the effective branding: images are only set while custom branding is enabled.
type Brand struct {
	Custom  bool
	Name    string
	Logo    *Image
	Favicon *Image
}

func settingKey(a Asset) string { return "branding_" + string(a) }

func (a Asset) Valid() bool { return a == AssetLogo || a == AssetFavicon }

func (a Asset) MaxBytes() int {
	if a == AssetLogo {
		return MaxLogoBytes
	}
	return MaxFaviconBytes
}

func LoadSettings(ctx context.Context, q store.Querier) Settings {
	s := Settings{AppName: DefaultName}
	if row, err := q.GetSetting(ctx, "branding"); err == nil {
		_ = json.Unmarshal(row.Value, &s)
	}
	return s
}

func LoadImage(ctx context.Context, q store.Querier, a Asset) *Image {
	row, err := q.GetSetting(ctx, settingKey(a))
	if err != nil {
		return nil
	}
	var img Image
	if json.Unmarshal(row.Value, &img) != nil || len(img.Data) == 0 {
		return nil
	}
	return &img
}

func Load(ctx context.Context, q store.Querier) Brand {
	s := LoadSettings(ctx, q)
	if !s.Enabled {
		return Brand{Name: DefaultName}
	}
	b := Brand{Custom: true, Name: s.AppName}
	if b.Name == "" {
		b.Name = DefaultName
	}
	b.Logo = LoadImage(ctx, q, AssetLogo)
	b.Favicon = LoadImage(ctx, q, AssetFavicon)
	return b
}

// NewImage validates an upload: logos must be PNG, favicons PNG or ICO.
func NewImage(a Asset, data []byte) (Image, error) {
	if len(data) == 0 || len(data) > a.MaxBytes() {
		return Image{}, ErrInvalidImage
	}
	img := Image{ContentType: http.DetectContentType(data), Data: data}
	switch {
	case img.ContentType == "image/png":
		cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil || cfg.Width == 0 || cfg.Height == 0 {
			return Image{}, ErrInvalidImage
		}
		img.Width, img.Height = cfg.Width, cfg.Height
	case img.ContentType == "image/x-icon" && a == AssetFavicon:
	default:
		return Image{}, ErrInvalidImage
	}
	sum := sha256.Sum256(data)
	img.Hash = hex.EncodeToString(sum[:8])
	return img, nil
}

func SaveImage(ctx context.Context, q store.Querier, a Asset, img Image) error {
	value, err := json.Marshal(img)
	if err != nil {
		return err
	}
	_, err = q.UpsertSetting(ctx, store.UpsertSettingParams{Key: settingKey(a), Value: value})
	return err
}

func RemoveImage(ctx context.Context, q store.Querier, a Asset) error {
	return SaveImage(ctx, q, a, Image{})
}

// URL is the public, cache-busted URL of a stored image, or "" when absent.
func URL(a Asset, img *Image) string {
	if img == nil {
		return ""
	}
	return "/api/v1/settings/branding/" + string(a) + "?v=" + img.Hash
}

// FaviconURL falls back to the logo; "" means use the default favicon.
func (b Brand) FaviconURL() (url, contentType string) {
	if b.Favicon != nil {
		return URL(AssetFavicon, b.Favicon), b.Favicon.ContentType
	}
	if b.Logo != nil {
		return URL(AssetLogo, b.Logo), b.Logo.ContentType
	}
	return "", ""
}
