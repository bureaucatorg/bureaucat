// Package mailer sends email over SMTP using settings configured in /admin.
package mailer

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"bereaucat/internal/store"
)

// SettingKey is the settings table key holding the SMTP configuration.
const SettingKey = "smtp"

const (
	TLSModeStartTLS = "starttls"
	TLSModeImplicit = "tls"
)

const sendTimeout = 30 * time.Second

// Settings is the SMTP configuration stored under SettingKey.
type Settings struct {
	Enabled     bool   `json:"enabled"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	FromAddress string `json:"from_address"`
	FromName    string `json:"from_name"`
	TLSMode     string `json:"tls_mode"`
	AppURL      string `json:"app_url"`
}

// Message is a single email to one recipient.
type Message struct {
	To      string
	Subject string
	Text    string
	HTML    string
}

// Load reads the SMTP settings. A missing row returns zero (disabled) settings.
func Load(ctx context.Context, s store.Querier) (Settings, error) {
	var cfg Settings
	setting, err := s.GetSetting(ctx, SettingKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	err = json.Unmarshal(setting.Value, &cfg)
	return cfg, err
}

// Validate checks that the settings are complete enough to send mail.
func (cfg Settings) Validate() error {
	if cfg.Host == "" {
		return errors.New("host is required")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	if cfg.TLSMode != TLSModeStartTLS && cfg.TLSMode != TLSModeImplicit {
		return errors.New("tls_mode must be 'starttls' or 'tls'")
	}
	if cfg.Username != "" && cfg.Password == "" {
		return errors.New("password is required when username is set")
	}
	if _, err := mail.ParseAddress(cfg.FromAddress); err != nil {
		return errors.New("from_address must be a valid email address")
	}
	u, err := url.Parse(cfg.AppURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return errors.New("app_url must be an absolute http(s) URL")
	}
	return nil
}

// Send delivers msg using cfg. TLS (1.2+) is always required, either implicit or via STARTTLS.
func Send(ctx context.Context, cfg Settings, msg Message) error {
	body, err := buildMessage(cfg, msg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	tlsCfg := &tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12}
	dialer := &net.Dialer{}

	var conn net.Conn
	if cfg.TLSMode == TLSModeImplicit {
		conn, err = (&tls.Dialer{NetDialer: dialer, Config: tlsCfg}).DialContext(ctx, "tcp", addr)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("connect %s: %w", addr, err)
	}
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	c, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp handshake: %w", err)
	}
	defer c.Close()

	if cfg.TLSMode == TLSModeStartTLS {
		if ok, _ := c.Extension("STARTTLS"); !ok {
			return errors.New("server does not support STARTTLS")
		}
		if err := c.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	if cfg.Username != "" {
		if err := c.Auth(smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
	}

	if err := c.Mail(cfg.FromAddress); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := c.Rcpt(msg.To); err != nil {
		return fmt.Errorf("rcpt to: %w", err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	if _, err := w.Write(body); err != nil {
		return fmt.Errorf("write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close body: %w", err)
	}
	return c.Quit()
}

func buildMessage(cfg Settings, msg Message) ([]byte, error) {
	body, contentType, err := alternativeBody(msg)
	if err != nil {
		return nil, err
	}

	from := mail.Address{Name: cfg.FromName, Address: cfg.FromAddress}
	to := mail.Address{Address: msg.To}
	domain := cfg.FromAddress[strings.LastIndex(cfg.FromAddress, "@")+1:]
	subject := strings.Join(strings.Fields(msg.Subject), " ")

	headers := []string{
		"From: " + from.String(),
		"To: " + to.String(),
		"Subject: " + mime.QEncoding.Encode("utf-8", subject),
		"Date: " + time.Now().Format(time.RFC1123Z),
		fmt.Sprintf("Message-ID: <%s@%s>", uuid.NewString(), domain),
		"MIME-Version: 1.0",
		"Content-Type: " + contentType,
	}

	var buf bytes.Buffer
	buf.WriteString(strings.Join(headers, "\r\n") + "\r\n\r\n")
	buf.Write(body)
	return buf.Bytes(), nil
}

func alternativeBody(msg Message) ([]byte, string, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for _, part := range []struct{ contentType, content string }{
		{"text/plain; charset=utf-8", msg.Text},
		{"text/html; charset=utf-8", msg.HTML},
	} {
		pw, err := mw.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {part.contentType},
			"Content-Transfer-Encoding": {"quoted-printable"},
		})
		if err != nil {
			return nil, "", err
		}
		qp := quotedprintable.NewWriter(pw)
		if _, err := qp.Write([]byte(part.content)); err != nil {
			return nil, "", err
		}
		if err := qp.Close(); err != nil {
			return nil, "", err
		}
	}
	if err := mw.Close(); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), "multipart/alternative; boundary=" + mw.Boundary(), nil
}
