package gocache

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	pb "gocache/cachepb"
	"gocache/consistenthash"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type HttpPool struct {
	self        string
	basePath    string
	mu          sync.Mutex // guards peers and httpGetters
	peers       *consistenthash.Map
	httpGetters map[string]*httpGetter
}

const (
	defaultbasePath = "/gocache/"
	defaultReplicas = 50
)

func NewHttpTool(self string) *HttpPool {
	return &HttpPool{
		self:     self,
		basePath: defaultbasePath,
	}
}

// Log info with server name
func (p *HttpPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handle all http requests
func (p *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("serving unexpected path" + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
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
	// Write the value to the response body as proto message
	body, err := proto.Marshal(&pb.Response{
		Value: v.ByteSlice(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")

	if _, err = w.Write(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Set updates the pool's list of peers
func (p *HttpPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer picks a peer according to key
func (p *HttpPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HttpPool)(nil)

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(in.GetGroup()), url.QueryEscape(in.GetKey()))
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}

	return nil
}

var _ PeerGetter = (*httpGetter)(nil)