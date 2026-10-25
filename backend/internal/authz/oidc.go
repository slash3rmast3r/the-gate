package authz

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/overthecloud/the-gate/internal/config"
)

type OIDC struct {
	cfg      config.Config
	mu       sync.Mutex
	states   map[string]time.Time
	authURL  string
	tokenURL string
	jwksURL  string
	userURL  string
}

type oidcDiscovery struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	JwksURI               string `json:"jwks_uri"`
}

func NewOIDC(cfg config.Config) *OIDC {
	return &OIDC{cfg: cfg, states: map[string]time.Time{}}
}

func (o *OIDC) Enabled() bool { return o.cfg.OIDCEnabled() }

func (o *OIDC) Discover(ctx context.Context) error {
	issuer := strings.TrimRight(o.cfg.OIDCIssuer, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, issuer+"/.well-known/openid-configuration", nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("oidc discovery %d: %s", resp.StatusCode, string(b))
	}
	var d oidcDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return err
	}
	o.authURL, o.tokenURL, o.userURL, o.jwksURL = d.AuthorizationEndpoint, d.TokenEndpoint, d.UserinfoEndpoint, d.JwksURI
	if o.authURL == "" || o.tokenURL == "" {
		return fmt.Errorf("oidc discovery incomplete")
	}
	return nil
}

func (o *OIDC) LoginURL() (string, error) {
	if o.authURL == "" {
		if err := o.Discover(context.Background()); err != nil {
			return "", err
		}
	}
	state := randomState()
	o.mu.Lock()
	o.states[state] = time.Now().Add(10 * time.Minute)
	o.mu.Unlock()

	q := url.Values{}
	q.Set("client_id", o.cfg.OIDCClientID)
	q.Set("response_type", "code")
	q.Set("scope", "openid profile email")
	q.Set("redirect_uri", o.cfg.OIDCRedirectURL)
	q.Set("state", state)
	return o.authURL + "?" + q.Encode(), nil
}

func (o *OIDC) ConsumeState(state string) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	exp, ok := o.states[state]
	delete(o.states, state)
	return ok && time.Now().Before(exp)
}

func (o *OIDC) Exchange(ctx context.Context, code string) (username string, err error) {
	if o.tokenURL == "" {
		if err := o.Discover(ctx); err != nil {
			return "", err
		}
	}
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", o.cfg.OIDCRedirectURL)
	form.Set("client_id", o.cfg.OIDCClientID)
	form.Set("client_secret", o.cfg.OIDCClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("token exchange %d: %s", resp.StatusCode, string(body))
	}
	var tok struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}
	if err := json.Unmarshal(body, &tok); err != nil {
		return "", err
	}
	if o.userURL != "" && tok.AccessToken != "" {
		ureq, _ := http.NewRequestWithContext(ctx, http.MethodGet, o.userURL, nil)
		ureq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
		uresp, err := http.DefaultClient.Do(ureq)
		if err == nil {
			defer uresp.Body.Close()
			var u struct {
				PreferredUsername string `json:"preferred_username"`
				Email             string `json:"email"`
				Name              string `json:"name"`
				Sub               string `json:"sub"`
			}
			_ = json.NewDecoder(uresp.Body).Decode(&u)
			if u.PreferredUsername != "" {
				return u.PreferredUsername, nil
			}
			if u.Email != "" {
				return u.Email, nil
			}
			if u.Name != "" {
				return u.Name, nil
			}
			if u.Sub != "" {
				return u.Sub, nil
			}
		}
	}
	return "oidc-user", nil
}

func randomState() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}