package services

import (
	"encoding/json"
	"fmt"
)

// EncodeQRCodeData encodes QR code data as a JSON string
func EncodeQRCodeData(data *QRCodeData) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal QR code data: %w", err)
	}
	return string(jsonData), nil
}

// DecodeQRCodeData decodes a JSON string into QR code data
func DecodeQRCodeData(jsonData string) (*QRCodeData, error) {
	var data QRCodeData
	err := json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal QR code data: %w", err)
	}
	return &data, nil
}