package metrics

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

type Counter struct {
	name   string
	labels string
	val    atomic.Int64
}

func (c *Counter) Inc() { c.val.Add(1) }

func (c *Counter) Add(delta int64) { c.val.Add(delta) }

func (c *Counter) Value() int64 { return c.val.Load() }

type Registry struct {
	mu       sync.RWMutex
	counters []*Counter
}

var Global = &Registry{}

func (r *Registry) Counter(name string, labelPairs ...string) *Counter {
	labels := renderLabels(labelPairs)
	key := name + labels

	r.mu.RLock()
	for _, c := range r.counters {
		if c.name+c.labels == key {
			r.mu.RUnlock()
			return c
		}
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, c := range r.counters {
		if c.name+c.labels == key {
			return c
		}
	}

	c := &Counter{name: name, labels: labels}
	r.counters = append(r.counters, c)

	return c
}

func (r *Registry) WritePrometheusText(w io.Writer) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	byName := map[string][]*Counter{}
	for _, c := range r.counters {
		byName[c.name] = append(byName[c.name], c)
	}

	names := make([]string, 0, len(byName))

	for n := range byName {
		names = append(names, n)
	}

	sort.Strings(names)

	for _, name := range names {
		_, _ = fmt.Fprintf(w, "# HELP %s A/B platform counter\n", name)
		_, _ = fmt.Fprintf(w, "# TYPE %s counter\n", name)

		for _, c := range byName[name] {
			_, _ = fmt.Fprintf(w, "%s%s %d\n", c.name, c.labels, c.val.Load())
		}
	}
}

func (r *Registry) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	r.WritePrometheusText(w)
}

// HTTPRequestsTotal returns a counter for ab_http_requests_total{method, path, status}.
func HTTPRequestsTotal(method, path, status string) *Counter {
	return Global.Counter("ab_http_requests_total", "method", method, "path", path, "status", status)
}

func renderLabels(pairs []string) string {
	if len(pairs) == 0 {
		return ""
	}

	var sb strings.Builder

	sb.WriteByte('{')

	for i := 0; i+1 < len(pairs); i += 2 {
		if i > 0 {
			sb.WriteByte(',')
		}

		fmt.Fprintf(&sb, `%s="%s"`, pairs[i], pairs[i+1])
	}

	sb.WriteByte('}')

	return sb.String()
}

