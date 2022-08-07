package gocache

import (
	"net/http"
	"strings"
)

type HttpTool struct {
	self     string
	basePath string
}

const (
	defaultbasePath = "/gocache/"
)

func NewHttpTool(self string) *HttpTool {
	return &HttpTool{
		self:     self,
		basePath: defaultbasePath,
	}
}

func (ht *HttpTool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, ht.basePath) {
		panic("serving unexpected path" + r.URL.Path)
	}
	parts := strings.SplitN(r.URL.Path[len(ht.basePath):], "/", 2)
	if len(parts) < 2 {
		http.Error(w, "badrequest", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]
	var g *Group
	if g = GetGroup(groupName); g == nil {
		http.Error(w, "badrequest invalid group", http.StatusBadRequest)
		return
	}
	v, err := g.Get(key)
	if err != nil {
		http.Error(w, "badrequest invalid key", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(v.ByteSlice())
}
