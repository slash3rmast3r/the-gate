package applogs

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

type Entry struct {
	Time     string `json:"time"`
	Kind     string `json:"kind"` // access | error
	Host     string `json:"host,omitempty"`
	Client   string `json:"client,omitempty"`
	Method   string `json:"method,omitempty"`
	Path     string `json:"path,omitempty"`
	Status   int    `json:"status,omitempty"`
	Duration string `json:"duration,omitempty"`
	Router   string `json:"router,omitempty"`
	Message  string `json:"message,omitempty"`
	Raw      string `json:"raw,omitempty"`
}

type Reader struct {
	AccessPath string
	ErrorPath  string
}

func New(dir string) *Reader {
	dir = strings.TrimRight(dir, "/")
	return &Reader{
		AccessPath: dir + "/access.log",
		ErrorPath:  dir + "/traefik.log",
	}
}

func (r *Reader) ForHost(host, kind string, limit int, query string) ([]Entry, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	host = strings.ToLower(strings.TrimSpace(host))
	query = strings.ToLower(strings.TrimSpace(query))
	kind = strings.ToLower(kind)
	if kind == "" {
		kind = "access"
	}

	path := r.AccessPath
	if kind == "error" {
		path = r.ErrorPath
	}

	lines, err := tailLines(path, 8000)
	if err != nil {
		if os.IsNotExist(err) {
			return []Entry{}, nil
		}
		return nil, err
	}

	out := make([]Entry, 0, limit)
	for i := len(lines) - 1; i >= 0 && len(out) < limit; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		var e Entry
		if kind == "access" {
			e = parseAccess(line)
			if host != "" && !hostMatch(e.Host, host) {
				continue
			}
		} else {
			e = parseError(line, host)
			if host != "" && e.Host == "" && !strings.Contains(strings.ToLower(line), host) {
				continue
			}
		}
		if query != "" {
			blob := strings.ToLower(e.Raw + " " + e.Message + " " + e.Path + " " + e.Client + " " + e.Method)
			if !strings.Contains(blob, query) {
				continue
			}
		}
		e.Kind = kind
		out = append(out, e)
	}
	return out, nil
}

func hostMatch(entryHost, want string) bool {
	entryHost = strings.ToLower(strings.TrimSpace(entryHost))
	if entryHost == want {
		return true
	}
	// strip port if present
	if i := strings.Index(entryHost, ":"); i > 0 {
		entryHost = entryHost[:i]
	}
	return entryHost == want
}

func parseAccess(line string) Entry {
	e := Entry{Raw: line, Time: time.Now().UTC().Format(time.RFC3339)}
	if strings.HasPrefix(line, "{") {
		var m map[string]any
		if json.Unmarshal([]byte(line), &m) == nil {
			e.Time = str(m, "StartUTC", "time", "Timestamp")
			e.Host = str(m, "RequestHost", "request_Host")
			e.Client = str(m, "ClientHost", "ClientAddr", "RemoteAddr")
			e.Method = str(m, "RequestMethod")
			e.Path = str(m, "RequestPath", "RequestURI")
			e.Router = str(m, "RouterName")
			e.Status = intVal(m, "DownstreamStatus", "OriginStatus")
			if d, ok := m["Duration"]; ok {
				e.Duration = stringify(d)
			}
			if e.Time == "" {
				e.Time = time.Now().UTC().Format(time.RFC3339)
			}
			return e
		}
	}
	// CLF-ish fallback: keep raw, try to pull host from quoted request
	e.Message = line
	return e
}

func parseError(line, host string) Entry {
	e := Entry{Raw: line, Message: line, Kind: "error"}
	if host != "" && strings.Contains(strings.ToLower(line), strings.ToLower(host)) {
		e.Host = host
	}
	// Traefik often prefixes with time
	if len(line) > 20 {
		e.Time = strings.TrimSpace(line[:19])
	}
	return e
}

func str(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			s := stringify(v)
			if s != "" {
				return s
			}
		}
	}
	return ""
}

func intVal(m map[string]any, keys ...string) int {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case float64:
				return int(t)
			case int:
				return t
			case string:
				n, _ := strconv.Atoi(t)
				return n
			}
		}
	}
	return 0
}

func stringify(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case json.Number:
		return t.String()
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func tailLines(path string, maxLines int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	// read at most last 4 MiB
	const maxBytes = 4 << 20
	start := int64(0)
	if info.Size() > maxBytes {
		start = info.Size() - maxBytes
	}
	if _, err := f.Seek(start, 0); err != nil {
		return nil, err
	}
	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	var lines []string
	first := true
	for sc.Scan() {
		line := sc.Text()
		if first && start > 0 {
			first = false
			continue // likely partial line
		}
		lines = append(lines, line)
		if len(lines) > maxLines {
			lines = lines[len(lines)-maxLines:]
		}
	}
	return lines, sc.Err()
}