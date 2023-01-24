package object

import (
	"opensvc.com/opensvc/core/keywords"
	"opensvc.com/opensvc/core/path"
	"opensvc.com/opensvc/util/funcopt"
	"opensvc.com/opensvc/util/key"
)

type (
	ccfg struct {
		core
	}

	//
	// Ccfg is the clusterwide configuration store.
	//
	// The content is the same as node.conf, and is overriden by
	// the definition found in node.conf.
	//
	Ccfg interface {
		Core
	}
)

var ccfgPrivateKeywords = []keywords.Keyword{
	{
		Section:     "DEFAULT",
		Option:      "id",
		DefaultText: "<random uuid>",
		Scopable:    false,
		Text:        "A RFC 4122 random uuid generated by the agent. To use as reference in resources definitions instead of the service name, so the service can be renamed without affecting the resources.",
	},
}

var ccfgKeywordStore = keywords.Store(append(ccfgPrivateKeywords, nodeCommonKeywords...))

func NewCluster(opts ...funcopt.O) (*ccfg, error) {
	return newCcfg(path.Cluster, opts...)
}

// newCcfg allocates a ccfg kind object.
func newCcfg(p any, opts ...funcopt.O) (*ccfg, error) {
	s := &ccfg{}
	err := s.init(s, p, opts...)
	return s, err
}

func (t ccfg) KeywordLookup(k key.T, sectionType string) keywords.Keyword {
	return keywordLookup(ccfgKeywordStore, k, t.path.Kind, sectionType)
}

func (t ccfg) Name() string {
	k := key.New("cluster", "name")
	return t.config.GetString(k)
}
