package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"

	"github.com/skip2/go-qrcode"
	"github.com/vpn-service/backend/src/utils"
)

// GenerateQRCode generates a QR code for a WireGuard configuration
func GenerateQRCode(config string) (string, error) {
	// Generate QR code
	qr, err := qrcode.New(config, qrcode.Medium)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code: %v", err)
	}

	// Set QR code options
	qr.BackgroundColor = 0xffffff
	qr.ForegroundColor = 0x000000

	// Create PNG image
	var buf bytes.Buffer
	if err := png.Encode(&buf, qr.Image(256)); err != nil {
		return "", fmt.Errorf("failed to encode QR code as PNG: %v", err)
	}

	// Encode as base64
	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	
	// Return data URL
	return fmt.Sprintf("data:image/png;base64,%s", base64Str), nil
}

// GenerateQRCodeForPeer generates a QR code for a peer configuration
func GenerateQRCodeForPeer(peerID, config string) (string, error) {
	// Generate QR code
	qrCode, err := GenerateQRCode(config)
	if err != nil {
		utils.LogError("Failed to generate QR code for peer %s: %v", peerID, err)
		return "", err
	}

	return qrCode, nil
}
