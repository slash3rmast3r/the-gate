package protections

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const defaultCRS = "v4.14.0"

func UpdateCRS(corazaDir, version string) (string, error) {
	if version == "" {
		version = defaultCRS
	}
	if err := os.MkdirAll(corazaDir, 0o755); err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://github.com/coreruleset/coreruleset/archive/refs/tags/%s.tar.gz", version)
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("download CRS: HTTP %d", resp.StatusCode)
	}

	tmp, err := os.MkdirTemp("", "gate-crs-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)

	if err := untarGz(resp.Body, tmp); err != nil {
		return "", err
	}
	entries, err := os.ReadDir(tmp)
	if err != nil {
		return "", err
	}
	var src string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "coreruleset-") {
			src = filepath.Join(tmp, e.Name())
			break
		}
	}
	if src == "" {
		return "", fmt.Errorf("crs archive layout unexpected")
	}

	rulesDest := filepath.Join(corazaDir, "rules")
	_ = os.RemoveAll(rulesDest)
	if err := os.MkdirAll(rulesDest, 0o755); err != nil {
		return "", err
	}
	rulesSrc := filepath.Join(src, "rules")
	copied := 0
	err = filepath.Walk(rulesSrc, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if !strings.HasSuffix(info.Name(), ".conf") {
			return nil
		}
		rel, _ := filepath.Rel(rulesSrc, path)
		dst := filepath.Join(rulesDest, rel)
		_ = os.MkdirAll(filepath.Dir(dst), 0o755)
		if err := copyFile(path, dst); err != nil {
			return err
		}
		copied++
		return nil
	})
	if err != nil {
		return "", err
	}
	example := filepath.Join(src, "crs-setup.conf.example")
	if _, err := os.Stat(example); err == nil {
		_ = copyFile(example, filepath.Join(corazaDir, "crs-setup.conf"))
	}
	return fmt.Sprintf("OWASP CRS %s installed (%d rule files)", version, copied), nil
}

func UpdateCrowdSecHub(container string) (string, error) {
	if container == "" {
		container = "gate-crowdsec"
	}
	if _, err := os.Stat("/var/run/docker.sock"); err != nil {
		return "", fmt.Errorf("docker.sock not mounted — cannot update CrowdSec hub from API")
	}
	out1, err := exec.Command("docker", "exec", container, "cscli", "hub", "update").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("hub update: %v (%s)", err, strings.TrimSpace(string(out1)))
	}
	out2, err := exec.Command("docker", "exec", container, "cscli", "hub", "upgrade").CombinedOutput()
	msg := strings.TrimSpace(string(out1) + "\n" + string(out2))
	if err != nil {
		return msg, fmt.Errorf("hub upgrade: %v (%s)", err, msg)
	}
	return "CrowdSec hub updated/upgraded:\n" + msg, nil
}

func untarGz(r io.Reader, dest string) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		target := filepath.Join(dest, hdr.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)+string(os.PathSeparator)) &&
			filepath.Clean(target) != filepath.Clean(dest) {
			return fmt.Errorf("invalid tar path")
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}