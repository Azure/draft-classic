package draft

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/strvals"

	"github.com/Azure/draft/pkg/rpc"
)

type AppContext struct {
	srv  *Server
	req  *rpc.UpRequest
	buf  *bytes.Buffer
	tag  string
	img  string
	out  io.Writer
	vals chartutil.Values
}

// newAppContext prepares state carried across the various draft stage boundaries.
func newAppContext(s *Server, req *rpc.UpRequest, out io.Writer) (*AppContext, error) {
	raw := bytes.NewBuffer(req.AppArchive.Content)
	// write build context to a buffer so we can also write to the sha1 hash.
	b := new(bytes.Buffer)
	h := sha1.New()
	w := io.MultiWriter(b, h)
	if _, err := io.Copy(w, raw); err != nil {
		return nil, err
	}
	// truncate checksum to the first 40 characters (20 bytes) this is the
	// equivalent of `shasum build.tar.gz | awk '{print $1}'`.
	imgtag := fmt.Sprintf("%.20x", h.Sum(nil))
	prefix := s.cfg.Registry.Org
	if prefix != "" {
		prefix += "/"
	}
	image := fmt.Sprintf("%s/%s%s:%s", s.cfg.Registry.URL, prefix, req.AppName, imgtag)

	// inject certain values into the chart such as the registry location,
	// the application name, and the application version.
	tplstr := "image.name=%s,image.org=%s,image.registry=%s,image.tag=%s,basedomain=%s,ondraft=true"
	inject := fmt.Sprintf(tplstr, req.AppName, s.cfg.Registry.Org, s.cfg.Registry.URL, imgtag, s.cfg.Basedomain)

	vals, err := chartutil.ReadValues([]byte(req.Values.Raw))
	if err != nil {
		return nil, err
	}
	if err := strvals.ParseInto(inject, vals.AsMap()); err != nil {
		return nil, err
	}
	return &AppContext{
		srv:  s,
		req:  req,
		buf:  b,
		tag:  imgtag,
		img:  image,
		out:  out,
		vals: vals,
	}, nil
}
