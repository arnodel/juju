// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package tags_test

import (
	"github.com/juju/names"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/environs/tags"
	"github.com/juju/juju/testing"
)

type tagsSuite struct {
	testing.BaseSuite
}

var _ = gc.Suite(&tagsSuite{})

func (*tagsSuite) TestResourceTagsUUID(c *gc.C) {
	testResourceTags(c, testing.ModelTag, nil, map[string]string{
		"juju-model-uuid": testing.ModelTag.Id(),
	})
	testResourceTags(c, names.NewModelTag(""), nil, map[string]string{
		"juju-model-uuid": "",
	})
}

func (*tagsSuite) TestResourceTagsResourceTaggers(c *gc.C) {
	testResourceTags(c, testing.ModelTag, []tags.ResourceTagger{
		resourceTagger(func() (map[string]string, bool) {
			return map[string]string{
				"over":   "ridden",
				"froman": "egg",
			}, true
		}),
		resourceTagger(func() (map[string]string, bool) {
			return nil, false
		}),
		resourceTagger(func() (map[string]string, bool) {
			return nil, true
		}),
		resourceTagger(func() (map[string]string, bool) {
			return map[string]string{"omit": "me"}, false
		}),
		resourceTagger(func() (map[string]string, bool) {
			return map[string]string{
				"over":  "easy",
				"extra": "play",
			}, true
		}),
	}, map[string]string{
		"juju-model-uuid": testing.ModelTag.Id(),
		"froman":          "egg",
		"over":            "easy",
		"extra":           "play",
	})
}

func testResourceTags(c *gc.C, tag names.ModelTag, taggers []tags.ResourceTagger, expectTags map[string]string) {
	tags := tags.ResourceTags(tag, taggers...)
	c.Assert(tags, jc.DeepEquals, expectTags)
}

type resourceTagger func() (map[string]string, bool)

func (r resourceTagger) ResourceTags() (map[string]string, bool) {
	return r()
}
