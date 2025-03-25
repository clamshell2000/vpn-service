package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/vpn-service/backend/src/config"
	"github.com/vpn-service/backend/src/utils"
)

// TemplateManager manages WireGuard configuration templates
type TemplateManager struct {
	config    *config.Config
	templates map[string]*template.Template
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(cfg *config.Config) (*TemplateManager, error) {
	tm := &TemplateManager{
		config:    cfg,
		templates: make(map[string]*template.Template),
	}

	// Load templates
	if err := tm.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %v", err)
	}

	return tm, nil
}

// loadTemplates loads all templates from the template directory
func (tm *TemplateManager) loadTemplates() error {
	// Get template directory
	templateDir := filepath.Join("vpn", "wireguard", "config_templates")
	
	// Read template directory
	files, err := ioutil.ReadDir(templateDir)
	if err != nil {
		return fmt.Errorf("failed to read template directory: %v", err)
	}

	// Load each template
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Get file name without extension
		name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		// Read template file
		templatePath := filepath.Join(templateDir, file.Name())
		templateData, err := ioutil.ReadFile(templatePath)
		if err != nil {
			return fmt.Errorf("failed to read template file %s: %v", templatePath, err)
		}

		// Parse template
		tmpl, err := template.New(name).Parse(string(templateData))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %v", name, err)
		}

		// Add template to map
		tm.templates[name] = tmpl
		utils.LogInfo("Loaded template: %s", name)
	}

	return nil
}

// GenerateConfig generates a configuration from a template
func (tm *TemplateManager) GenerateConfig(templateName string, data map[string]interface{}) (string, error) {
	// Get template
	tmpl, ok := tm.templates[templateName]
	if !ok {
		// Try to use generic template
		tmpl, ok = tm.templates["generic"]
		if !ok {
			return "", fmt.Errorf("template not found: %s", templateName)
		}
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}

	return buf.String(), nil
}

// SaveConfig saves a configuration to a file
func (tm *TemplateManager) SaveConfig(configName, config string) (string, error) {
	// Create config directory if it doesn't exist
	configDir := filepath.Join(tm.config.WireGuard.ConfigDir, "clients")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %v", err)
	}

	// Create config file
	configPath := filepath.Join(configDir, configName+".conf")
	if err := ioutil.WriteFile(configPath, []byte(config), 0644); err != nil {
		return "", fmt.Errorf("failed to write config file: %v", err)
	}

	return configPath, nil
}

// GetTemplateNames gets all template names
func (tm *TemplateManager) GetTemplateNames() []string {
	names := make([]string, 0, len(tm.templates))
	for name := range tm.templates {
		names = append(names, name)
	}
	return names
}

// GetDeviceTemplate gets the template name for a device type
func (tm *TemplateManager) GetDeviceTemplate(deviceType string) string {
	// Normalize device type
	deviceType = strings.ToLower(deviceType)

	// Check if template exists
	if _, ok := tm.templates[deviceType]; ok {
		return deviceType
	}

	// Return generic template
	return "generic"
}
