package httpt

import (
	"github.com/gin-gonic/gin/render"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

var htmlContentTypes = []string{"text/html charset=utf-8"}

type htmlInfo struct {
	Name       string
	Path       string
	Content    []byte
	ModifyTime int64
}

type HtmlCache struct {
	Root  string
	mutex sync.RWMutex
	cache map[string]*htmlInfo
}

type htmlRenderer struct {
	Name  string
	Data  interface{}
	cache *HtmlCache
}

func (r *htmlRenderer) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = htmlContentTypes
	}
}

func (o *HtmlCache) Instance(name string, data interface{}) render.Render {
	return &htmlRenderer{
		Name:  name,
		Data:  data,
		cache: o,
	}
}

func (o *HtmlCache) loop(refreshInterval int) {
	for {
		time.Sleep(time.Duration(refreshInterval) * time.Second)
		o.refresh()
	}
}

func (o *HtmlCache) refresh() {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	for _, one := range o.cache {
		stat, err := os.Stat(one.Path)
		if err == nil {
			var newmod = stat.ModTime().Unix()
			if newmod > one.ModifyTime {
				content, err := ioutil.ReadFile(one.Path)
				if err == nil {
					one.Content = content
					one.ModifyTime = newmod
				}
			}
		}
	}

}

func (r *htmlRenderer) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	var err error
	var content []byte

	r.cache.mutex.RLock()
	var info = r.cache.cache[r.Name]
	r.cache.mutex.RUnlock()

	if info == nil {
		var filename = r.Name
		var filepath = r.cache.Root + filename
		content, err = ioutil.ReadFile(filepath)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			return err
		}
		var stat, _ = os.Stat(filepath)
		info = new(htmlInfo)
		info.Name = r.Name
		info.Path = filepath
		info.Content = content
		info.ModifyTime = stat.ModTime().Unix()
		r.cache.mutex.Lock()
		r.cache.cache[r.Name] = info
		r.cache.mutex.Unlock()
	} else {
		content = info.Content
	}
	_, err = w.Write(content)
	return err
}
