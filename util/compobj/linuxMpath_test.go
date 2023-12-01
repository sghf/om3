package main

import (
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestMpathAdd(t *testing.T) {
	testCases := map[string]struct {
		jsonRules     string
		expectError   bool
		expectedRules []interface{}
	}{
		"with a full rule": {
			jsonRules:   `[{"key":"lala", "op":"=", "value" : "ok"}]`,
			expectError: false,
			expectedRules: []interface{}{CompMpath{
				Key:   "lala",
				Op:    "=",
				Value: "ok",
			}},
		},

		"with missing key": {
			jsonRules:     `[{"op":"=", "value" : "ok"}]`,
			expectError:   true,
			expectedRules: nil,
		},

		"with missing op": {
			jsonRules:     `[{"key":"lala", "value" : "ok"}]`,
			expectError:   true,
			expectedRules: nil,
		},

		"with missing value": {
			jsonRules:     `[{"op":"=", "key" : "ok"}]`,
			expectError:   true,
			expectedRules: nil,
		},

		"with wrong op": {
			jsonRules:     `[{"key":"lala", "op":">>>", "value" : "ok"}]`,
			expectError:   true,
			expectedRules: nil,
		},

		"when value is a bool": {
			jsonRules:     `[{"key":"lala", "op":"=", "value" : true}]`,
			expectError:   true,
			expectedRules: nil,
		},

		"with string value and op >=": {
			jsonRules:     `[{"key":"lala", "op":">=", "value" : "true"}]`,
			expectError:   true,
			expectedRules: nil,
		},

		"with a full rule but device does not have precision of the vendor and product": {
			jsonRules:     `[{"key":"lala.device", "op":"=", "value" : "ok"}]`,
			expectError:   true,
			expectedRules: []interface{}{},
		},

		"with a full rule but device does not have precision of the product": {
			jsonRules:     `[{"key":"lala.device.{vendor}", "op":"=", "value" : "ok"}]`,
			expectError:   true,
			expectedRules: []interface{}{},
		},

		"with a full rule and device in key": {
			jsonRules:   `[{"key":"lala.device.{vendor}.{product}", "op":"=", "value" : "ok"}]`,
			expectError: false,
			expectedRules: []interface{}{CompMpath{
				Key:   "lala.device.{vendor}.{product}",
				Op:    "=",
				Value: "ok",
			}},
		},

		"with a full rule but multipath does not have precision of the wwid": {
			jsonRules:     `[{"key":"lala.multipath", "op":"=", "value" : "ok"}]`,
			expectError:   true,
			expectedRules: []interface{}{},
		},

		"with a full rule and multipath in key": {
			jsonRules:   `[{"key":"lala.multipath.{wwid}", "op":"=", "value" : "ok"}]`,
			expectError: false,
			expectedRules: []interface{}{CompMpath{
				Key:   "lala.multipath.{wwid}",
				Op:    "=",
				Value: "ok",
			}},
		},
	}

	for name, c := range testCases {
		t.Run(name, func(t *testing.T) {
			obj := CompMpaths{Obj: &Obj{rules: make([]interface{}, 0), verbose: true}}
			if c.expectError {
				require.Error(t, obj.Add(c.jsonRules))
			} else {
				require.NoError(t, obj.Add(c.jsonRules))
				require.Equal(t, c.expectedRules, obj.rules)
			}
		})
	}
}

func TestLoadMpathData(t *testing.T) {
	oriOsReadFile := osReadFile
	defer func() { osReadFile = oriOsReadFile }()

	testCases := map[string]struct {
		filePath     string
		expectedData MpathConf
	}{
		"with only a default section": {
			filePath: "./testdata/linuxMpath_conf_default",
			expectedData: MpathConf{
				BlackList: MpathBlackList{
					Name:     "blacklist",
					Wwids:    []string{},
					Devnodes: []string{},
					Devices:  []MpathSection{},
				},
				BlackListExceptions: MpathBlackList{
					Name:     "blacklist_exceptions",
					Wwids:    []string{},
					Devnodes: []string{},
					Devices:  []MpathSection{},
				},
				Defaults: MpathSection{
					Name:   "default",
					Indent: 1,
					Attr:   map[string][]string{"user_friendly_names": {"yes"}, "path_grouping_policy": {"multibus"}},
				},
				Devices:    []MpathSection{},
				Multipaths: []MpathSection{},
				Overrides: MpathSection{
					Name:   "overrides",
					Indent: 1,
					Attr:   map[string][]string{},
				},
			},
		},

		"with only blacklist section": {
			filePath: "./testdata/linuxMpath_conf_blacklist",
			expectedData: MpathConf{
				BlackList: MpathBlackList{
					Name:     "blacklist",
					Wwids:    []string{"*", `laal`},
					Devnodes: []string{`^hd[a-z]`},
					Devices: []MpathSection{{
						Name:   "device",
						Indent: 2,
						Attr:   map[string][]string{"vendor": {"IBM"}, "product": {"3S42"}},
					}, {
						Name:   "device",
						Indent: 2,
						Attr:   map[string][]string{"vendor": {"HP"}, "product": {"*"}},
					}},
				},
				BlackListExceptions: MpathBlackList{
					Name:     "blacklist_exceptions",
					Wwids:    []string{},
					Devnodes: []string{},
					Devices:  []MpathSection{},
				},
				Defaults: MpathSection{
					Name:   "default",
					Indent: 1,
					Attr:   map[string][]string{},
				},
				Devices:    []MpathSection{},
				Multipaths: []MpathSection{},
				Overrides: MpathSection{
					Name:   "overrides",
					Indent: 1,
					Attr:   map[string][]string{},
				},
			},
		},

		"with only blacklist_exceptions section": {
			filePath: "./testdata/linuxMpath_conf_blacklist_exceptions",
			expectedData: MpathConf{
				BlackList: MpathBlackList{
					Name:     "blacklist",
					Wwids:    []string{},
					Devnodes: []string{},
					Devices:  []MpathSection{},
				},
				BlackListExceptions: MpathBlackList{
					Name:     "blacklist_exceptions",
					Wwids:    []string{"*", `laal`},
					Devnodes: []string{`^hd[a-z]`},
					Devices: []MpathSection{{
						Name:   "device",
						Indent: 2,
						Attr:   map[string][]string{"vendor": {"IBM"}, "product": {"3S42"}},
					}, {
						Name:   "device",
						Indent: 2,
						Attr:   map[string][]string{"vendor": {"HP"}, "product": {"*"}},
					}},
				},
				Defaults: MpathSection{
					Name:   "default",
					Indent: 1,
					Attr:   map[string][]string{},
				},
				Devices:    []MpathSection{},
				Multipaths: []MpathSection{},
				Overrides: MpathSection{
					Name:   "overrides",
					Indent: 1,
					Attr:   map[string][]string{},
				},
			},
		},

		"with only devices section": {
			filePath: "./testdata/linuxMpath_conf_devices",
			expectedData: MpathConf{
				BlackList: MpathBlackList{
					Name:     "blacklist",
					Wwids:    []string{},
					Devnodes: []string{},
					Devices:  []MpathSection{},
				},
				BlackListExceptions: MpathBlackList{
					Name:     "blacklist_exceptions",
					Wwids:    []string{},
					Devnodes: []string{},
					Devices:  []MpathSection{},
				},
				Defaults: MpathSection{
					Name:   "default",
					Indent: 1,
					Attr:   map[string][]string{},
				},
				Devices: []MpathSection{
					{
						Name:   "device",
						Indent: 2,
						Attr:   map[string][]string{"vendor": {"IBM"}, "product": {"3S42"}},
					}, {
						Name:   "device",
						Indent: 2,
						Attr:   map[string][]string{"vendor": {"HP"}, "product": {"*"}},
					},
				},
				Multipaths: []MpathSection{},
				Overrides: MpathSection{
					Name:   "overrides",
					Indent: 1,
					Attr:   map[string][]string{}},
			},
		},

		"with only multipaths section": {
			filePath: "./testdata/linuxMpath_conf_multipaths",
			expectedData: MpathConf{
				BlackList: MpathBlackList{
					Name:     "blacklist",
					Wwids:    []string{},
					Devnodes: []string{},
					Devices:  []MpathSection{},
				},
				BlackListExceptions: MpathBlackList{
					Name:     "blacklist_exceptions",
					Wwids:    []string{},
					Devnodes: []string{},
					Devices:  []MpathSection{},
				},
				Defaults: MpathSection{
					Name:   "default",
					Indent: 1,
					Attr:   map[string][]string{},
				},
				Devices: []MpathSection{},
				Multipaths: []MpathSection{{
					Name:   "multipath",
					Indent: 2,
					Attr:   map[string][]string{"wwid": {"3600508b4000156d70001200000b0000"}},
				},
					{
						Name:   "multipath",
						Indent: 2,
						Attr:   map[string][]string{"wwid": {"1DEC_____321816758474"}, "alias": {"red"}, "rr_weight": {"priorities"}},
					},
				},
				Overrides: MpathSection{
					Name:   "overrides",
					Indent: 1,
					Attr:   map[string][]string{},
				},
			},
		},

		"with only a default override": {
			filePath: "./testdata/linuxMpath_conf_overrides",
			expectedData: MpathConf{
				BlackList: MpathBlackList{
					Name:     "blacklist",
					Wwids:    []string{},
					Devnodes: []string{},
					Devices:  []MpathSection{},
				},
				BlackListExceptions: MpathBlackList{
					Name:     "blacklist_exceptions",
					Wwids:    []string{},
					Devnodes: []string{},
					Devices:  []MpathSection{},
				},
				Defaults: MpathSection{
					Name:   "default",
					Indent: 1,
					Attr:   map[string][]string{},
				},
				Devices:    []MpathSection{},
				Multipaths: []MpathSection{},
				Overrides: MpathSection{
					Name:   "overrides",
					Indent: 1,
					Attr:   map[string][]string{"user_friendly_names": {"yes"}, "path_grouping_policy": {"multibus"}},
				},
			},
		},

		"with a full multipath file": {
			filePath: "./testdata/linuxMpath_conf_golden",
			expectedData: MpathConf{
				BlackList: MpathBlackList{
					Name:     "blacklist",
					Wwids:    []string{"*", `laal`},
					Devnodes: []string{`^hd[a-z]`},
					Devices: []MpathSection{{
						Name:   "device",
						Indent: 2,
						Attr:   map[string][]string{"vendor": {"IBM"}, "product": {"3S42"}},
					}, {
						Name:   "device",
						Indent: 2,
						Attr:   map[string][]string{"vendor": {"HP"}, "product": {"*"}},
					}},
				},
				BlackListExceptions: MpathBlackList{
					Name:     "blacklist_exceptions",
					Wwids:    []string{"*", `laal`},
					Devnodes: []string{`^hd[a-z]`},
					Devices: []MpathSection{{
						Name:   "device",
						Indent: 2,
						Attr:   map[string][]string{"vendor": {"IBM"}, "product": {"3S42"}},
					}, {
						Name:   "device",
						Indent: 2,
						Attr:   map[string][]string{"vendor": {"HP"}, "product": {"*"}},
					}},
				},
				Defaults: MpathSection{
					Name:   "default",
					Indent: 1,
					Attr:   map[string][]string{"user_friendly_names": {"yes"}, "path_grouping_policy": {"multibus"}},
				},
				Devices: []MpathSection{
					{
						Name:   "device",
						Indent: 2,
						Attr:   map[string][]string{"vendor": {"IBM"}, "product": {"3S42"}},
					}, {
						Name:   "device",
						Indent: 2,
						Attr:   map[string][]string{"vendor": {"HP"}, "product": {"*"}},
					},
				},
				Multipaths: []MpathSection{{
					Name:   "multipath",
					Indent: 2,
					Attr:   map[string][]string{"wwid": {"3600508b4000156d70001200000b0000"}},
				},
					{
						Name:   "multipath",
						Indent: 2,
						Attr:   map[string][]string{"wwid": {"1DEC_____321816758474"}, "alias": {"red"}, "rr_weight": {"priorities"}},
					},
				},
				Overrides: MpathSection{
					Name:   "overrides",
					Indent: 1,
					Attr:   map[string][]string{"user_friendly_names": {"yes"}, "path_grouping_policy": {"multibus"}},
				},
			},
		},
		"with a full multipath file and a different order": {
			filePath: "./testdata/linuxMpath_conf_golden2",
			expectedData: MpathConf{
				BlackList: MpathBlackList{
					Name:     "blacklist",
					Wwids:    []string{"*", `laal`},
					Devnodes: []string{`^hd[a-z]`},
					Devices: []MpathSection{{
						Name:   "device",
						Indent: 2,
						Attr:   map[string][]string{"vendor": {"IBM"}, "product": {"3S42"}},
					}, {
						Name:   "device",
						Indent: 2,
						Attr:   map[string][]string{"vendor": {"HP"}, "product": {"*"}},
					}},
				},
				BlackListExceptions: MpathBlackList{
					Name:     "blacklist_exceptions",
					Wwids:    []string{"*", `laal`},
					Devnodes: []string{`^hd[a-z]`},
					Devices: []MpathSection{{
						Name:   "device",
						Indent: 2,
						Attr:   map[string][]string{"vendor": {"IBM"}, "product": {"3S42"}},
					}, {
						Name:   "device",
						Indent: 2,
						Attr:   map[string][]string{"vendor": {"HP"}, "product": {"*"}},
					}},
				},
				Defaults: MpathSection{
					Name:   "default",
					Indent: 1,
					Attr:   map[string][]string{"user_friendly_names": {"yes"}, "path_grouping_policy": {"multibus"}},
				},
				Devices: []MpathSection{
					{
						Name:   "device",
						Indent: 2,
						Attr:   map[string][]string{"vendor": {"IBM"}, "product": {"3S42"}},
					}, {
						Name:   "device",
						Indent: 2,
						Attr:   map[string][]string{"vendor": {"HP"}, "product": {"*"}},
					},
				},
				Multipaths: []MpathSection{{
					Name:   "multipath",
					Indent: 2,
					Attr:   map[string][]string{"wwid": {"3600508b4000156d70001200000b0000"}},
				},
					{
						Name:   "multipath",
						Indent: 2,
						Attr:   map[string][]string{"wwid": {"1DEC_____321816758474"}, "alias": {"red"}, "rr_weight": {"priorities"}},
					},
				},
				Overrides: MpathSection{
					Name:   "overrides",
					Indent: 1,
					Attr:   map[string][]string{"user_friendly_names": {"yes"}, "path_grouping_policy": {"multibus"}},
				},
			},
		},
	}

	obj := CompMpaths{Obj: &Obj{rules: make([]interface{}, 0), verbose: true}}
	for name, c := range testCases {
		t.Run(name, func(t *testing.T) {
			osReadFile = func(name string) ([]byte, error) {
				return os.ReadFile(c.filePath)
			}
			mPathData, err := obj.loadMpathData()
			require.NoError(t, err)
			require.Equal(t, "", cmp.Diff(c.expectedData, mPathData))
		})
	}
}

func TestGetConfValuesMpath(t *testing.T) {
	fullConf := MpathConf{
		BlackList: MpathBlackList{
			Name:     "blacklist",
			Wwids:    []string{"*2", `laal2`},
			Devnodes: []string{`^hd[a-z]2`},
			Devices: []MpathSection{{
				Name:   "device",
				Indent: 2,
				Attr:   map[string][]string{"vendor": {"IBM"}, "product": {"3S42"}, "color": {"black2"}},
			}, {
				Name:   "device",
				Indent: 2,
				Attr:   map[string][]string{"vendor": {"HP"}, "product": {"*"}},
			}},
		},
		BlackListExceptions: MpathBlackList{
			Name:     "blacklist_exceptions",
			Wwids:    []string{"*", `laal`},
			Devnodes: []string{`^hd[a-z]`},
			Devices: []MpathSection{{
				Name:   "device",
				Indent: 2,
				Attr:   map[string][]string{"vendor": {"IBM"}, "product": {"3S42"}, "color": {"black"}},
			}, {
				Name:   "device",
				Indent: 2,
				Attr:   map[string][]string{"vendor": {"HP"}, "product": {"*"}},
			}},
		},
		Defaults: MpathSection{
			Name:   "default",
			Indent: 1,
			Attr:   map[string][]string{"user_friendly_names": {"yes"}, "path_grouping_policy": {"multibus"}},
		},
		Devices: []MpathSection{
			{
				Name:   "device",
				Indent: 2,
				Attr:   map[string][]string{"vendor": {"IBM"}, "product": {"3S42"}, "color": {"blue"}},
			}, {
				Name:   "device",
				Indent: 2,
				Attr:   map[string][]string{"vendor": {"HP"}, "product": {"*"}},
			},
		},
		Multipaths: []MpathSection{{
			Name:   "multipath",
			Indent: 2,
			Attr:   map[string][]string{"wwid": {"3600508b4000156d70001200000b0000"}},
		},
			{
				Name:   "multipath",
				Indent: 2,
				Attr:   map[string][]string{"wwid": {"1DEC_____321816758474"}, "alias": {"red"}, "rr_weight": {"priorities"}},
			},
		},
		Overrides: MpathSection{
			Name:   "overrides",
			Indent: 1,
			Attr:   map[string][]string{"user_friendly_names": {"yes"}, "path_grouping_policy": {"multibus"}},
		},
	}

	testCases := map[string]struct {
		key            string
		conf           MpathConf
		expectError    bool
		expectedValues []string
	}{
		"with a false first section in key": {
			key:            "fake.user_friendly_names",
			conf:           fullConf,
			expectError:    true,
			expectedValues: nil,
		},

		"with overrides but no section after": {
			key:            "overrides",
			conf:           fullConf,
			expectError:    true,
			expectedValues: nil,
		},

		"with mutipaths but no .multipath after": {
			key:            "multipaths.zozo.att",
			conf:           fullConf,
			expectError:    true,
			expectedValues: nil,
		},

		"with mutipaths but no section after": {
			key:            "multipaths",
			conf:           fullConf,
			expectError:    true,
			expectedValues: nil,
		},

		"with mutipaths but no attribute at the end": {
			key:            "multipaths.multipath",
			conf:           fullConf,
			expectError:    true,
			expectedValues: nil,
		},

		"with devices but no .device after": {
			key:            "devices.zozo.att",
			conf:           fullConf,
			expectError:    true,
			expectedValues: nil,
		},

		"with devices but no section after": {
			key:            "devices",
			conf:           fullConf,
			expectError:    true,
			expectedValues: nil,
		},

		"with devices but no attribute at the end": {
			key:            "devices.device",
			conf:           fullConf,
			expectError:    true,
			expectedValues: nil,
		},

		"with default but no section after": {
			key:            "default",
			conf:           fullConf,
			expectError:    true,
			expectedValues: nil,
		},

		"with blacklist but no section after": {
			key:            "blacklist",
			conf:           fullConf,
			expectError:    true,
			expectedValues: nil,
		},

		"with blacklist but false section after": {
			key:            "blacklist.falseSection",
			conf:           fullConf,
			expectError:    true,
			expectedValues: nil,
		},

		"with blacklist with device and no attribute": {
			key:            "blacklist.device",
			conf:           fullConf,
			expectError:    true,
			expectedValues: nil,
		},

		"with blacklist_exceptions but no section after": {
			key:            "blacklist_exceptions",
			conf:           fullConf,
			expectError:    true,
			expectedValues: nil,
		},

		"with blacklist_exceptions but false section after": {
			key:            "blacklist_exceptions.falseSection",
			conf:           fullConf,
			expectError:    true,
			expectedValues: nil,
		},

		"with blacklist_exceptions with device and no attribute": {
			key:            "blacklist_exceptions.device",
			conf:           fullConf,
			expectError:    true,
			expectedValues: nil,
		},

		"with overrides": {
			key:            "overrides.user_friendly_names",
			conf:           fullConf,
			expectError:    false,
			expectedValues: []string{"yes"},
		},

		"with multipaths": {
			key:            "multipaths.multipath.{1DEC_____321816758474}.alias",
			conf:           fullConf,
			expectError:    false,
			expectedValues: []string{"red"},
		},

		"with devices": {
			key:            `devices.device.{"IBM"}.{3S42}.color`,
			conf:           fullConf,
			expectError:    false,
			expectedValues: []string{"blue"},
		},

		"with default": {
			key:            `default.path_grouping_policy`,
			conf:           fullConf,
			expectError:    false,
			expectedValues: []string{"multibus"},
		},

		"with blacklist_exceptions and wwid": {
			key:            `blacklist_exceptions.wwid`,
			conf:           fullConf,
			expectError:    false,
			expectedValues: []string{"*", "laal"},
		},

		"with blacklist_exceptions and devnode": {
			key:            `blacklist_exceptions.devnode`,
			conf:           fullConf,
			expectError:    false,
			expectedValues: []string{"^hd[a-z]"},
		},

		"with blacklist_exceptions and device": {
			key:            `blacklist_exceptions.device.{IBM}.{3S42}.color`,
			conf:           fullConf,
			expectError:    false,
			expectedValues: []string{"black"},
		},

		"with blacklist and wwid": {
			key:            `blacklist.wwid`,
			conf:           fullConf,
			expectError:    false,
			expectedValues: []string{"*2", "laal2"},
		},

		"with blacklist and devnode": {
			key:            `blacklist.devnode`,
			conf:           fullConf,
			expectError:    false,
			expectedValues: []string{"^hd[a-z]2"},
		},

		"with blacklist and device": {
			key:            `blacklist.device.{IBM}.{3S42}.color`,
			conf:           fullConf,
			expectError:    false,
			expectedValues: []string{"black2"},
		},
	}

	obj := CompMpaths{Obj: &Obj{rules: make([]interface{}, 0), verbose: true}}
	for name, c := range testCases {
		t.Run(name, func(t *testing.T) {
			if c.expectError {
				_, err := obj.getConfValues(c.key, c.conf)
				require.Error(t, err)
			} else {
				values, err := obj.getConfValues(c.key, c.conf)
				require.NoError(t, err)
				require.Equal(t, c.expectedValues, values)
			}
		})
	}
}
