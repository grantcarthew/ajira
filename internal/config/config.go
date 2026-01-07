package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"
)

const (
	DefaultTimeout = 30 * time.Second
)

type Config struct {
	BaseURL     string
	Email       string
	APIToken    string
	Project     string
	HTTPTimeout time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		HTTPTimeout: DefaultTimeout,
	}

	var errs []error

	cfg.BaseURL = os.Getenv("JIRA_BASE_URL")
	if cfg.BaseURL == "" {
		errs = append(errs, errors.New("missing required environment variable: JIRA_BASE_URL"))
	} else {
		if err := validateBaseURL(cfg.BaseURL); err != nil {
			errs = append(errs, err)
		}
	}

	cfg.Email = os.Getenv("JIRA_EMAIL")
	if cfg.Email == "" {
		errs = append(errs, errors.New("missing required environment variable: JIRA_EMAIL"))
	}

	cfg.APIToken = os.Getenv("JIRA_API_TOKEN")
	if cfg.APIToken == "" {
		cfg.APIToken = os.Getenv("ATLASSIAN_API_TOKEN")
	}
	if cfg.APIToken == "" {
		errs = append(errs, errors.New("missing required environment variable: JIRA_API_TOKEN (or ATLASSIAN_API_TOKEN)"))
	}

	cfg.Project = os.Getenv("JIRA_PROJECT")

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return cfg, nil
}

func validateBaseURL(baseURL string) error {
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid JIRA_BASE_URL: %w", err)
	}
	if u.Scheme != "https" {
		return errors.New("invalid JIRA_BASE_URL: must use https scheme")
	}
	if u.Host == "" {
		return errors.New("invalid JIRA_BASE_URL: missing host")
	}
	return nil
}
